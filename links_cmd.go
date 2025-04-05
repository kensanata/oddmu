package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/subcommands"
	"io"
	"os"
	"strings"
)

type linksCmd struct {
}

func (cmd *linksCmd) SetFlags(f *flag.FlagSet) {
}

func (*linksCmd) Name() string     { return "links" }
func (*linksCmd) Synopsis() string { return "list outgoing links for a page" }
func (*linksCmd) Usage() string {
	return `links <page name> ...:
  Lists all the links on a page. Use a single - to read Markdown from stdin.
`
}

func (cmd *linksCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	return linksCli(os.Stdout, f.Args())
}

// linksCli runs the links command on the command line. It is used
// here with an io.Writer for easy testing.
func linksCli(w io.Writer, args []string) subcommands.ExitStatus {
	if len(args) == 1 && args[0] == "-" {
		body, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(w, "Cannot read from stdin: %s\n", err)
			return subcommands.ExitFailure
		}
		p := &Page{Body: body}
		for _, link := range p.links() {
			fmt.Fprintln(w, link)
		}
		return subcommands.ExitSuccess
	}
	for _, name := range args {
		if !strings.HasSuffix(name, ".md") {
			fmt.Fprintf(os.Stderr, "%s does not end in '.md'\n", name)
			return subcommands.ExitFailure
		}
		name = name[0:len(name)-3]
		p, err := loadPage(name)
		if err != nil {
			fmt.Fprintf(w, "Loading %s: %s\n", name, err)
			return subcommands.ExitFailure
		}
		for _, link := range p.links() {
			fmt.Fprintln(w, link)
		}
	}
	return subcommands.ExitSuccess
}
