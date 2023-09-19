package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"github.com/google/subcommands"
)

type htmlCmd struct {
	template bool
}

func (*htmlCmd) Name() string     { return "html" }
func (*htmlCmd) Synopsis() string { return "Render a page as HTML." }
func (*htmlCmd) Usage() string {
  return `html <page name>:
  Render a page as HTML
`
}

func (cmd *htmlCmd) SetFlags(f *flag.FlagSet) {
  f.BoolVar(&cmd.template, "view", false, "Use the 'view.html' template.")
}

func (cmd *htmlCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	for _, arg := range f.Args() {
		p, err := loadPage(arg)
		if err != nil {
			fmt.Printf("Cannot load %s: %s\n", arg, err)
		} else {
			if cmd.template {
				p.handleTitle(true)
				p.renderHtml()
				t := "view.html"
				err := templates.ExecuteTemplate(os.Stdout, t, p)
				if err != nil {
					fmt.Printf("Cannot execute %t template for %s: %s\n", t, arg, err)
					return subcommands.ExitFailure
				}
			} else {
				// do not handle title
				p.renderHtml()
				fmt.Println(p.Html)
			}
		}
	}
	return subcommands.ExitSuccess
}

type searchCmd struct {
	page int
}

func (cmd *searchCmd) SetFlags(f *flag.FlagSet) {
  f.IntVar(&cmd.page, "page", 1, "the page in the search result set")
}

func (*searchCmd) Name() string     { return "search" }
func (*searchCmd) Synopsis() string { return "Search pages and print a list of links." }
func (*searchCmd) Usage() string {
  return `search <terms>:
  Search for pages matching terms and print the result set as a
  Markdown list. Before searching, it indexes all the pages. Thus,
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

// commands does the command line parsing in case Oddmu is called with
// some arguments. Without any arguments, the wiki server is started.
// At this point we already know that there is at least one
// subcommand.
func commands() {
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(&htmlCmd{}, "")
	subcommands.Register(&searchCmd{}, "")

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
