package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/subcommands"
	"io"
	"os"
)

type listCmd struct {
}

func (cmd *listCmd) SetFlags(f *flag.FlagSet) {
}

func (*listCmd) Name() string     { return "list" }
func (*listCmd) Synopsis() string { return "List pages with name and title." }
func (*listCmd) Usage() string {
	return `list:
  List all pages with name and title, separated by a tabulator.
`
}

func (cmd *listCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	return listCli(os.Stdout, f.Args())
}

// listCli runs the list command on the command line. It is used
// here with an io.Writer for easy testing.
func listCli(w io.Writer, args []string) subcommands.ExitStatus {
	index.load()
	index.RLock()
	defer index.RUnlock()
	for name, title := range index.titles {
		fmt.Fprintf(w, "%s\t%s\n", name, title)
	}
	return subcommands.ExitSuccess
}
