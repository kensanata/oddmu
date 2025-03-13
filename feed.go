package main

import (
	"bytes"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"html/template"
	"os"
	"path"
	"path/filepath"
	"time"
)

// Item is a Page plus a Date.
type Item struct {

	// Page is the page being used as the feed item.
	Page

	// Date is the last modification date of the file storing the page. As the pages used by Oddmu are plain
	// Markdown files, they don't contain any metadata. Instead, the last modification date of the file is used.
	// This makes it work well with changes made to the files outside of Oddmu.
	Date string
}

// Feed is an Item used for the feed itself, plus an array of items based on the linked pages.
type Feed struct {

	// Item is the page containing the list of links. It's title is used for the feed and it's last modified time is
	// used for the publication date. Thus, if linked pages change but the page with the links doesn't change, the
	// publication date remains unchanged.
	Item

	// Items are based on the pages linked in list items starting with an asterisk ("*"). Links in
	// list items starting with a minus ("-") are ignored!
	Items []Item
}

// feed returns a RSS 2.0 feed for any page. The feed items it contains are the pages linked from in list items starting
// with an asterisk ("*").
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
		name := path.Join(p.Dir(), string(link.Destination))
		fi, err := os.Stat(filepath.FromSlash(name) + ".md")
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
