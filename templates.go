package main

import (
	"github.com/fsnotify/fsnotify"
	"html/template"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync"
	"time"
)

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
