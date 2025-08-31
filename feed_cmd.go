package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/subcommands"
	"io"
	"os"
	"strings"
	"time"
)

type feedCmd struct {
}

func (*feedCmd) Name() string     { return "feed" }
func (*feedCmd) Synopsis() string { return "render a page as feed" }
func (*feedCmd) Usage() string {
	return `feed <page name> ...:
  Render one or more pages as a single feed.
  Use a single - to read Markdown from stdin.
`
}

func (cmd *feedCmd) SetFlags(f *flag.FlagSet) {
}

func (cmd *feedCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	return feedCli(os.Stdout, f.Args())
}

func feedCli(w io.Writer, args []string) subcommands.ExitStatus {
	if len(args) == 1 && args[0] == "-" {
		body, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(w, "Cannot read from stdin: %s\n", err)
			return subcommands.ExitFailure
		}
		p := &Page{Name: "stdin", Body: body}
		return p.printFeed(w, time.Now())
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
		ti, _ := p.ModTime()
		status := p.printFeed(w, ti)
		if status != subcommands.ExitSuccess {
			return status
		}
	}
	return subcommands.ExitSuccess
}

func (p *Page) printFeed(w io.Writer, ti time.Time) subcommands.ExitStatus {
	f := feed(p, ti, 0)
	if len(f.Items) == 0 {
		fmt.Fprintf(os.Stderr, "Empty feed for %s\n", p.Name)
		return subcommands.ExitFailure
	}
	_, err := w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>`))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot write prefix: %s\n", err)
		return subcommands.ExitFailure
	}
	loadTemplates()
	templates.RLock()
	defer templates.RUnlock()
	err = templates.template["feed.html"].Execute(w, f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot execute template: %s\n", err)
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}
