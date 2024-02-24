// Oddmu is a wiki web server and a static site generator.
//
// The types exported are the ones needed to write the templates.
package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/subcommands"
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

// makeHandler returns a handler that uses the URL path without the first path element as its argument, e.g. if the URL
// path is /edit/foo/bar, the editHandler is called with "foo/bar" as its argument. This uses the second group from the
// validPath regular expression. The boolean argument indicates whether the following path is required. When false, a
// URL like /upload/ is OK. The argument can also be provided using a form parameter, i.e. call /edit/?id=foo/bar.
func makeHandler(fn func(http.ResponseWriter, *http.Request, string), required bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// no hidden files or directories
		for _, segment := range strings.Split(r.URL.Path, "/") {
			if strings.HasPrefix(segment, ".") {
				http.Error(w, "can neither confirm nor deny the existence of this resource", http.StatusForbidden)
				return
			}
		}
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
		// no hidden files or directories
		for _, segment := range strings.Split(id, "/") {
			if strings.HasPrefix(segment, ".") {
				http.Error(w, "can neither confirm nor deny the existence of this resource", http.StatusForbidden)
				return
			}
		}
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

// When stdin is a socket, getListener returns a listener that listens
// on the socket passed as stdin.  This allows systemd-style socket
// activation.
// Otherwise, getListener returns a net.Listener listening on the address from
// ODDMU_ADDRESS and the port from ODDMU_PORT.
// ODDMU_ADDRESS may be either an IPV4 address or an IPv6 address.
// If ODDMU_ADDRESS is unspecified, then the
// listener listens on all available unicast addresses, both IPv4 and IPv6.
func getListener() (net.Listener, error) {
	address := os.Getenv("ODDMU_ADDRESS")
	port := getPort()

	stat, err := os.Stdin.Stat()
	if stat == nil {
		return nil, err
	}
	if stat.Mode().Type() == fs.ModeSocket {
		// Listening socket passed on stdin, through systemd socket
		// activation or similar:
		log.Println("Serving a wiki on a listening socket passed by systemd.")
		return net.FileListener(os.Stdin)
	}
	if strings.ContainsRune(address, ':') {
		address = fmt.Sprintf("[%s]:%s", address, port)
	} else {
		address = fmt.Sprintf("%s:%s", address, port)
	}
	log.Printf("Serving a wiki at address %s", address)
	return net.Listen("tcp", address)
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

// scheduleInstallWatcher calls watches.install and prints some messages before and after. For testing, call watch.init
// directly and skip the messages.
func scheduleInstallWatcher() {
	log.Print("Installing watcher")
	n, err := watches.install()
	if err == nil {
		if n == 1 {
			log.Println("Installed watchers for one directory")
		} else {
			log.Printf("Installed watchers for %d directories", n)
		}
	} else {
		log.Printf("Installing watcher failed: %s", err)
	}
}

// serve starts the web server using [http.Serve]. The listener is determined via [getListener]. The various handlers
// are created using [makeHandler] if their path starts with an action segment. For example, the URL path "/view/index"
// is understood to contain the "view" action and so [viewHandler] is called with the argument "index". The one handler
// that doesn't need this is [rootHandler].
//
// The handlers often come in pairs. One handler to show the user interface and one handler to make the change:
//   - [editHandler] shows the edit form and [saveHandler] saves changes to a page
//   - [addHandler] shows the add form and [appendHandler] appends the addition to a page
//   - [uploadHandler] shows the upload form and [dropHandler] saves the uploaded files
//
// Some handlers only do something and the links or forms to call them is expected to be part of the view template:
//   - [archiveHandler] zips up the current directory
//   - [diffHandler] shows the changes made in the last 60min to a page
//   - [searchHandler] shows search results
//
// At the same time as the server starts up, pages are indexed via [scheduleLoadIndex], languages are loaded via
// [scheduleLoadLanguages] and the current directory and its subdirectories is watched for changes using watchers
// installed via [scheduleInstallWatcher].
func serve() {
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/archive/", makeHandler(archiveHandler, true))
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
	go scheduleInstallWatcher()
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
	subcommands.Register(&missingCmd{}, "")
	subcommands.Register(&notifyCmd{}, "")
	subcommands.Register(&replaceCmd{}, "")
	subcommands.Register(&searchCmd{}, "")
	subcommands.Register(&staticCmd{}, "")
	subcommands.Register(&versionCmd{}, "")

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}

// main runs [serve] if called without arguments and it runs [commands] if called with arguments.
// The first argument is the subcommand.
func main() {
	if len(os.Args) == 1 {
		serve()
	} else {
		commands()
	}
}
