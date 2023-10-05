package main

import (
	"bytes"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"html/template"
	"os"
	"path"
	"time"
)

type Item struct {
	Page
	Date string
}

type Feed struct {
	Item
	Items []Item
}

func feed(p *Page, ti time.Time) *Feed {
	feed := new(Feed)
	feed.Name = p.Name
	feed.Title = p.Title
	feed.Date = ti.Format(time.RFC1123Z)
	parser, _ := wikiParser()
	doc := markdown.Parse(p.Body, parser)
	items := make([]Item, 0)
	inListItem := false
	ast.WalkFunc(doc, func(node ast.Node, entering bool) ast.WalkStatus {
		// set the flag if we're in a list item
		listItem, ok := node.(*ast.ListItem)
		if ok && listItem.BulletChar == '*' {
			inListItem = entering
			return ast.GoToNext
		}
		// if we're not in a list item, continue
		if !inListItem || !entering {
			return ast.GoToNext
		}
		// if we're in a link and it's local
		link, ok := node.(*ast.Link)
		if !ok || bytes.Contains(link.Destination, []byte("//")) {
			return ast.GoToNext
		}
		name := path.Join(path.Dir(p.Name), string(link.Destination))
		fi, err := os.Stat(name + ".md")
		if err != nil {
			return ast.GoToNext
		}
		p2, err := loadPage(name)
		if err != nil {
			return ast.GoToNext
		}
		p2.handleTitle(false)
		p2.renderHtml()
		it := Item{Date: fi.ModTime().Format(time.RFC1123Z)}
		it.Title = p2.Title
		it.Name = p2.Name
		it.Html = template.HTML(template.HTMLEscaper(p2.Html))
		it.Hashtags = p2.Hashtags
		items = append(items, it)
		if len(items) >= 10 {
			return ast.Terminate
		}
		return ast.GoToNext
	})
	feed.Items = items
	return feed
}
