package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/google/subcommands"
	"html/template"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"
)

// validPath is a regular expression where the second group matches a page, so when the editHandler is called, a URL
// path of "/edit/foo" results in the editHandler being called with title "foo". The regular expression doesn't define
// the handlers (this happens in the main function).
var validPath = regexp.MustCompile("^/([^/]+)/(.*)$")

// titleRegexp is a regular expression matching a level 1 header line in a Markdown document. The first group matches
// the actual text and is used to provide an title for pages. If no title exists in the document, the page name is used
// instead.
var titleRegexp = regexp.MustCompile("(?m)^#\\s*(.*)\n+")

// templateFiles are the various HTML template files used.
var templateFiles = []string{"edit.html", "add.html", "view.html",
	"diff.html", "search.html", "static.html", "upload.html", "feed.html"}

// templates are the parsed HTML templates used. See renderTemplate and loadTemplates.
var templates map[string]*template.Template

// loadTemplates loads the templates. These aren't always required. If the templates are required and cannot be loaded,
// this a fatal error and the program exits. Also start a watcher for these templates so that they are reloaded when
// they are changed.
func loadTemplates() {
	if (templates != nil) {
		return;
	}
	templates = make(map[string]*template.Template)
	for _, filename := range templateFiles {
		t, err := template.ParseFiles(filename)
		if err != nil {
			log.Println("Cannot parse:", filename, err)
			os.Exit(1)
		}
		templates[filename] = t
	}
	// create a watcher for the directory containing the templates (and never close it)
	w, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println("Creating a watcher for the templates:", err)
	}
	go watch(w)
	err = w.Add(".")
	if err != nil {
		log.Println("Add root directory to the watcher for the templates:", err)
	}
}

// Todo holds a map and a mutex. The map contains the template names that have been requested and the exact time at
// which they have been requested. Adding the same file multiple times, such as when the watch function sees multiple
// Write events for the same template, the time keeps getting updated so that when the go routine runs to reload the
// templates, it only reloads the templates that haven't been updated in the last second. The go routine is what forces
// us to use the RWMutex for the map.
type Todo struct {
	sync.RWMutex
	files map[string]time.Time
}

// watch reloads templates that have changed. Since there can be multiple writes to a template file, there's a 1s delay
// before a template file is actually reloaded. The reason is that writing a template can cause multiple Write events
// and we don't want to keep reloading the template while it is being written. Instead, each Write event adds an entry
// to the Todo struct, or updates the time, and starts a go routine. If a template gets three consecutive Write events,
// the first two go routine invocations won't do anything, since the time kept getting updated. Only the last invocation
// will reload the template and remove the entry.
func watch(w *fsnotify.Watcher) {
	var todo Todo;
	todo.files = make(map[string]time.Time)
	for {
		select {
		// Read from Errors.
		case err, ok := <-w.Errors:
			if !ok { // Channel was closed (i.e. Watcher.Close() was called).
				return
			}
			log.Println("Watcher:", err)
		// Read from Events.
		case e, ok := <-w.Events:
			if !ok { // Channel was closed (i.e. Watcher.Close() was called).
				return
			}
			if (strings.HasSuffix(e.Name, ".html") &&
				strings.HasPrefix(e.Name, "./") &&
				slices.Contains(templateFiles, e.Name[2:]) &&
				e.Op.Has(fsnotify.Write)) {
				todo.Lock()
				todo.files[e.Name] = time.Now()
				todo.Unlock()
				timer := time.NewTimer(time.Second)
				go func() {
					<- timer.C
					todo.Lock()
					defer todo.Unlock()
					for f, t := range todo.files {
						if (t.Add(time.Second).Before(time.Now().Add(time.Nanosecond))) {
							delete(todo.files, f)
							t, err := template.ParseFiles(f)
							if (err != nil) {
								log.Println("Template:", f, err)
							} else {
								templates[f[2:]] = t
								log.Println("Parsed", f)
							}
						}
					}
				}()
			}
		}
	}
}

// renderTemplate is the helper that is used to render the templates with data. If the templates cannot be found, that's
// fatal.
func renderTemplate(w http.ResponseWriter, tmpl string, data any) {
	loadTemplates()
	err := templates[tmpl+".html"].Execute(w, data)
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
