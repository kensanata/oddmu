package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/google/subcommands"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

type staticCmd struct {
	jobs int
}

func (cmd *staticCmd) SetFlags(f *flag.FlagSet) {
	f.IntVar(&cmd.jobs, "jobs", 2, "how many jobs to use")
}

func (*staticCmd) Name() string     { return "static" }
func (*staticCmd) Synopsis() string { return "generate static HTML files for all pages" }
func (*staticCmd) Usage() string {
	return `static [-jobs n] <dir name>:
  Create static copies in the given directory. Per default, two jobs
  are used to read and write files, but more can be assigned.
`
}

func (cmd *staticCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	args := f.Args()
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "Exactly one target directory is required")
		return subcommands.ExitFailure
	}
	dir := filepath.Clean(args[0])
	_, err := os.Stat(dir)
	if err == nil {
		fmt.Fprintf(os.Stderr, "%s already exists\n", dir)
		return subcommands.ExitFailure
	}
	return staticCli(dir, cmd.jobs, false)
}

type args struct { path string; info fs.FileInfo }

// staticCli generates a static site in the designated directory. The quiet flag is used to suppress output when running
// tests.
func staticCli(dir string, jobs int, quiet bool) subcommands.ExitStatus {
	index.load()
	index.RLock()
	defer index.RUnlock()
	loadLanguages()
	loadTemplates()
	tasks := make(chan args)
	results := make(chan error)
	done := make(chan bool)
	stop := make(chan error)
	for i := 0; i < jobs; i++ {
		go staticWorker(dir, tasks, results, done)
	}
	go staticWalk(dir, tasks, stop)
	go staticWatch(jobs, results, done)
	n, err := staticProgressIndicator(results, stop, quiet)
	if !quiet {
		fmt.Printf("\r%d files processed\n", n)
	}
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}

// staticWalk walks the directory tree. Any directory it finds, it recreates in the destination directory. Any file it
// finds, it puts into the tasks channel for the staticWorker. When the directory walk is finished, the tasks channel is
// closed. If there's an error on the stop channel, the walk returns that error.
func staticWalk (dir string, tasks chan(args), stop chan(error)) {
	// The error returned here is what's in the stop channel but at the very end, a worker might return an error
	// even though the walk is already done. This is why we cannot rely on the return value of the walk.
	filepath.Walk(".", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		select {
		case err := <- stop:
			return err
		default:
			base := filepath.Base(path)
			// skip hidden directories and files
			if path != "." && strings.HasPrefix(base, ".") {
				if info.IsDir() {
					return filepath.SkipDir
				} else {
					return nil
				}
			}
			// skip backup files, avoid recursion
			if strings.HasSuffix(path, "~") || strings.HasPrefix(path, dir) {
				return nil
			}
			// recreate subdirectories
			if info.IsDir() {
				return os.Mkdir(filepath.Join(dir, path), 0755)
			}
			tasks <- args{ path: path, info: info }
			return nil
		}
	})
	close(tasks)
}

// staticWatch counts the values coming out of the done channel. When the count matches the number of jobs started, we
// know that all the tasks have been processed and the results channel is closed.
func staticWatch(jobs int, results chan(error), done chan(bool)) {
	for i := 0; i < jobs; i++ {
		<- done
	}
	close(results)
}

// staticWorker takes arguments off the tasks channel (the file to process) and put results in the results channel (any
// errors encountered); when they're done they send true on the done channel.
func staticWorker(dir string, tasks chan(args), results chan(error), done chan(bool)) {
	task, ok := <- tasks
	for ok {
		results <- staticFile(dir, task.path, task.info)
		task, ok = <- tasks
	}
	done <- true
}

// staticProgressIndicator watches the results channel and does a countdown. If the result channel reports an error,
// that is put into the stop channel so that staticWalk stops adding to the tasks channel.
func staticProgressIndicator(results chan(error), stop chan(error), quiet bool) (int, error) {
	n := 0
	t := time.Now()
	var err error
	for result := range results {
		if result != nil {
			err := result
			// this stops the walker from adding more tasks
			stop <- err
		} else {
			n++
			if !quiet && n % 13 == 0 {
				if time.Since(t) > time.Second {
					fmt.Printf("\r%d", n)
					t = time.Now()
				}
			}
		}
	}
	return n, err
}

