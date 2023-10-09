package main

import (
	"bytes"
	"github.com/microcosm-cc/bluemonday"
	"html/template"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Page is a struct containing information about a single page. Title
// is the title extracted from the page content using titleRegexp.
// Name is the filename without extension (so a filename of "foo.md"
// results in the Name "foo"). Body is the Markdown content of the
// page and Html is the rendered HTML for that Markdown. Score is a
// number indicating how well the page matched for a search query.
type Page struct {
	Title    string
	Name     string
	Language string
	Body     []byte
	Html     template.HTML
	Score    int
	Hashtags []string
}

// santize uses bluemonday to sanitize the HTML. An exceptions is made
// for the b tag because this is used for snippets.
func sanitizeStrict(s string) template.HTML {
	policy := bluemonday.StrictPolicy()
	policy.AllowElements("b")
	return template.HTML(policy.Sanitize(s))
}

// santizeBytes uses bluemonday to sanitize the HTML.
func sanitizeBytes(bytes []byte) template.HTML {
	policy := bluemonday.UGCPolicy()
	policy.AllowAttrs("class", "title").OnElements("a") // for hashtags and accounts
	return template.HTML(policy.SanitizeBytes(bytes))
}

// nameEscape returns the page name safe for use in URLs. That is,
// percent escaping is used except for the slashes.
func nameEscape(s string) string {
	parts := strings.Split(s, "/")
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	return strings.Join(parts, "/")
}

// save saves a Page. The filename is based on the Page.Name and gets
// the ".md" extension. Page.Body is saved, without any carriage
// return characters ("\r"). Page.Title and Page.Html are not saved.
// There is no caching. Before removing or writing a file, the old
// copy is renamed to a backup, appending "~". There is no error
// checking for this.
func (p *Page) save() error {
	filename := p.Name + ".md"
	s := bytes.ReplaceAll(p.Body, []byte{'\r'}, []byte{})
	if len(s) == 0 {
		p.removeFromIndex()
		return os.Rename(filename, filename+"~")
	}
	p.Body = s
	p.updateIndex()
	d := filepath.Dir(filename)
	if d != "." {
		err := os.MkdirAll(d, 0755)
		if err != nil {
			log.Printf("Creating directory %s failed: %s", d, err)
			return err
		}
	}
	_ = os.Rename(filename, filename+"~")
	return os.WriteFile(filename, s, 0644)
}

func (p *Page) notify() error {
	c, err := loadPage("changes")
	p.handleTitle(false)
	if p.Title == "" {
		p.Title = p.Name
	}
	esc := nameEscape(p.Name)
	if err != nil {
		// create a new page
		c = &Page{Name: "changes", Body: []byte("# Changes\n\n* [" + p.Title + "](" + esc + ")\n")}
	} else {
		// remove the old match, if one exists
		re := regexp.MustCompile(`(?m)^\* \[[^\]]+\]\(` + esc + `\)\n`)
		loc := re.FindIndex(c.Body)
		if loc != nil {
			r := c.Body[:loc[0]]
			if loc[1] < len(c.Body) {
				r = append(r, c.Body[loc[1]+1:]...)
			}
			c.Body = r
		}
		// locate the beginning of the list to insert the line
		re = regexp.MustCompile(`(?m)^\* \[[^\]]+\]\([^\)]+\)\n`)
		loc = re.FindIndex(c.Body)
		if loc == nil {
			// if no list was found, use the end of the page
			loc = []int{len(c.Body)}
		}
		r := []byte("")
		r = append(r, c.Body[:loc[0]]...)
		if len(r) > 0 && r[len(r)-1] != '\n' {
			r = append(r, '\n')
		}
		if len(r) > 1 && r[len(r)-2] != '\n' {
			r = append(r, '\n')
		}
		r = append(r, []byte("* ["+p.Title+"]("+esc+")\n")...)
		r = append(r, c.Body[loc[0]:]...)
		c.Body = r
	}
	return c.save()
}

// loadPage loads a Page given a name. The filename loaded is that
// Page.Name with the ".md" extension. The Page.Title is set to the
// Page.Name (and possibly changed, later). The Page.Body is set to
// the file content. The Page.Html remains undefined (there is no
// caching).
func loadPage(name string) (*Page, error) {
	filename := name + ".md"
	body, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: name, Name: name, Body: body, Language: ""}, nil
}

// handleTitle extracts the title from a Page and sets Page.Title, if
// any. If replace is true, the page title is also removed from
// Page.Body. Make sure not to save this! This is only for rendering.
func (p *Page) handleTitle(replace bool) {
	s := string(p.Body)
	m := titleRegexp.FindStringSubmatch(s)
	if m != nil {
		p.Title = m[1]
		if replace {
			p.Body = []byte(strings.Replace(s, m[0], "", 1))
		}
	}
}

// score sets Page.Title and computes Page.Score.
func (p *Page) score(q string) {
	p.handleTitle(true)
	p.Score = score(q, string(p.Body)) + score(q, p.Title)
}

// summarize sets Page.Html to an extract and sets Page.Language.
func (p *Page) summarize(q string) {
	t := p.plainText()
	p.Name = nameEscape(p.Name)
	p.Html = sanitizeStrict(snippets(q, t))
	p.Language = language(t)
}

// Dir returns the directory the page is in. It's either the empty string if the page is in the Oddmu working directory,
// or it ends in a slash.
func (p *Page) Dir() string {
	d := filepath.Dir(p.Name)
	if d == "." {
		return ""
	}
	return d + "/"
}
