package main

import(
	"bytes"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
	"net/url"
)

// wikiLink returns an inline parser function. This indirection is
// required because we want to call the previous definition in case
// this is not a wikiLink.
func wikiLink(p *parser.Parser,	fn func(p *parser.Parser, data []byte, offset int) (int, ast.Node)) func(p *parser.Parser, data []byte, offset int) (int, ast.Node) {
	return func (p *parser.Parser, original []byte, offset int) (int, ast.Node) {
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
		text := data[2:i+1]
		link := &ast.Link{
			Destination: []byte(url.PathEscape(string(text))),
		}
		ast.AppendChild(link, &ast.Text{Leaf: ast.Leaf{Literal: text}})
		return i+3, link
	}
}


func hashtag(p *parser.Parser, data []byte, offset int) (int, ast.Node) {
	data = data[offset:]
	i := 0
	n := len(data)
	for i < n && !parser.IsSpace(data[i]) {
		i++
	}
	if i == 0 {
		return 0, nil
	}
	link := &ast.Link{
		Destination: append([]byte("/search?q=%23"), data[1:i]...),
		Title:       data[0:i],
	}
	text := bytes.ReplaceAll(data[0:i], []byte("_"), []byte(" "))
	ast.AppendChild(link, &ast.Text{Leaf: ast.Leaf{Literal: text}})
	return i, link
}

// account links a social media account @account@domain to
// https://domain/user/account.
func account(p *parser.Parser, data []byte, offset int) (int, ast.Node) {
	data = data[offset:]
	i := 1 // skip @ of username
	n := len(data)
	d := 0
	for i < n && (
		data[i] >= 'a' && data[i] <= 'z' ||
		data[i] >= 'A' && data[i] <= 'Z' ||
		data[i] >= '0' && data[i] <= '9' ||
		data[i] == '@' ||
		data[i] == '.' ||
		data[i] == '-') {
		if data[i] == '@' {
			if d != 0 {
				return 0, nil
			} else {
				d = i+1 // skip @ of domain
			}
		}
		i++
	}
	for i > 1 && (
		data[i-1] == '.' ||
		data[i-1] == '-') {
		i--
	}
	if i == 0 {
		return 0, nil
	}
	user := data[0:d-1] // includes @
	domain := data[d:i]
	dest := make([]byte, 0, 14 + len(domain) + len(user))
	dest = append(dest, []byte("https://")...) // len 8
	dest = append(dest, domain...)
	dest = append(dest, []byte("/users/")...) // len 7
	dest = append(dest, user[1:]...) // skip @
	link := &ast.Link{
		Destination: dest,
		Title:       data[0:i],
	}
	ast.AppendChild(link, &ast.Text{Leaf: ast.Leaf{Literal: user}})
	return i, link
}

// renderHtml renders the Page.Body to HTML and sets Page.Html.
func (p *Page) renderHtml() {
	parser := parser.New()
	prev := parser.RegisterInline('[', nil)
	parser.RegisterInline('[', wikiLink(parser, prev))
	parser.RegisterInline('#', hashtag)
	parser.RegisterInline('@', account)
	maybeUnsafeHTML := markdown.ToHTML(p.Body, parser, nil)
	p.Name = nameEscape(p.Name)
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