// staticFile is used to walk the file trees and do the right thing for the destination directory: create
// subdirectories, link files, render HTML files.
func staticFile(dir, path string, info fs.FileInfo) error {
	// render pages
	if strings.HasSuffix(path, ".md") {
		p, err := staticPage(path, dir)
		if err != nil {
			return err
		}
		return staticFeed(path, dir, p, info.ModTime())
	}
	// remaining files are linked unless this is a template
	if slices.Contains(templateFiles, filepath.Base(path)) {
		return nil
	}
	return os.Link(path, filepath.Join(dir, path))
}

// staticPage takes the filename of a page (ending in ".md") and generates a static HTML page.
func staticPage(path, dir string) (*Page, error) {
	name := strings.TrimSuffix(path, ".md")
	p, err := loadPage(filepath.ToSlash(name))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot load %s: %s\n", name, err)
		return nil, err
	}
	p.handleTitle(true)
	// instead of p.renderHtml() we do it all ourselves, appending ".html" to all the local links
	parser, hashtags := wikiParser()
	doc := markdown.Parse(p.Body, parser)
	ast.WalkFunc(doc, staticLinks)
	opts := html.RendererOptions{
		Flags: html.CommonFlags,
	}
	renderer := html.NewRenderer(opts)
	maybeUnsafeHTML := markdown.Render(doc, renderer)
	p.Name = nameEscape(p.Name)
	p.Html = unsafeBytes(maybeUnsafeHTML)
	p.Language = language(p.plainText())
	p.Hashtags = *hashtags
	return p, write(p, filepath.Join(dir, name+".html"), "", "static.html")
}

// staticFeed writes a .rss file for a page, but only if it's an index page or a page that might be used as a hashtag
func staticFeed(path, dir string, p *Page, ti time.Time) error {
	// render feed, maybe
	name := strings.TrimSuffix(path, ".md")
	base := filepath.Base(name)
	_, ok := index.token["#"+strings.ToLower(base)]
	if base == "index" || ok {
		f := feed(p, ti)
		if len(f.Items) > 0 {
			return write(f, filepath.Join(dir, name + ".rss"), `<?xml version="1.0" encoding="UTF-8"?>`, "feed.html" )
		}
	}
	return nil
}

// staticLinks checks a node and if it is a link to a local page, it appends ".html" to the link destination.
func staticLinks(node ast.Node, entering bool) ast.WalkStatus {
	if entering {
		switch v := node.(type) {
		case *ast.Link:
			// not an absolute URL, not a full URL, not a mailto: URI
			if !bytes.HasPrefix(v.Destination, []byte("/")) &&
				!bytes.Contains(v.Destination, []byte("://")) &&
				!bytes.HasPrefix(v.Destination, []byte("mailto:")) {
				// pointing to a page file (instead of an image file, for example).
				fn, err := url.PathUnescape(string(v.Destination))
				if err != nil {
					return ast.GoToNext
				}
				_, err = os.Stat(fn + ".md")
				if err != nil {
					return ast.GoToNext
				}
				v.Destination = append(v.Destination, []byte(".html")...)
			}
		}
	}
	return ast.GoToNext
}

// write a page or feed with an appropriate template to a specific destination, overwriting it.
func write(data any, destination, prefix, templateFile string) error {
	_, err := os.Stat(destination)
	if err == nil {
		fmt.Fprintf(os.Stderr, "%s already exists\n", destination)
		return nil
	}
	dst, err := os.Create(destination)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot create %s: %s\n", destination, err)
		return err
	}
	_, err = dst.Write([]byte(prefix))
	if err != nil {
		return err
	}
	templates.RLock()
	defer templates.RUnlock()
	err = templates.template[templateFile].Execute(dst, data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot execute %s template for %s: %s\n", templateFile, destination, err)
		return err
	}
	return nil
}
