package main

import (
	"github.com/fsnotify/fsnotify"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"
)

// templateFiles are the various HTML template files used. These files must exist in the root directory for Oddmu to be
// able to generate HTML output. This always requires a template.
var templateFiles = []string{"edit.html", "add.html", "view.html",
	"diff.html", "search.html", "static.html", "upload.html", "feed.html"}

// templates are the parsed HTML templates used. See renderTemplate and loadTemplates. Subdirectories may contain their
// own templates which override the templates in the root directory.
var templates map[string]*template.Template

// templateWatcher is the watcher instance for the wiki. Every directory is added to this watcher.
var templateWatcher *fsnotify.Watcher

// loadTemplates loads the templates. These aren't always required. If the templates are required and cannot be loaded,
// this a fatal error and the program exits. Also start a watcher for these templates so that they are reloaded when
// they are changed.
func loadTemplates() {
	if templates != nil {
		return
	}
	// create a watcher for the directory containing the templates (and never close it)
	w, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println("Creating a watcher for the templates:", err)
	}
	go watch(w)
	templateWatcher = w
	w.Add(".")
	// walk the directory, load templates and add directories
	templates = make(map[string]*template.Template)
	filepath.Walk(".", loadTemplate)
}

// loadTemplate is used to walk the directory. It loads all the template files it finds, including the ones in
// subdirectories.
func loadTemplate(path string, info fs.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if info.IsDir() &&
		!strings.HasPrefix(filepath.Base(path), ".") {
		templateWatcher.Add(path)
	} else if strings.HasSuffix(path, ".html") &&
		slices.Contains(templateFiles, filepath.Base(path)) {
		t, err := template.ParseFiles(path)
		if err != nil {
			log.Println("Cannot parse template:", path, err)
			// ignore error
		} else {
			// log.Println("Parse template:", path)
			templates[path] = t
		}
	}
	return nil
}

// Todo holds a map and a mutex. The map contains the template names that have been requested and the exact time at
// which they have been requested. Adding the same file multiple times, such as when the watch function sees multiple
// Write events for the same template, the time keeps getting updated so that when the go routine runs to reload the
// templates, it only reloads the templates that haven't been updated in the last second. The go routine is what forces
// us to use the RWMutex for the map. The mutex in the structure also ensures that the templates map doesn't get messed
// up.
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
	var todo Todo
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
			// log.Println(e)
			if strings.HasSuffix(e.Name, ".html") &&
				strings.HasPrefix(e.Name, "./") &&
				slices.Contains(templateFiles, filepath.Base(e.Name)) &&
				e.Op.Has(fsnotify.Write) {
				todo.Lock()
				todo.files[e.Name] = time.Now()
				todo.Unlock()
				timer := time.NewTimer(time.Second)
				go func() {
					<-timer.C
					todo.Lock()
					defer todo.Unlock()
					for f, t := range todo.files {
						if t.Add(time.Second).Before(time.Now().Add(time.Nanosecond)) {
							delete(todo.files, f)
							t, err := template.ParseFiles(f)
							if err != nil {
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
	t := templates[tmpl+".html"]
	if t == nil {
		log.Println("Template not found:", tmpl)
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}
	err := t.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// makeDir makes a new subdirectory for pages and adds it to templateWatcher.
func makeDir(d string) error {
	err := os.MkdirAll(d, 0755)
	if err != nil {
		log.Printf("Creating directory %s failed: %s", d, err)
		return err
	}
	templateWatcher.Add(d)
	return nil
}
