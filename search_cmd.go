package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/subcommands"
	"github.com/muesli/reflow/wordwrap"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type searchCmd struct {
	dir     string
	page    int
	all     bool
	extract bool
	files   bool
	quiet   bool
}

func (cmd *searchCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&cmd.dir, "dir", "", "search only pages within this sub-directory")
	f.IntVar(&cmd.page, "page", 1, "the page in the search result set, default 1")
	f.BoolVar(&cmd.all, "all", false, "show all the pages and ignore -page")
	f.BoolVar(&cmd.extract, "extract", false, "print page extract instead of link list")
	f.BoolVar(&cmd.files, "files", false, "show just the filenames")
	f.BoolVar(&cmd.quiet, "quiet", false, "suppress summary line at the top")
}

func (*searchCmd) Name() string     { return "search" }
func (*searchCmd) Synopsis() string { return "search pages and print a list of links" }
func (*searchCmd) Usage() string {
	return `search [-dir string] [-page <n>|-all] [-extract|-files] [-quiet] <terms>:
  Search for pages matching terms and print the result set as a
  Markdown list. Before searching, all the pages are indexed. Thus,
  startup is slow. The benefit is that the page order is exactly as
  when the wiki runs.
`
}

func (cmd *searchCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	return searchCli(os.Stdout, cmd, f.Args())
}

// searchCli runs the search command on the command line. It is used
// here with an io.Writer for easy testing.
func searchCli(w io.Writer, cmd *searchCmd, args []string) subcommands.ExitStatus {
	dir, err := checkDir(cmd.dir)
	if err != nil {
		return subcommands.ExitFailure
	}
	index.reset()
	index.load()
	q := strings.Join(args, " ")
	items, more := search(q, dir, "", cmd.page, true)
	if !cmd.quiet {
		fmt.Fprint(os.Stderr, "Search for ", q)
		if !cmd.all {
			fmt.Fprint(os.Stderr, ", page ", cmd.page)
		}
		fmt.Fprint(os.Stderr, ": ", len(items))
		if len(items) == 1 {
			fmt.Fprint(os.Stderr, " result\n")
		} else {
			fmt.Fprint(os.Stderr, " results\n")
		}
	}
	if cmd.extract {
		searchExtract(w, items)
	} else if cmd.files {
		for _, p := range items {
			name := filepath.FromSlash(p.Name) + ".md\n"
			fmt.Fprintf(w, name)
		}
	} else {
		for _, p := range items {
			name := p.Name
			if strings.HasPrefix(name, dir) {
				name = strings.Replace(name, dir, "", 1)
			}
			fmt.Fprintf(w, "* [%s](%s)\n", p.Title, name)
		}
	}
	if more {
		fmt.Fprintf(os.Stderr, "There are more results\n")
	}
	return subcommands.ExitSuccess
}

// searchExtract prints the search extracts to stdout with highlighting for a terminal.
func searchExtract(w io.Writer, items []*Result) {
	heading := func(s string) string { return "\x1b[1;4m" + s + "\x1b[0m" } // bold + underline
	match := func(s string) string { return "\x1b[1m" + s + "\x1b[0m" }     // bold
	re := regexp.MustCompile(`<b>(.*?)</b>`)
	for _, p := range items {
		s := re.ReplaceAllString(string(p.Html), match(`$1`))
		fmt.Fprintln(w, heading(p.Title))
		if p.Name != p.Title {
			fmt.Fprintln(w, p.Name)
		}
		for _, s := range strings.Split(wordwrap.String(s, 72), "\n") {
			fmt.Fprintln(w, "    ", s)
		}
		for _, img := range p.Images {
			name, err := url.PathUnescape(img.Name)
			if err != nil {
				name = img.Name
			}
			fmt.Fprintln(w, "    - ", name)
			for _, s := range strings.Split(wordwrap.String(img.Title, 70), "\n") {
				fmt.Fprintln(w, "      ", s)
			}
		}
	}
}
