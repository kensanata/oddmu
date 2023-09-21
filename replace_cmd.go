package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/subcommands"
	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

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
	return replaceCli(os.Stdout, cmd.confirm, f.Args())
}

func replaceCli(w io.Writer, confirm bool, args []string) subcommands.ExitStatus {
	if len(args) != 2 {
		fmt.Fprintln(w, "Replace takes exactly two arguments.")
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
			if confirm {
				fmt.Fprintln(w, path)
				_ = os.Rename(path, path+"~")
				err = os.WriteFile(path, result, 0644)
				if err != nil {
					return err
				}
			} else {
				edits := myers.ComputeEdits(span.URIFromPath(path+"~"), string(body), string(result))
				diff := fmt.Sprint(gotextdiff.ToUnified(path+"~", path, string(body), edits))
				fmt.Fprintln(w, diff)
			}
		}
		return nil
	})
	if err != nil {
		fmt.Fprintln(w, err)
		return subcommands.ExitFailure
	}
	if changes == 1 {
		fmt.Fprintln(w, "1 change was made.")
	} else {
		fmt.Fprintf(w, "%d changes were made.\n", changes)
	}
	if !confirm && changes > 0 {
		fmt.Fprintln(w, "This is a dry run. Use -confirm to make it happen.")
	}
	return subcommands.ExitSuccess
}
