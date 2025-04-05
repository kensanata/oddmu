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

type htmlCmd struct {
	template string
}

func (*htmlCmd) Name() string     { return "html" }
func (*htmlCmd) Synopsis() string { return "render a page as HTML" }
func (*htmlCmd) Usage() string {
	return `html [-template <template name>] <page name> ...:
  Render one or more pages as HTML.
  Use a single - to read Markdown from stdin.
`
}

func (cmd *htmlCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&cmd.template, "template", "",
		"use the given HTML file as a template (probably view.html or static.html).")
}

func (cmd *htmlCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	return htmlCli(os.Stdout, cmd.template, f.Args())
}

func htmlCli(w io.Writer, template string, args []string) subcommands.ExitStatus {
	if len(args) == 1 && args[0] == "-" {
		body, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(w, "Cannot read from stdin: %s\n", err)
			return subcommands.ExitFailure
		}
		p := &Page{Name: "stdin", Body: body}
		return p.printHtml(w, template)
	}
	for _, name := range args {
		if !strings.HasSuffix(name, ".md") {
			fmt.Fprintf(os.Stderr, "%s does not end in '.md'\n", name)
			return subcommands.ExitFailure
		}
		name = name[0:len(name)-3]
		p, err := loadPage(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot load %s: %s\n", name, err)
			return subcommands.ExitFailure
		}
		status := p.printHtml(w, template)
		if status != subcommands.ExitSuccess {
			return status
		}
	}
	return subcommands.ExitSuccess
}

func (p *Page) printHtml(w io.Writer, template string) subcommands.ExitStatus {
	if len(template) > 0 {
		t := template
		loadTemplates()
		p.handleTitle(true)
		p.renderHtml()
		err := templates.template[t].Execute(w, p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot execute %s template for %s: %s\n", t, p.Name, err)
			return subcommands.ExitFailure
		}
	} else {
		// do not handle title
		p.renderHtml()
		fmt.Fprintln(w, p.Html)
	}
	return subcommands.ExitSuccess
}
