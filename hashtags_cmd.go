package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/subcommands"
	"io"
	"os"
	"sort"
)

type hashtagsCmd struct {
}

func (cmd *hashtagsCmd) SetFlags(f *flag.FlagSet) {
}

func (*hashtagsCmd) Name() string     { return "hashtags" }
func (*hashtagsCmd) Synopsis() string { return "Hashtag overview." }
func (*hashtagsCmd) Usage() string {
	return `hashtags:
  Count the use of all hashtags and list them, separated by a tabulator.
`
}

func (cmd *hashtagsCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	return hashtagsCli(os.Stdout)
}

// hashtagsCli runs the hashtags command on the command line. It is used
// here with an io.Writer for easy testing.
func hashtagsCli(w io.Writer) subcommands.ExitStatus {
	index.load()
	index.RLock()
	defer index.RUnlock()

	type hashtag struct {
		label string
		count int
	}

	hashtags := []hashtag{}

	for token, docids := range index.token {
		hashtags = append(hashtags, hashtag{label: token, count: len(docids)})
	}

	sort.Slice(hashtags, func(i, j int) bool {
		return hashtags[i].count > hashtags[j].count
	})

	fmt.Fprintln(w, "Rank\tHashtag\tCount")
	for i, hashtag := range hashtags {
		fmt.Fprintf(w, "%d\t%s\t%d\n", i, hashtag.label, hashtag.count)
	}

	return subcommands.ExitSuccess
}
