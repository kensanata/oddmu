package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/subcommands"
	"os"
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
			return subcommands.ExitFailure
		}
		initAccounts()
		if cmd.template {
			p.handleTitle(true)
			p.renderHtml()
			t := "view.html"
			err := templates.ExecuteTemplate(os.Stdout, t, p)
			if err != nil {
				fmt.Printf("Cannot execute %s template for %s: %s\n", t, arg, err)
				return subcommands.ExitFailure
			}
		} else {
			// do not handle title
			p.renderHtml()
			fmt.Println(p.Html)
		}
	}
	return subcommands.ExitSuccess
}
