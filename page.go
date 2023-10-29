package main

import (
	"bytes"
	"github.com/microcosm-cc/bluemonday"
	"html/template"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"
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

var blogRe = regexp.MustCompile(`^\d\d\d\d-\d\d-\d\d`)

// santizeStrict uses bluemonday to sanitize the HTML away. No elements are allowed except for the b tag because this is
// used for snippets.
func sanitizeStrict(s string) template.HTML {
	policy := bluemonday.StrictPolicy()
	policy.AllowElements("b")
	return template.HTML(policy.Sanitize(s))
}

// santizeBytes uses bluemonday to sanitize the HTML used for pages. This is where you make changes if you want to be
// more lenient.
func sanitizeBytes(bytes []byte) template.HTML {
	policy := bluemonday.UGCPolicy()
	policy.AllowURLSchemes("gemini", "gopher")
	policy.AllowAttrs("class", "title").OnElements("a") // for hashtags and accounts
	policy.AllowAttrs("loading").OnElements("img") // for lazy loading
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

// handleTitle extracts the title from a Page and sets Page.Title, if any. If replace is true, the page title is also
// removed from Page.Body. Make sure not to save this! This is only for rendering. In a template, the title is a
// separate attribute and is not repeated in the HTML.
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

// isBlog returns true if the page name starts with an ISO date
func (p *Page) isBlog() bool {
	name := path.Base(p.Name)
	return blogRe.MatchString(name)
}

// Dir returns the directory the page is in. It's either the empty string if the page is in the Oddmu working directory,
// or it ends in a slash. This is used to create the upload link in "view.html", for example.
func (p *Page) Dir() string {
	d := filepath.Dir(p.Name)
	if d == "." {
		return ""
	}
	return d + "/"
}

// Today returns the date, as a string, for use in templates.
func (p *Page) Today() string {
	return time.Now().Format(time.DateOnly)
}
