package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/subcommands"
	"io"
	"os"
	"strings"
	"time"
	textTemplate "text/template"
	htmlTemplate "html/template"
)

type exportCmd struct {
	templateName string
}

func (cmd *exportCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&cmd.templateName, "template", "feed.html", "template filename")
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

  Options:

  -template "filename" specifies the template to use (default: feed.html)
`
}

func (cmd *exportCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	return exportCli(os.Stdout, cmd.templateName)
}

// exportCli runs the export command on the command line. It is used
// here with an io.Writer for easy testing.
func exportCli(w io.Writer, templateName string) subcommands.ExitStatus {
	index.load()
	feed := new(Feed)
	items := []Item{}
	// feed.Name remains unset
	feed.Date = time.Now().Format(time.RFC3339)
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
		it := Item{Date: fi.ModTime().Format(time.RFC3339)}
		it.Title = p.Title
		it.Name = p.Name
		it.Html = htmlTemplate.HTML(htmlTemplate.HTMLEscaper(p.Html))
		it.Hashtags = p.Hashtags
		items = append(items, it)
	}
	feed.Items = items
	// No effort is made to work with the templates var.
	if strings.HasSuffix(templateName, ".html") ||
		strings.HasSuffix(templateName, ".xml") ||
		strings.HasSuffix(templateName, ".rss") {
		w.Write([]byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n"))
		t, err := htmlTemplate.ParseFiles(templateName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Parsing %s: %s\n", templateName, err)
			return subcommands.ExitFailure
		}
		err = t.Execute(w, feed)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Writing feed: %s\n", err)
			return subcommands.ExitFailure
		}
	} else {
		t, err := textTemplate.ParseFiles(templateName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Parsing %s: %s\n", templateName, err)
			return subcommands.ExitFailure
		}
		err = t.Execute(w, feed)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Writing feed: %s\n", err)
			return subcommands.ExitFailure
		}
	}
	return subcommands.ExitSuccess
}
