package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/subcommands"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type listCmd struct {
	dir string
}

func (cmd *listCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&cmd.dir, "dir", "", "list only pages within this sub-directory")
}

func (*listCmd) Name() string     { return "list" }
func (*listCmd) Synopsis() string { return "List pages with name and title." }
func (*listCmd) Usage() string {
	return `list [-dir string]:
  List all pages with name and title, separated by a tabulator.
`
}

func (cmd *listCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	return listCli(os.Stdout, cmd.dir, f.Args())
}

// listCli runs the list command on the command line. It is used
// here with an io.Writer for easy testing.
func listCli(w io.Writer, dir string, args []string) subcommands.ExitStatus {
	if dir != "" {
		fi, err := os.Stat(dir)
		if err != nil {
			fmt.Println(err)
			return subcommands.ExitFailure
		}
		if !fi.IsDir() {
			fmt.Println("This is not a sub-directory:", dir)
			return subcommands.ExitFailure
		}
		dir = filepath.ToSlash(dir);
		if (!strings.HasSuffix(dir, "/")) {
			dir += "/"
		}
	}
	index.load()
	index.RLock()
	defer index.RUnlock()
	for name, title := range index.titles {
		if strings.HasPrefix(name, dir) {
			name = strings.Replace(name, dir, "", 1)
			fmt.Fprintf(w, "%s\t%s\n", name, title)
		}
	}
	return subcommands.ExitSuccess
}
