package main

import (
	"bytes"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"net/url"
	"path"
	"path/filepath"
)

// wikiLink returns an inline parser function. This indirection is
// required because we want to call the previous definition in case
// this is not a wikiLink.
func wikiLink(fn func(p *parser.Parser, data []byte, offset int) (int, ast.Node)) func(p *parser.Parser, data []byte, offset int) (int, ast.Node) {
	return func(p *parser.Parser, original []byte, offset int) (int, ast.Node) {
		data := original[offset:]
		n := len(data)
		// minimum: [[X]]
		if n < 5 || data[1] != '[' {
			return fn(p, original, offset)
		}
		i := 2
		for i+1 < n && data[i] != ']' && data[i+1] != ']' {
			i++
		}
		text := data[2 : i+1]
		link := &ast.Link{
			Destination: []byte(url.PathEscape(string(text))),
		}
		ast.AppendChild(link, &ast.Text{Leaf: ast.Leaf{Literal: text}})
		return i + 3, link
	}
}

// hashtag returns an inline parser function. This indirection is
// required because we want to receive an array of hashtags found.
func hashtag() (func(p *parser.Parser, data []byte, offset int) (int, ast.Node), *[]string) {
	hashtags := make([]string, 0)
	return func(p *parser.Parser, data []byte, offset int) (int, ast.Node) {
		data = data[offset:]
		i := 0
		n := len(data)
		for i < n && !parser.IsSpace(data[i]) {
			i++
		}
		if i <= 1 {
			return 0, nil
		}
		hashtags = append(hashtags, string(data[1:i]))
		link := &ast.Link{
			AdditionalAttributes: []string{`class="tag"`},
			Destination:          append([]byte("/search/?q=%23"), data[1:i]...),
		}
		text := bytes.ReplaceAll(data[0:i], []byte("_"), []byte(" "))
		ast.AppendChild(link, &ast.Text{Leaf: ast.Leaf{Literal: text}})
		return i, link
	}, &hashtags
}

// wikiParser returns a parser with the Oddmu specific changes. Specifically: [[wiki links]], #hash_tags,
// @webfinger@accounts. It also uses the CommonExtensions and Block Attributes, and no MathJax ($).
func wikiParser() (*parser.Parser, *[]string) {
	extensions := (parser.CommonExtensions | parser.AutoHeadingIDs | parser.Attributes) & ^parser.MathJax
	p := parser.NewWithExtensions(extensions)
	prev := p.RegisterInline('[', nil)
	p.RegisterInline('[', wikiLink(prev))
	fn, hashtags := hashtag()
	p.RegisterInline('#', fn)
	if useWebfinger {
		p.RegisterInline('@', accountLink)
		// handle escape with \@
		var escape func(p *parser.Parser, data []byte, offset int) (int, ast.Node);
		newEscape := func(p *parser.Parser, data []byte, offset int) (int, ast.Node) {
			i := offset + 1
			if len(data) > i && data[i] == '@' {
				return 2, &ast.Text{Leaf: ast.Leaf{Literal: data[i:i+1]}}
			}
			return escape(p, data, offset)
		}
		escape = p.RegisterInline('\\', newEscape)
	}
	return p, hashtags
}

// wikiRenderer is a Renderer for Markdown that adds lazy loading of images and disables fractions support. Remember
// that there is no HTML sanitization.
func wikiRenderer() *html.Renderer {
	// sync with staticPage
	htmlFlags := html.CommonFlags & ^html.SmartypantsFractions | html.LazyLoadImages
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)
	return renderer
}

// renderHtml renders the Page.Body to HTML and sets Page.Html, Page.Hashtags, and escapes Page.Name.
func (p *Page) renderHtml() {
	parser, hashtags := wikiParser()
	renderer := wikiRenderer()
	maybeUnsafeHTML := markdown.ToHTML(p.Body, parser, renderer)
	p.Name = nameEscape(p.Name)
	p.Html = unsafeBytes(maybeUnsafeHTML)
	p.Hashtags = *hashtags
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

// images returns an array of ImageData.
func (p *Page) images() []ImageData {
	dir := path.Dir(filepath.ToSlash(p.Name))
	images := make([]ImageData, 0)
	parser := parser.New()
	doc := markdown.Parse(p.Body, parser)
	ast.WalkFunc(doc, func(node ast.Node, entering bool) ast.WalkStatus {
		if entering {
			switch v := node.(type) {
			case *ast.Image:
				// not an absolute URL, not a full URL, not a mailto: URI
				text := toString(v)
				if len(text) > 0 {
					name := path.Join(dir, string(v.Destination))
					image := ImageData{Title: text, Name: name}
					images = append(images, image)
				}
				return ast.SkipChildren
			}
		}
		return ast.GoToNext
	})
	return images
}

// toString for a node returns the text nodes' literals, concatenated. There is no whitespace added so the expectation
// is that there is only one child node. Otherwise, there may be a space missing between the literals, depending on the
// exact child nodes they belong to.
func toString(node ast.Node) string {
	b := new(bytes.Buffer)
	ast.WalkFunc(node, func(node ast.Node, entering bool) ast.WalkStatus {
		if entering {
			switch v := node.(type) {
			case *ast.Text:
				b.Write(v.Literal)
			}
		}
		return ast.GoToNext
	})
	return b.String()
}
