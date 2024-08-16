package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/subcommands"
	"html/template"
	"io"
	"os"
	"time"
)

type exportCmd struct {
}

func (cmd *exportCmd) SetFlags(f *flag.FlagSet) {
}

func (*exportCmd) Name() string     { return "export" }
func (*exportCmd) Synopsis() string { return "export the whole site as one big RSS feed" }
func (*exportCmd) Usage() string {
	return `export:
  Export the entire site as one big RSS feed. This may allow you to
  import the whole site into a different content management system.
  The feed contains every page, in HTML format, so the Markdown files
  are part of the feed, but none of the other files.

  The RSS feed is printed to stdout so you probably want to redirect
  it:

    oddmu export > /tmp/export.rss
`
}

func (cmd *exportCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	return exportCli(os.Stdout)
}

// exportCli runs the export command on the command line. It is used
// here with an io.Writer for easy testing.
func exportCli(w io.Writer) subcommands.ExitStatus {
	index.load()
	feed := new(Feed)
	items := []Item{}
	// feed.Name remains unset
	feed.Date = time.Now().Format(time.RFC1123Z)
	for name, title := range index.titles {
		if name == "index" {
			feed.Title = title
		}
		p, err := loadPage(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Loading %s: %s\n", name, err)
			return subcommands.ExitFailure
		}
		p.handleTitle(false)
		p.renderHtml()
		fi, err := os.Stat(name + ".md")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Stat %s: %s\n", name, err)
			return subcommands.ExitFailure
		}
		it := Item{Date: fi.ModTime().Format(time.RFC1123Z)}
		it.Title = p.Title
		it.Name = p.Name
		it.Html = template.HTML(template.HTMLEscaper(p.Html))
		it.Hashtags = p.Hashtags
		items = append(items, it)
	}
	feed.Items = items
	// No effort is made to work with the templates var.
	f := "feed.html"
	t, err := template.ParseFiles(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Parsing %s: %s\n", f, err)
		return subcommands.ExitFailure
	}
	w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>`))
	err = t.Execute(w, feed)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Writing feed: %s\n", err)
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}
