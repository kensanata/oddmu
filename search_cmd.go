package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/subcommands"
	"io"
	"os"
	"regexp"
	"strings"
)

type searchCmd struct {
	page int
	extract bool
}

func (cmd *searchCmd) SetFlags(f *flag.FlagSet) {
	f.IntVar(&cmd.page, "page", 1, "the page in the search result set")
	f.BoolVar(&cmd.extract, "extract", false, "print page extract instead of link list")
}

func (*searchCmd) Name() string     { return "search" }
func (*searchCmd) Synopsis() string { return "Search pages and print a list of links." }
func (*searchCmd) Usage() string {
	return `search [-page <n>] <terms>:
  Search for pages matching terms and print the result set as a
  Markdown list. Before searching, all the pages are indexed. Thus,
  startup is slow. The benefit is that the page order is exactly as
  when the wiki runs.
`
}

func (cmd *searchCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	return searchCli(os.Stdout, cmd.page, cmd.extract, f.Args())
}

// searchCli runs the search command on the command line. It is used
// here with an io.Writer for easy testing.
func searchCli(w io.Writer, n int, extract bool, args []string) subcommands.ExitStatus {
	index.load()
	q := strings.Join(args, " ")
	items, more := search(q, ".", n)
	if len(items) == 1 {
		fmt.Fprintf(w, "Search for %s, page %d: 1 result\n", q, n)
	} else {
		fmt.Fprintf(w, "Search for %s, page %d: %d results\n", q, n, len(items))
	}
	if extract {
		heading := lipgloss.NewStyle().Bold(true).Underline(true)
		quote := lipgloss.NewStyle().PaddingLeft(4).Width(78)
		match := lipgloss.NewStyle().Bold(true)
		re := regexp.MustCompile(`<b>(.*?)</b>`)
		for _, p := range items {
			s := re.ReplaceAllString(string(p.Html), match.Render(`$1`))
			fmt.Fprintln(w, heading.Render(p.Title))
			if p.Name != p.Title {
				fmt.Fprintln(w, p.Name)
			}
			fmt.Fprintln(w, quote.Render(s))
		}		
	} else {
		for _, p := range items {
			fmt.Fprintf(w, "* [%s](%s)\n", p.Title, p.Name)
		}
	}
	if more {
		fmt.Fprintf(w, "There are more results\n")
	}
	return subcommands.ExitSuccess
}
