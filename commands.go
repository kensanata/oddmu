package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"github.com/google/subcommands"
	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
)

type htmlCmd struct {
	template bool
}

func (*htmlCmd) Name() string     { return "html" }
func (*htmlCmd) Synopsis() string { return "Render a page as HTML." }
func (*htmlCmd) Usage() string {
  return `html [-view] <page name>:
  Render a page as HTML.
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

type replaceCmd struct {
	confirm bool
}

func (cmd *replaceCmd) SetFlags(f *flag.FlagSet) {
  f.BoolVar(&cmd.confirm, "confirm", false, "do the replacement instead of just doing a dry run")
}

func (*replaceCmd) Name() string     { return "replace" }
func (*replaceCmd) Synopsis() string { return "Search and replace a regular expression." }
func (*replaceCmd) Usage() string {
  return `replace [-confirm] <regexp> <replacement>:
  Search a regular expression and replace it. By default, this is a
  dry run and nothing is saved. The replacement can use $1, $2, etc.
  to refer to capture groups in the regular expression.
`
}

func (cmd *replaceCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	args := f.Args()
	if len(args) != 2 {
		fmt.Println("Replace takes exactly two arguments.")
		return subcommands.ExitFailure
	}
	re := regexp.MustCompile(args[0])
	repl := []byte(args[1])
	changes := 0
	err := filepath.Walk(".", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || strings.HasPrefix(path, ".") || !strings.HasSuffix(path, ".md") {
			return nil
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		result := re.ReplaceAll(body, repl)
		if !slices.Equal(result, body) {
			changes++
			if !cmd.confirm {
				edits := myers.ComputeEdits(span.URIFromPath(path + "~"), string(body), string(result))
				diff := fmt.Sprint(gotextdiff.ToUnified(path + "~", path, string(body), edits))
				fmt.Println(diff)
			} else {
				fmt.Println(path)
				_ = os.Rename(path, path+"~")
				err = os.WriteFile(path, result, 0644)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}
	if changes == 1 {
		fmt.Println("1 change was made.")
	} else {
		fmt.Printf("%d changes were made.\n", changes)
	}
	if !cmd.confirm && changes > 0 {
		fmt.Println("This is a dry run. Use -confirm to make it happen.")
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
	subcommands.Register(&replaceCmd{}, "")

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
