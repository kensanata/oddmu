package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/google/subcommands"
	"io"
	"os"
	"strings"
)

type tocCmd struct {
}

func (cmd *tocCmd) SetFlags(f *flag.FlagSet) {
}

func (*tocCmd) Name() string     { return "toc" }
func (*tocCmd) Synopsis() string { return "print the table of contents (toc) for a page" }
func (*tocCmd) Usage() string {
	return `toc <page name> ...:
  Print the table of contents (toc) for a page.
  Use a single - to read Markdown from stdin.
  If only a single level one heading is appears
  in the page, it is dropped from the table of
  contents.
`
}

func (cmd *tocCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	return tocCli(os.Stdout, f.Args())
}

// tocCli runs the toc command on the command line. It is used
// here with an io.Writer for easy testing.
func tocCli(w io.Writer, args []string) subcommands.ExitStatus {
	if len(args) == 1 && args[0] == "-" {
		body, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(w, "Cannot read from stdin: %s\n", err)
			return subcommands.ExitFailure
		}
		p := &Page{Body: body}
		p.toc().print(w)
		return subcommands.ExitSuccess
	}
	for _, name := range args {
		if !strings.HasSuffix(name, ".md") {
			fmt.Fprintf(os.Stderr, "%s does not end in '.md'\n", name)
			return subcommands.ExitFailure
		}
		name = name[0 : len(name)-3]
		p, err := loadPage(name)
		if err != nil {
			fmt.Fprintf(w, "Loading %s: %s\n", name, err)
			return subcommands.ExitFailure
		}
		p.toc().print(w)
	}
	return subcommands.ExitSuccess
}

// Toc represents an array of headings
type Toc []*ast.Heading

// toc parses the page content and returns a Toc.
func (p *Page) toc() Toc {
	var headings Toc
	parser, _ := wikiParser()
	doc := markdown.Parse(p.Body, parser)
	ast.WalkFunc(doc, func(node ast.Node, entering bool) ast.WalkStatus {
		if !entering {
			switch v := node.(type) {
			case *ast.Heading:
				headings = append(headings, v)
			}
		}
		return ast.GoToNext
	})
	return headings
}

// print prints the Toc to the io.Writer. If the table of contents first heading is a level one heading and there are no
// other level one headings, this is a "regular" table of contents. For a regular table of contents, the first entry is
// skipped.
func (toc Toc) print(w io.Writer) {
	minLevel := 0
	levelOneCount := 0
	for _, h := range toc {
		if h.Level == 1 {
			levelOneCount++
		}
		if h.Level < minLevel || minLevel == 0 {
			minLevel = h.Level
		}
	}
	for i, h := range toc {
		if i == 0 && h.Level == 1 && levelOneCount == 1 {
			minLevel++
			continue
		}
		for j := minLevel; j < h.Level; j++ {
			fmt.Fprint(w, "  ")
		}
		fmt.Fprint(w, "* [")
		for _, c := range h.GetChildren() {
			fmt.Fprint(w, string(c.AsLeaf().Literal))
		}
		fmt.Fprintf(w, "](#%s)\n", h.HeadingID)
	}
}
