package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/subcommands"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type staticCmd struct {
}

func (*staticCmd) Name() string     { return "static" }
func (*staticCmd) Synopsis() string { return "Render site into static HTML files." }
func (*staticCmd) Usage() string {
	return `static <dir name>:
  Create static copies in the given directory.
`
}

func (cmd *staticCmd) SetFlags(f *flag.FlagSet) {
}

func (cmd *staticCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	return staticCli(os.Stdout, f.Args())
}

func staticCli(w io.Writer, args []string) subcommands.ExitStatus {
	if len(args) != 1 {
		fmt.Fprintln(w, "Exactly one target directory is required")
		return subcommands.ExitFailure
	}
	dir := args[0]
	err := os.Mkdir(dir, 0755)
	if err != nil {
		fmt.Fprintf(w, "Cannot create %s: %s\n", dir, err)
		return subcommands.ExitFailure
	}
	initAccounts()
	err = filepath.Walk(".", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		filename := path
		if strings.HasPrefix(filename, ".") || strings.HasSuffix(filename, "~") {
			return nil
		}
		if info.IsDir() {
			return os.Mkdir(filepath.Join(dir, filename), 0755)
		}
		if strings.HasSuffix(filename, ".md") {
			name := strings.TrimSuffix(filename, ".md")
			p, err := loadPage(name)
			if err != nil {
				fmt.Fprintf(w, "Cannot load %s: %s\n", name, err)
				return err
			}
			p.handleTitle(true)
			p.renderHtml()
			t := "view.html"
			f, err := os.Create(filepath.Join(dir, name + ".html"))
			if err != nil {
				fmt.Fprintf(w, "Cannot create %s in %s: %s\n", filename, dir, err)
				return err
			}
			err = templates.ExecuteTemplate(f, t, p)
			if err != nil {
				fmt.Fprintf(w, "Cannot execute %s template for %s: %s\n", t, name, err)
				return err
			}
			return nil
		}
		return os.Link(filename, filepath.Join(dir, filename))
	})
	if err != nil {
		// no message needed
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}
