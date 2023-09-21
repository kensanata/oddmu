package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/subcommands"
)

type searchCmd struct {
	page int
}

func (cmd *searchCmd) SetFlags(f *flag.FlagSet) {
	f.IntVar(&cmd.page, "page", 1, "the page in the search result set")
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
	index.load()
	for _, q := range f.Args() {
		items, more, _ := search(q, cmd.page)
		if len(items) == 1 {
			fmt.Printf("Search for %s, page %d: 1 result\n", q, cmd.page)
		} else {
			fmt.Printf("Search for %s, page %d: %d results\n", q, cmd.page, len(items))
		}
		for _, p := range items {
			fmt.Printf("* [%s](%s) (%d)\n", p.Title, p.Name, p.Score)
		}
		if more {
			fmt.Printf("There are more results\n")
		}
	}
	return subcommands.ExitSuccess
}
