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
	regexp  bool
}

func (cmd *replaceCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&cmd.confirm, "confirm", false, "do the replacement instead of just doing a dry run")
	f.BoolVar(&cmd.regexp, "regexp", false, "the search string is a regular expression")
}

func (*replaceCmd) Name() string     { return "replace" }
func (*replaceCmd) Synopsis() string { return "search and replace in all the pages" }
func (*replaceCmd) Usage() string {
	return `replace [-confirm] [-regexp] <term> <replacement>:
  Search a string or a regular expression and replace it. By default,
  this is a dry run and nothing is saved. If this is a regular
  expression, the replacement can use $1, $2, etc. to refer to capture
  groups in the regular expression.
`
}

func (cmd *replaceCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	return replaceCli(os.Stdout, cmd.confirm, cmd.regexp, f.Args())
}

func replaceCli(w io.Writer, isConfirmed bool, isRegexp bool, args []string) subcommands.ExitStatus {
	if len(args) != 2 {
		fmt.Fprintln(os.Stderr, "Replace takes exactly two arguments.")
		return subcommands.ExitFailure
	}
	var re *regexp.Regexp
	if isRegexp {
		re = regexp.MustCompile(args[0])
	} else {
		re = regexp.MustCompile(regexp.QuoteMeta(args[0]))
	}
	repl := []byte(args[1])
	changes := 0
	err := filepath.Walk(".", func(fp string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// skip hidden directories and files
		if fp != "." && strings.HasPrefix(filepath.Base(fp), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			} else {
				return nil
			}
		}
		// skipp all but page files
		if !strings.HasSuffix(fp, ".md") {
			return nil
		}
		body, err := os.ReadFile(fp)
		if err != nil {
			return err
		}
		result := re.ReplaceAll(body, repl)
		if !slices.Equal(result, body) {
			changes++
			if isConfirmed {
				fmt.Fprintln(w, fp)
				_ = os.Rename(fp, fp + "~")
				err = os.WriteFile(fp, result, 0644)
				if err != nil {
					return err
				}
			} else {
				edits := myers.ComputeEdits(span.URIFromPath(fp + "~"), string(body), string(result))
				diff := fmt.Sprint(gotextdiff.ToUnified(fp + "~", fp, string(body), edits))
				fmt.Fprintln(w, diff)
			}
		}
		return nil
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return subcommands.ExitFailure
	}
	if changes == 1 {
		if isConfirmed {
			fmt.Fprintln(w, "1 file was changed.")
		} else {
			fmt.Fprintln(w, "1 file would be changed.")
		}
	} else {
		if isConfirmed {
			fmt.Fprintf(w, "%d files were changed.\n", changes)
		} else {
			fmt.Fprintf(w, "%d files would be changed.\n", changes)
		}
	}
	if !isConfirmed && changes > 0 {
		fmt.Fprintln(w, "This is a dry run. Use -confirm to make it happen.")
	}
	return subcommands.ExitSuccess
}
