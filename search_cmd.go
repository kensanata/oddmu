package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"github.com/google/subcommands"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type searchCmd struct {
	page  int
	exact bool
}

func (cmd *searchCmd) SetFlags(f *flag.FlagSet) {
	f.IntVar(&cmd.page, "page", 1, "the page in the search result set")
	f.BoolVar(&cmd.exact, "exact", false, "look for exact matches (do not use the trigram index)")
}

func (*searchCmd) Name() string     { return "search" }
func (*searchCmd) Synopsis() string { return "Search pages and print a list of links." }
func (*searchCmd) Usage() string {
	return `search [-page <n>] <terms>:
  Search for pages matching terms and print the result set as a
  Markdown list. Before searching, all the pages are indexed. Thus,
  startup is slow. The benefit is that the page order and scores are
  exactly as when the wiki runs.
`
}

func (cmd *searchCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	return searchCli(os.Stdout, cmd.page, cmd.exact, f.Args())
}

// searchCli runs the search command on the command line. It is used
// here with an io.Writer for easy testing.
func searchCli(w io.Writer, n int, exact bool, args []string) subcommands.ExitStatus {
	var fn func(q string, n int) ([]*Page, bool, int)
	if exact {
		fn = searchExact
	} else {
		index.load()
		fn = search
	}
	for _, q := range args {
		items, more, _ := fn(q, n)
		if len(items) == 1 {
			fmt.Fprintf(w, "Search for %s, page %d: 1 result\n", q, n)
		} else {
			fmt.Fprintf(w, "Search for %s, page %d: %d results\n", q, n, len(items))
		}
		for _, p := range items {
			fmt.Fprintf(w, "* [%s](%s) (%d)\n", p.Title, p.Name, p.Score)
		}
		if more {
			fmt.Fprintf(w, "There are more results\n")
		}
	}
	return subcommands.ExitSuccess
}

// searchExact opens all the files and searches them, one by one.
func searchExact(q string, page int) ([]*Page, bool, int) {
	if len(q) == 0 {
		return make([]*Page, 0), false, 0
	}
	terms := bytes.Fields([]byte(q))
	pages := make(map[string]*Page)
	names := make([]string, 0)
	index.titles = make(map[string]string)
	err := filepath.Walk(".", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		filename := path
		if info.IsDir() || strings.HasPrefix(filename, ".") || !strings.HasSuffix(filename, ".md") {
			return nil
		}
		name := strings.TrimSuffix(filename, ".md")
		p, err := loadPage(name)
		if err != nil {
			return err
		}
		for _, term := range terms {
			if !bytes.Contains(p.Body, term) {
				return nil
			}
		}
		p.handleTitle(false)
		pages[p.Name] = p
		index.titles[p.Name] = p.Title
		names = append(names, p.Name)
		return nil
	})
	if err != nil {
		return make([]*Page, 0), false, 0
	}
	slices.SortFunc(names, sortNames(q))
	from := itemsPerPage * (page - 1)
	if from > len(names) {
		return make([]*Page, 0), false, 0
	}
	to := from + itemsPerPage
	if to > len(names) {
		to = len(names)
	}
	items := make([]*Page, 0)
	for i := from; i < to; i++ {
		p := pages[names[i]]
		p.score(q)
		p.summarize(q)
		items = append(items, p)
	}
	return items, to < len(names), len(names)/itemsPerPage + 1
}
