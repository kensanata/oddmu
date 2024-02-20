package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/subcommands"
	"io"
	"os"
	"runtime/debug"
	"strings"
)

type versionCmd struct {
	full bool
}

func (cmd *versionCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&cmd.full, "full", false, "show all the debug information")
}

func (*versionCmd) Name() string     { return "version" }
func (*versionCmd) Synopsis() string { return "report build information" }
func (*versionCmd) Usage() string {
	return `version [-full]:
  Report the exact version control commit this is built from, or the
  full debug information about this build.
`
}

func (cmd *versionCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	return versionCli(os.Stdout, cmd.full, f.Args())
}

func versionCli(w io.Writer, full bool, args []string) subcommands.ExitStatus {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		fmt.Println("This binary contains no debug info.")
	} else if full {
		fmt.Println(info)
	} else {
		fmt.Println(info.Path)
		for _, setting := range info.Settings {
			if strings.HasPrefix(setting.Key, "vcs") {
				fmt.Printf("%s=%s\n", setting.Key, setting.Value)
			}
		}
	}
	return subcommands.ExitSuccess
}
