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
func (*htmlCmd) Synopsis() string { return "render a page as HTML" }
func (*htmlCmd) Usage() string {
	return `html [-view] <page name>:
  Render a page as HTML.
`
}

func (cmd *htmlCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&cmd.useTemplate, "view", false, "use the 'view.html' template.")
}

func (cmd *htmlCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	return htmlCli(os.Stdout, cmd.useTemplate, f.Args())
}

func htmlCli(w io.Writer, useTemplate bool, args []string) subcommands.ExitStatus {
	for _, arg := range args {
		p, err := loadPage(arg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot load %s: %s\n", arg, err)
			return subcommands.ExitFailure
		}
		if useTemplate {
			p.handleTitle(true)
			p.renderHtml()
			t := "view.html"
			loadTemplates()
			err := templates.template[t].Execute(w, p)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Cannot execute %s template for %s: %s\n", t, arg, err)
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
