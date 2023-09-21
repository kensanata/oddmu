package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/subcommands"
	"html/template"
	"net/http"
	"os"
	"regexp"
)

// Templates are parsed at startup.
var templates = template.Must(
	template.ParseFiles("edit.html", "add.html", "view.html",
		"search.html", "upload.html"))

// validPath is a regular expression where the second group matches a
// page, so when the editHandler is called, a URL path of "/edit/foo"
// results in the editHandler being called with title "foo". The
// regular expression doesn't define the handlers (this happens in the
// main function).
var validPath = regexp.MustCompile("^/([^/]+)/(.*)$")

// titleRegexp is a regular expression matching a level 1 header line
// in a Markdown document. The first group matches the actual text and
// is used to provide an title for pages. If no title exists in the
// document, the page name is used instead.
var titleRegexp = regexp.MustCompile("(?m)^#\\s*(.*)\n+")

// renderTemplate is the helper that is used render the templates with
// data.
func renderTemplate(w http.ResponseWriter, tmpl string, data any) {
	err := templates.ExecuteTemplate(w, tmpl+".html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// makeHandler returns a handler that uses the URL path without the
// first path element as its argument, e.g. if the URL path is
// /edit/foo/bar, the editHandler is called with "foo/bar" as its
// argument. This uses the second group from the validPath regular
// expression. The boolean argument indicates whether the following
// path is required. When false, a URL /upload/ is OK.
func makeHandler(fn func(http.ResponseWriter, *http.Request, string), required bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m != nil && (!required || len(m[2]) > 0) {
			fn(w, r, m[2])
		} else {
			http.NotFound(w, r)
		}
	}
}

// getPort returns the environment variable ODDMU_PORT or the default
// port, "8080".
func getPort() string {
	port := os.Getenv("ODDMU_PORT")
	if port == "" {
		port = "8080"
	}
	return port
}

// scheduleLoadIndex calls index.load and prints some messages before
// and after. For testing, call index.load directly and skip the
// messages.
func scheduleLoadIndex() {
	fmt.Print("Indexing pages\n")
	n, err := index.load()
	if err == nil {
		fmt.Printf("Indexed %d pages\n", n)
	} else {
		fmt.Println("Indexing failed")
	}
}

// scheduleLoadLanguages calls loadLanguages and prints some messages before
// and after. For testing, call loadLanguages directly and skip the
// messages.
func scheduleLoadLanguages() {
	fmt.Print("Loading languages\n")
	n := loadLanguages()
	fmt.Printf("Loaded %d languages\n", n)
}

func serve() {
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/view/", makeHandler(viewHandler, true))
	http.HandleFunc("/edit/", makeHandler(editHandler, true))
	http.HandleFunc("/save/", makeHandler(saveHandler, true))
	http.HandleFunc("/add/", makeHandler(addHandler, true))
	http.HandleFunc("/append/", makeHandler(appendHandler, true))
	http.HandleFunc("/upload/", makeHandler(uploadHandler, false))
	http.HandleFunc("/drop/", makeHandler(dropHandler, false))
	http.HandleFunc("/search", searchHandler)
	go scheduleLoadIndex()
	go scheduleLoadLanguages()
	initAccounts()
	port := getPort()
	fmt.Printf("Serving a wiki on port %s\n", port)
	http.ListenAndServe(":"+port, nil)
}

// commands does the command line parsing in case Oddmu is called with
// some arguments. Without any arguments, the wiki server is started.
// At this point we already know that there is at least one
// subcommand.
func commands() {
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(&htmlCmd{}, "")
	subcommands.Register(&searchCmd{}, "")
	subcommands.Register(&replaceCmd{}, "")

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}

func main() {
	if len(os.Args) == 1 {
		serve()
	} else {
		commands()
	}
}
