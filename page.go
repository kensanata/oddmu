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

// Page is a struct containing information about a single page. Title is the title extracted from the page content using
// titleRegexp. Name is the path without extension (so a path of "foo.md" results in the Name "foo"). Body is the
// Markdown content of the page and Html is the rendered HTML for that Markdown.
type Page struct {
	Title    string
	Name     string
	Body     []byte
	Html     template.HTML
	Hashtags []string
}

// Link is a struct containing a title and a name. Name is the path without extension (so a path of "foo.md" results in
// the Name "foo").
type Link struct {
	Title string
	Url   string
}

// blogRe is a regular expression that matches blog pages. If the filename of a blog page starts with an ISO date
// (YYYY-MM-DD), then it's a blog page.
var blogRe = regexp.MustCompile(`^\d\d\d\d-\d\d-\d\d`)

// santizeStrict uses bluemonday to sanitize the HTML away. No elements are allowed except for the b tag because this is
// used for snippets.
func sanitizeStrict(s string) template.HTML {
	policy := bluemonday.StrictPolicy()
	policy.AllowElements("b")
	return template.HTML(policy.Sanitize(s))
}

// unsafeBytes does not use bluemonday to sanitize the HTML used for pages. This is where you make changes if you want
// to be more lenient. If you look at the git repository, there are older versions containing the function sanitizeBytes
// which would do elaborate checking.
func unsafeBytes(bytes []byte) template.HTML {
	return template.HTML(bytes)
}

// nameEscape returns the page name safe for use in URLs. That is, percent escaping is used except for the slashes.
func nameEscape(s string) string {
	parts := strings.Split(s, "/")
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	return strings.Join(parts, "/")
}

// save saves a Page. The path is based on the Page.Name and gets the ".md" extension. Page.Body is saved, without any
// carriage return characters ("\r"). Page.Title and Page.Html are not saved. There is no caching. Before removing or
// writing a file, the old copy is renamed to a backup, appending "~". Errors are not logged but returned.
func (p *Page) save() error {
	fp := filepath.FromSlash(p.Name) + ".md"
	watches.ignore(fp)
	s := bytes.ReplaceAll(p.Body, []byte{'\r'}, []byte{})
	if len(s) == 0 {
		log.Println("Delete", p.Name)
		index.remove(p)
		return os.Rename(fp, fp+"~")
	}
	p.Body = s
	index.update(p)
	d := filepath.Dir(fp)
	if d != "." {
		err := os.MkdirAll(d, 0755)
		if err != nil {
			return err
		}
	}
	err := backup(fp)
	if err != nil {
		return err
	}
	return os.WriteFile(fp, s, 0644)
}

// backup a file by renaming it unless the existing backup is less than an hour old. A backup gets a tilde appended to
// it ("~"). This is true even if the file refers to a binary file like "image.png" and most applications don't know
// what to do with a file called "image.png~". This expects a filepath. The backup file gets its modification time set
// to now so that subsequent edits don't immediately overwrite it again.
func backup(fp string) error {
	_, err := os.Stat(fp)
	if err != nil {
		return nil
	}
	bp := fp + "~"
	fi, err := os.Stat(bp)
	if err != nil || time.Since(fi.ModTime()).Minutes() >= 60 {
		err = os.Rename(fp, bp)
		if err != nil {
			return err
		}
		ts := time.Now()
		return os.Chtimes(bp, ts, ts)
	}
	return nil
}

// loadPage loads a Page given a name. The path loaded is that Page.Name with the ".md" extension. The Page.Title is set
// to the Page.Name (and possibly changed, later). The Page.Body is set to the file content. The Page.Html remains
// undefined (there is no caching).
func loadPage(name string) (*Page, error) {
	name = strings.TrimPrefix(name, "./") // result of a path.TreeWalk starting with "."
	body, err := os.ReadFile(filepath.FromSlash(name) + ".md")
	if err != nil {
		return nil, err
	}
	return &Page{Title: name, Name: name, Body: body}, nil
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

// summarize sets Page.Html to an extract.
func (p *Page) summarize(q string) {
	t := p.plainText()
	p.Html = sanitizeStrict(snippets(q, t))
}

// IsBlog returns true if the page name starts with an ISO date
func (p *Page) IsBlog() bool {
	name := path.Base(p.Name)
	return blogRe.MatchString(name)
}

const upperhex = "0123456789ABCDEF"

// Path returns the page name with semicolon, comma and questionmark escaped because html/template doesn't escape those.
// This is suitable for use in HTML templates.
func (p *Page) Path() string {
	s := p.Name
	n := strings.Count(s, ";") + strings.Count(s, ",") + strings.Count(s, "?")
	if n == 0 {
		return p.Name
	}
	t := make([]byte, len(s) + 2*n)
	j := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case ';', ',', '?':
			t[j] = '%'
			t[j+1] = upperhex[s[i]>>4]
			t[j+2] = upperhex[s[i]&15]
			j += 3
		default:
			t[j] = s[i]
			j++
		}
	}
	return string(t);
}

// Dir returns the directory part of the page name. It's either the empty string if the page is in the Oddmu working
// directory, or it ends in a slash. This is used to create the upload link in "view.html", for example.
func (p *Page) Dir() string {
	d := path.Dir(p.Name)
	if d == "." {
		return ""
	}
	return d + "/"
}

// Base returns the basename of the page name: no directory. This is used to create the upload link in "view.html", for
// example.
func (p *Page) Base() string {
	n := path.Base(p.Name)
	if n == "." {
		return ""
	}
	return n
}

// Today returns the date, as a string, for use in templates.
func (p *Page) Today() string {
	return time.Now().Format(time.DateOnly)
}

// Parents returns a Link array to parent pages, up the directory structure.
func (p *Page) Parents() []*Link {
	links := make([]*Link, 0)
	index.RLock()
	defer index.RUnlock()
	// foo/bar/baz ⇒ index, foo/index
	elems := strings.Split(p.Name, "/")
	if len(elems) == 1 {
		return links
	}
	s := ""
	for i := 0; i < len(elems)-1; i++ {
		name := s + "index"
		title, ok := index.titles[name]
		if !ok {
			title = "…"
		}
		link := &Link{Title: title, Url: strings.Repeat("../", len(elems)-i-1) + "index"}
		links = append(links, link)
		s += elems[i] + "/"
	}
	return links
}
