package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/subcommands"
	htmlTemplate "html/template"
	"io"
	"os"
	"strings"
	textTemplate "text/template"
	"time"
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
	index.load()
	return exportCli(os.Stdout, cmd.templateName, &index)
}

// exportCli runs the export command on the command line. In order to make testing easier, it takes a Writer and an
// indexStore. The Writer is important so that test code can provide a buffer instead of os.Stdout; the indexStore is
// important so that test code can ensure no other test running in parallel can interfere with the list of known pages
// (by adding or deleting pages).
func exportCli(w io.Writer, templateName string, idx *indexStore) subcommands.ExitStatus {
	loadLanguages()
	feed := new(Feed)
	items := []Item{}
	// feed.Name remains unset
	feed.Date = time.Now().Format(time.RFC3339)
	for name, title := range idx.titles {
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
		it.Body = p.Body
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
