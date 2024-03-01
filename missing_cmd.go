package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/google/subcommands"
	"io"
	"io/fs"
	"net/url"
	"os"
	"path"
	"path/filepath"
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
	return missingCli(os.Stdout)
}

func missingCli(w io.Writer) subcommands.ExitStatus {
	names, err := existingPages()
	if err != nil {
		fmt.Fprintln(w, err)
		return subcommands.ExitFailure
	}
	found := false
	for name, isPage := range names {
		if !isPage {
			continue
		}
		p, err := loadPage(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Loading %s: %s\n", p.Name, err)
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
				// check whether the destinatino is a known page
				destination, err := url.PathUnescape(u.Path)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Cannot decode %s: %s\n", link, err)
					return subcommands.ExitFailure
				}
				_, ok := names[destination]
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

func existingPages() (map[string]bool, error) {
	names := make(map[string]bool)
	err := filepath.Walk(".", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// skip hidden directories and files
		if path != "." && strings.HasPrefix(filepath.Base(path), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			} else {
				return nil
			}
		}
		if strings.HasSuffix(path, ".md") {
			name := filepath.ToSlash(strings.TrimSuffix(path, ".md"))
			names[name] = true
		} else {
			names[path] = false
		}
		return nil
	})
	return names, err
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
				dir := p.Dir()
				links = append(links, path.Join(dir, link))
			}
		}
		return ast.GoToNext
	})
	return links
}
