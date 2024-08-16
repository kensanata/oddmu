package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/google/subcommands"
	"io"
	"net/url"
	"os"
	"path"
	"strings"
)

type missingCmd struct {
}

func (*missingCmd) Name() string     { return "missing" }
func (*missingCmd) Synopsis() string { return "list missing pages" }
func (*missingCmd) Usage() string {
	return `missing:
  Listing pages with links to missing pages. This command does not
  understand links to directories being redirected to index pages.
  A link such as [up](..) is reported as a link to a missing page.
  Rewrite it as [up](../index) for it to work as intended.
`
}

func (cmd *missingCmd) SetFlags(f *flag.FlagSet) {
}

func (cmd *missingCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	return missingCli(os.Stdout, &index)
}

// missingCli implements the finding of links to missing pages. In order to make testing easier, it takes a Writer and
// an indexStore. The Writer is important so that test code can provide a buffer instead of os.Stdout; the indexStore is
// important so that test code can ensure no other test running in parallel can interfere with the list of known pages
// (by adding or deleting pages).
func missingCli(w io.Writer, idx *indexStore) subcommands.ExitStatus {
	found := false
	for name := range idx.titles {
		p, err := loadPage(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Loading %s: %s\n", name, err)
			return subcommands.ExitFailure
		}
		for _, link := range p.links() {
			u, err := url.Parse(link)
			if err != nil {
				fmt.Fprintln(os.Stderr, p.Name, err)
				return subcommands.ExitFailure
			}
			if u.Scheme == "" && u.Path != "" && !strings.HasPrefix(u.Path, "/") {
				// feeds can work if the matching page works
				u.Path = strings.TrimSuffix(u.Path, ".rss")
				// links to the source file can work
				u.Path = strings.TrimSuffix(u.Path, ".md")
				// pages containing a colon need the ./ prefix
				u.Path = strings.TrimPrefix(u.Path, "./")
				// check whether the destination is a known page
				destination, err := url.PathUnescape(u.Path)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Cannot decode %s: %s\n", link, err)
					return subcommands.ExitFailure
				}
				_, ok := idx.titles[destination]
				// links to directories can work
				if !ok {
					_, ok = idx.titles[path.Join(destination, "index")]
				}
				if !ok {
					if !found {
						fmt.Fprintln(w, "Page\tMissing")
						found = true
					}
					fmt.Fprintf(w, "%s\t%s\n", p.Name, link)
				}
			}
		}
	}
	if !found {
		fmt.Fprintln(w, "No missing pages found.")
	}
	return subcommands.ExitSuccess
}

// links parses the page content and returns an array of link destinations.
func (p *Page) links() []string {
	var links []string
	parser, _ := wikiParser()
	doc := markdown.Parse(p.Body, parser)
	ast.WalkFunc(doc, func(node ast.Node, entering bool) ast.WalkStatus {
		if entering {
			switch v := node.(type) {
			case *ast.Link:
				link := string(v.Destination)
				url, err := url.Parse(link)
				if err != nil {
					// no error reporting
					return ast.GoToNext
				}
				if url.IsAbs() {
					links = append(links, link)
				} else {
					dir := p.Dir()
					links = append(links, path.Join(dir, link))
				}
			}
		}
		return ast.GoToNext
	})
	return links
}
