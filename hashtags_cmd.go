package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/subcommands"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"
)

type hashtagsCmd struct {
	update bool
	dryRun bool
}

func (cmd *hashtagsCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&cmd.update, "update", false, "create and update hashtag pages")
	f.BoolVar(&cmd.dryRun, "dry-run", false, "only report the changes it would make")
}

func (*hashtagsCmd) Name() string     { return "hashtags" }
func (*hashtagsCmd) Synopsis() string { return "hashtag overview" }
func (*hashtagsCmd) Usage() string {
	return `hashtags:
  Count the use of all hashtags and list them, separated by a tabulator.
`
}

func (cmd *hashtagsCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if cmd.update {
		return hashtagsUpdateCli(os.Stdout, cmd.dryRun)
	}
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
		fmt.Fprintf(w, "%d\t%s\t%d\n", i+1, hashtag.label, hashtag.count)
	}

	return subcommands.ExitSuccess
}

// hashtagsUpdateCli runs the hashtags command on the command line and creates and updates the hashtag pages in the
// current directory. That is, pages in subdirectories are skipped! It is used here with an io.Writer for easy testing.
func hashtagsUpdateCli(w io.Writer, dryRun bool) subcommands.ExitStatus {
	index.load()
	// no locking necessary since this is for the command-line
	namesMap := make(map[string]string)
	for hashtag, docids := range index.token {
		if len(docids) <= 5 {
			if dryRun {
				fmt.Fprintf(w, "Skipping #%s because there are not enough entries (%d)\n", hashtag, len(docids))
			}
			continue
		}
		title, ok := namesMap[hashtag]
		if (!ok) {
			title = hashtagName(namesMap, hashtag, docids)
			namesMap[hashtag] = title
		}
		pageName := strings.ReplaceAll(title, " ", "_")
		h, err := loadPage(pageName)
		original := ""
		new := false
		if err != nil {
			new = true
			h = &Page{Name: pageName, Body: []byte("# " + title + "\n\n#" + pageName + "\n\nBlog posts:\n\n")}
		} else {
			original = string(h.Body)
		}
		for _, docid := range docids {
			name := index.documents[docid]
			if strings.Contains(name, "/") {
				continue
			}
			p, err := loadPage(name)
			if err != nil {
				fmt.Fprintf(w, "Loading %s: %s\n", name, err)
				return subcommands.ExitFailure
			}
			if !p.IsBlog() {
				continue
			}
			p.handleTitle(false)
			if p.Title == "" {
				p.Title = p.Name
			}
			esc := nameEscape(p.Base())
			link := "* [" + p.Title + "](" + esc + ")\n"
			// I guess & used to get escaped and now no longer does
			re := regexp.MustCompile(`(?m)^\* \[[^\]]+\]\(` + strings.ReplaceAll(esc, "&", "(&|%26)") + `\)\n`)
			addLinkToPage(h, link, re)
		}
		// only save if something changed
		if string(h.Body) != original {
			if dryRun {
				if new {
					fmt.Fprintf(w, "Creating %s.md\n", title)
				} else {
					fmt.Fprintf(w, "Updating %s.md\n", title)
				}
				fn := h.Name + ".md"
				edits := myers.ComputeEdits(span.URIFromPath(fn), original, string(h.Body))
				diff := fmt.Sprint(gotextdiff.ToUnified(fn + "~", fn, original, edits))
				fmt.Fprint(w, diff)
			} else {
				err = h.save()
				if err != nil {
					fmt.Fprintf(w, "Saving hashtag %s failed: %s", hashtag, err)
					return subcommands.ExitFailure
				}
			}
		}
	}
	return subcommands.ExitSuccess
}

// Go through all the documents in the same directory and look for hashtag matches in the rendered HTML in order to
// determine the most likely capitalization.
func hashtagName (namesMap map[string]string, hashtag string, docids []docid) string {
	candidate := make(map[string]int)
	var mostPopular string
	for _, docid := range docids {
		name := index.documents[docid]
		if strings.Contains(name, "/") {
			continue
		}
		p, err := loadPage(name)
		if err != nil {
			continue
		}
		// parsing finds all the hashtags
		parser, _ := wikiParser()
		doc := markdown.Parse(p.Body, parser)
		ast.WalkFunc(doc, func(node ast.Node, entering bool) ast.WalkStatus {
			if entering {
				switch v := node.(type) {
				case *ast.Link:
					for _, attr := range v.AdditionalAttributes {
						if attr == `class="tag"` {
							tagName := []byte("")
							ast.WalkFunc(v, func(node ast.Node, entering bool) ast.WalkStatus {
								if entering && node.AsLeaf() != nil {
									tagName = append(tagName, node.AsLeaf().Literal...)
								}
								return ast.GoToNext
							})
							tag := string(tagName[1:])
							if strings.EqualFold(hashtag, strings.ReplaceAll(tag, " ", "_")) {
								_, ok := candidate[tag]
								if ok {
									candidate[tag] += 1
								} else {
									candidate[tag] = 1
								}
							}
						}
					}
				}
			}
			return ast.GoToNext
		})
		count := 0
		for key, val := range candidate {
			if val > count {
				mostPopular = key
				count = val
			}
		}
		// shortcut
		if count >= 5 {
			return mostPopular
		}
	}
	return mostPopular
}
