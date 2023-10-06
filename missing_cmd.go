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
	"path/filepath"
	"strings"
)

type missingCmd struct {
}

func (*missingCmd) Name() string     { return "missing" }
func (*missingCmd) Synopsis() string { return "Listing the missing pages." }
func (*missingCmd) Usage() string {
	return `missing:
  Listing pages with links to missing pages.
`
}

func (cmd *missingCmd) SetFlags(f *flag.FlagSet) {
}

func (cmd *missingCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	return missingCli(os.Stdout, f.Args())
}

func missingCli(w io.Writer, args []string) subcommands.ExitStatus {
	names := make(map[string]bool)
	err := filepath.Walk(".", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		filename := path
		if info.IsDir() || strings.HasPrefix(filename, ".") {
			return nil
		}
		if strings.HasSuffix(filename, ".md") {
			name := strings.TrimSuffix(filename, ".md")
			names[name] = true
		} else {
			names[filename] = false
		}
		return nil
	})
	if err != nil {
		fmt.Fprintln(w, err)
		return subcommands.ExitFailure
	}
	fmt.Fprintln(w, "Page\tMissing")
	for name, isPage := range names {
		if !isPage {
			continue
		}
		p, err := loadPage(name)
		if err != nil {
			fmt.Fprintf(w, "Loading %s: %s\n", name, err)
			return subcommands.ExitFailure
		}
		for _, link := range p.links() {
			u, err := url.Parse(link)
			if err != nil {
				fmt.Fprintf(w, "Cannot parse %s: %s", link, err)
				return subcommands.ExitFailure
			}
			if u.Scheme == "" && !strings.HasPrefix(u.Path, "/") {
				// feeds can work if the matching page works
				link = strings.TrimSuffix(link, ".rss")
				destination, err := url.PathUnescape(u.Path)
				if err != nil {
					fmt.Fprintf(w, "Cannot decode %s: %s\n", link, err)
					return subcommands.ExitFailure
				}
				_, ok := names[destination]
				if !ok {
					fmt.Fprintf(w, "%s\t%s\n", name, link)
				}
			}
		}
	}
	return subcommands.ExitSuccess
}

func (p *Page) links() []string {
	var links []string
	parser, _ := wikiParser()
	doc := markdown.Parse(p.Body, parser)
	ast.WalkFunc(doc, func(node ast.Node, entering bool) ast.WalkStatus {
		if entering {
			switch v := node.(type) {
			case *ast.Link:
				links = append(links, string(v.Destination))
			}
		}
		return ast.GoToNext
	})
	return links
}
