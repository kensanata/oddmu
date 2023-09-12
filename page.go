package main

import (
	"bytes"
	"fmt"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
	"github.com/microcosm-cc/bluemonday"
	"html/template"
	"os"
	"path/filepath"
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
}

func sanitize(s string) template.HTML {
	return template.HTML(bluemonday.UGCPolicy().Sanitize(s))
}

func sanitizeBytes(bytes []byte) template.HTML {
	return template.HTML(bluemonday.UGCPolicy().SanitizeBytes(bytes))
}

// save saves a Page. The filename is based on the Page.Name and gets
// the ".md" extension. Page.Body is saved, without any carriage
// return characters ("\r"). The file permissions used are readable
// and writeable for the current user, i.e. u+rw or 0600. Page.Title
// and Page.Html are not saved no caching. There is no caching.
func (p *Page) save() error {
	filename := p.Name + ".md"
	s := bytes.ReplaceAll(p.Body, []byte{'\r'}, []byte{})
	if len(s) == 0 {
		return os.Remove(filename)
	}
	p.Body = s
	p.updateIndex()
	d := filepath.Dir(filename)
	if d != "." {
		err := os.MkdirAll(d, 0755)
		if err != nil {
			fmt.Printf("Creating directory %s failed", d)
			return err
		}
	}
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

// renderHtml renders the Page.Body to HTML and sets Page.Html.
func (p *Page) renderHtml() {
	parser := parser.New()
	maybeUnsafeHTML := markdown.ToHTML(p.Body, parser, nil)
	p.Html = sanitizeBytes(maybeUnsafeHTML)
	p.Language = language(p.plainText())
}

// plainText renders the Page.Body to plain text and returns it,
// ignoring all the Markdown and all the newlines. The result is one
// long single line of text.
func (p *Page) plainText() string {
	parser := parser.New()
	doc := markdown.Parse(p.Body, parser)
	text := []byte("")
	ast.WalkFunc(doc, func(node ast.Node, entering bool) ast.WalkStatus {
		if entering && node.AsLeaf() != nil {
			text = append(text, node.AsLeaf().Literal...)
			text = append(text, []byte(" ")...)
		}
		return ast.GoToNext
	})
	// Some Markdown still contains newlines
	for i, c := range text {
		if c == '\n' {
			text[i] = ' '
		}
	}
	// Remove trailing space
	for len(text) > 0 && text[len(text)-1] == ' ' {
		text = text[0 : len(text)-1]
	}
	return string(text)
}

// summarize for query string q sets Page.Html to an extract.
func (p *Page) summarize(q string) {
	p.handleTitle(true)
	p.Score = score(q, string(p.Body)) + score(q, p.Title)
	t := p.plainText()
	p.Html = sanitize(snippets(q, t))
	p.Language = language(t)
}
