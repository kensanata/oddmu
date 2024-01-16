package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/google/subcommands"
	"html/template"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
)

// validPath is a regular expression where the second group matches a page, so when the editHandler is called, a URL
// path of "/edit/foo" results in the editHandler being called with title "foo". The regular expression doesn't define
// the handlers (this happens in the main function).
var validPath = regexp.MustCompile("^/([^/]+)/(.*)$")

// titleRegexp is a regular expression matching a level 1 header line in a Markdown document. The first group matches
// the actual text and is used to provide an title for pages. If no title exists in the document, the page name is used
// instead.
var titleRegexp = regexp.MustCompile("(?m)^#\\s*(.*)\n+")

// renderTemplate is the helper that is used render the templates with data. If the templates cannot be found, that's
// fatal.
func renderTemplate(w http.ResponseWriter, tmpl string, data any) {
	templates := loadTemplates()
	err := templates.ExecuteTemplate(w, tmpl+".html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// makeHandler returns a handler that uses the URL path without the first path element as its argument, e.g. if the URL
// path is /edit/foo/bar, the editHandler is called with "foo/bar" as its argument. This uses the second group from the
// validPath regular expression. The boolean argument indicates whether the following path is required. When false, a
// URL like /upload/ is OK. The argument can also be provided using a form parameter, i.e. call /edit/?id=foo/bar.
func makeHandler(fn func(http.ResponseWriter, *http.Request, string), required bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m != nil && (!required || len(m[2]) > 0) {
			fn(w, r, m[2])
			return
		}
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Cannot parse form", 400)
			return
		}
		id := r.Form.Get("id")
		if m != nil {
			fn(w, r, id)
			return
		}
		http.NotFound(w, r)
	}
}

// getPort returns the environment variable ODDMU_PORT or the default port, "8080".
func getPort() string {
	port := os.Getenv("ODDMU_PORT")
	if port == "" {
		port = "8080"
	}
	return port
}

// getListener returns a net.Listener listening on the address from
// ODDMU_ADDRESS and the port from ODDMU_PORT.
// ODDMU_ADDRESS may be either an IPV4 address, an IPv6 address, or the
// path to a Unix-domain socket.  In the latter case, the value of ODDMU_PORT
// is ignored, because it is not applicable.
// If ODDMU_ADDRESS is unspecified, then the listener listens on all
// available unicast addresses, both IPv4 and IPv6.
// When ODDMU_ADDRESS begins with a / it is taken to be the path of a
// Unix domain socket.
func getListener() (net.Listener, error) {
	family := "tcp"
	address := os.Getenv("ODDMU_ADDRESS")
	port := getPort()
	if strings.ContainsRune(address, '/') {
		family = "unix"
		// Remove stale Unix-domain socket.  ENOENT is ignored, and often
		// expected.
		err := os.Remove(address)
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
	} else if strings.ContainsRune(address, ':') {
		address = fmt.Sprintf("[%s]:%s", address, port)
	} else {
		address = fmt.Sprintf("%s:%s", address, port)
	}
	log.Printf("Serving a wiki at address %s", address)
	return net.Listen(family, address)
}

// scheduleLoadIndex calls index.load and prints some messages before and after. For testing, call index.load directly
// and skip the messages.
func scheduleLoadIndex() {
	log.Print("Indexing pages")
	n, err := index.load()
	if err == nil {
		log.Printf("Indexed %d pages", n)
	} else {
		log.Printf("Indexing failed: %s", err)
	}
}

// scheduleLoadLanguages calls loadLanguages and prints some messages before and after. For testing, call loadLanguages
// directly and skip the messages.
func scheduleLoadLanguages() {
	log.Print("Loading languages")
	n := loadLanguages()
	log.Printf("Loaded %d languages", n)
}

// loadTemplates loads the templates. These aren't always required. If the templates are required and cannot be loaded,
// this a fatal error and the program exits.
func loadTemplates() *template.Template {
	templates, err := template.ParseFiles("edit.html", "add.html", "view.html",
		"diff.html", "search.html", "static.html", "upload.html", "feed.html")
	if err != nil {
		log.Println("Templates:", err)
		os.Exit(1)
	}
	return templates
}

func serve() {
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/view/", makeHandler(viewHandler, true))
	http.HandleFunc("/diff/", makeHandler(diffHandler, true))
	http.HandleFunc("/edit/", makeHandler(editHandler, true))
	http.HandleFunc("/save/", makeHandler(saveHandler, true))
	http.HandleFunc("/add/", makeHandler(addHandler, true))
	http.HandleFunc("/append/", makeHandler(appendHandler, true))
	http.HandleFunc("/upload/", makeHandler(uploadHandler, false))
	http.HandleFunc("/drop/", makeHandler(dropHandler, false))
	http.HandleFunc("/search/", makeHandler(searchHandler, false))
	go scheduleLoadIndex()
	go scheduleLoadLanguages()
	initAccounts()
	listener, err := getListener()
	if listener == nil {
		log.Println(err)
	} else {
		err := http.Serve(listener, nil)
		if err != nil {
			log.Println(err)
		}
	}
}

// commands does the command line parsing in case Oddmu is called with some arguments. Without any arguments, the wiki
// server is started. At this point we already know that there is at least one subcommand.
func commands() {
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(&htmlCmd{}, "")
	subcommands.Register(&listCmd{}, "")
	subcommands.Register(&staticCmd{}, "")
	subcommands.Register(&searchCmd{}, "")
	subcommands.Register(&replaceCmd{}, "")
	subcommands.Register(&missingCmd{}, "")
	subcommands.Register(&notifyCmd{}, "")

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
