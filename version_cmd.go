package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/subcommands"
	"io"
	"os"
	"runtime/debug"
)

type versionCmd struct {
}

func (cmd *versionCmd) SetFlags(f *flag.FlagSet) {
}

func (*versionCmd) Name() string     { return "version" }
func (*versionCmd) Synopsis() string { return "report build information" }
func (*versionCmd) Usage() string {
	return `version:
  Report all the debug information about this build.
`
}

func (cmd *versionCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	return versionCli(os.Stdout, f.Args())
}

func versionCli(w io.Writer, args []string) subcommands.ExitStatus {
	if len(args) > 0 {
		fmt.Fprintln(os.Stderr, "Version takes no arguments.")
		return subcommands.ExitFailure
	}
	info, _ := debug.ReadBuildInfo()
	fmt.Println(info)
	return subcommands.ExitSuccess
}
