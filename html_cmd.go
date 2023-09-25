package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/subcommands"
	"io"
	"os"
)

type htmlCmd struct {
	useTemplate bool
}

func (*htmlCmd) Name() string     { return "html" }
func (*htmlCmd) Synopsis() string { return "Render a page as HTML." }
func (*htmlCmd) Usage() string {
	return `html [-view] <page name>:
  Render a page as HTML.
`
}

func (cmd *htmlCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&cmd.useTemplate, "view", false, "Use the 'view.html' template.")
}

func (cmd *htmlCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	return htmlCli(os.Stdout, cmd.useTemplate, f.Args())
}

func htmlCli(w io.Writer, useTemplate bool, args []string) subcommands.ExitStatus {
	for _, arg := range args {
		p, err := loadPage(arg)
		if err != nil {
			fmt.Fprintf(w, "Cannot load %s: %s\n", arg, err)
			return subcommands.ExitFailure
		}
		initAccounts()
		if useTemplate {
			p.handleTitle(true)
			p.renderHtml()
			t := "view.html"
			err := templates.ExecuteTemplate(w, t, p)
			if err != nil {
				fmt.Fprintf(w, "Cannot execute %s template for %s: %s\n", t, arg, err)
				return subcommands.ExitFailure
			}
		} else {
			// do not handle title
			p.renderHtml()
			fmt.Fprintln(w, p.Html)
		}
	}
	return subcommands.ExitSuccess
}
