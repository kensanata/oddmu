package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"path/filepath"
	"github.com/google/subcommands"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"io"
        "net/url"
	"os"
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
	names := make(map[string]bool);
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
		fmt.Println(err)
		return subcommands.ExitFailure
	}
	fmt.Println("Page\tMissing")
	for name, isPage := range names {
		if !isPage {
			continue
		}
		p, err := loadPage(name)
		if err != nil {
			fmt.Printf("Loading %s: %s\n", name, err)
			return subcommands.ExitFailure
		}
		for _, link := range p.links() {
			if !strings.HasPrefix(link, "/") &&
				!strings.HasPrefix(link, "http:") &&
				!strings.HasPrefix(link, "https:") &&
				!strings.HasPrefix(link, "mailto:") &&
				!strings.HasPrefix(link, "gopher:") &&
				!strings.HasPrefix(link, "gemini:") {
                                destination, err := url.PathUnescape(link)
                                if err != nil {
                                	fmt.Printf("Cannot decode %s: %s\n", link, err)
                                	return subcommands.ExitFailure
                                }
				_, ok := names[destination]
				if !ok {
					fmt.Printf("%s\t%s\n", name, link)
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
		switch v := node.(type) {
		case *ast.Link:
			links = append(links, string(v.Destination))
		}
		return ast.GoToNext
	})
	return links
}
