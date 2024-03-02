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
	return versionCli(os.Stdout, cmd.full)
}

func versionCli(w io.Writer, full bool) subcommands.ExitStatus {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		w.Write([]byte("This binary contains no debug info.\n"))
	} else if full {
		w.Write([]byte(info.String()))
	} else {
		w.Write([]byte(info.Path + "\n"))
		for _, setting := range info.Settings {
			if strings.HasPrefix(setting.Key, "vcs") {
				_, err := fmt.Fprintf(w, "%s=%s\n", setting.Key, setting.Value)
				if err != nil {
					fmt.Fprintln(os.Stderr, err.Error())
					return subcommands.ExitFailure
				}
			}
		}
	}
	return subcommands.ExitSuccess
}
