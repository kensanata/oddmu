package main

import (
 	"github.com/fsnotify/fsnotify"
	"io/fs"
 	"log"
	"os"
	"path/filepath"
 	"slices"
 	"strings"
 	"sync"
 	"time"
)

// Watches holds a map and a mutex. The map contains the template names that have been requested and the exact time at
// which they have been requested. Adding the same file multiple times, such as when the watch function sees multiple
// Write events for the same file, the time keeps getting updated so that when the go routine runs, it only acts on
// files that haven't been updated in the last second. The go routine is what forces us to use the RWMutex for the map.
type Watches struct {
 	sync.RWMutex
	files map[string]time.Time
	watcher *fsnotify.Watcher
}

var watches Watches

func init() {
	watches.files = make(map[string]time.Time)
}

// install initializes watches and installs watchers for all directories and subdirectories.
func (w *Watches) install() (int, error) {
	// create a watcher for the root directory and never close it
	var err error
	w.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Println("Creating a watcher for file changes:", err)
		return 0, err
	}
	go w.watch()
	err = filepath.Walk(".", w.add)
	if err != nil {
		return 0, err
	}
	return len(w.watcher.WatchList()), nil
}

// add installs a watch for every directory that isn't hidden.
func (w *Watches) add(path string, info fs.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if info.IsDir() {
		if path != "." && strings.HasPrefix(filepath.Base(path), ".") {
			return filepath.SkipDir
		}
		err := w.watcher.Add(path)
		if err != nil {
			log.Println("Cannot add watch:", path)
			return err
		}
		// log.Println("Watching", path)
	}
	return nil
}

// watch reloads templates that have changed and reindexes fils that have changed. Since there can be multiple writes to
// a file, there's a 1s delay before a file is actually handled. The reason is that writing a file can cause multiple
// Write events and we don't want to keep reloading the template while it is being written. Instead, each Write event
// adds an entry to the files map, or updates the file's time, and starts a go routine. Example: If a file gets three
// consecutive Write events, the first two go routine invocations won't do anything, since the time kept getting
// updated. Only the last invocation will act upon the event.
func (w *Watches) watch() {
	for {
		select {
		// Read from Errors.
		case err, ok := <-w.watcher.Errors:
			if !ok { // Channel was closed (i.e. Watcher.Close() was called).
				return
			}
			log.Println("Watcher:", err)
		// Read from Events.
		case e, ok := <-w.watcher.Events:
			if !ok { // Channel was closed (i.e. Watcher.Close() was called).
				return
			}
			w.watchHandle(e)
		}
	}
}

// watchHandle is called for every fsnotify.Event. It handles template updates, page updates (both on a 1s timer), and
// the addition of directories (immediately).
func (w *Watches) watchHandle(e fsnotify.Event) {
	// log.Println(e)
	if e.Op.Has(fsnotify.Write) &&
		(strings.HasSuffix(e.Name, ".html") &&
			slices.Contains(templateFiles, filepath.Base(e.Name)) ||
			strings.HasSuffix(e.Name, ".md")) {
		w.Lock()
		w.files[e.Name] = time.Now()
		w.Unlock()
		timer := time.NewTimer(time.Second)
		go func() {
			<-timer.C
			w.watchTimer()
		}()
	} else if e.Op.Has(fsnotify.Create) {
		w.watchDo(e.Name)
	}
}

// watchTimer checks if any files in the list are templates that need reloading or pages that need reindexing.
func (w *Watches) watchTimer() {
	w.Lock()
	defer w.Unlock()
	for path, t := range w.files {
		if t.Add(time.Second).Before(time.Now().Add(time.Nanosecond)) {
			delete(w.files, path)
			w.watchDo(path)
		}
	}
}

// Do the right thing right now. For Create events such as directories being created or files being moved into a watched
// directory, this is the right thing to do. When a file is being written to, watchHandle will have started timers and
// all that.
func (w *Watches) watchDo(path string) {
	if strings.HasSuffix(path, ".html") {
		updateTemplate(path)
	} else if strings.HasSuffix(path, ".md") {
		p, err := loadPage(path[:len(path)-3]) // page name without ".md"
		if err != nil {
			log.Println("Cannot load page", path)
		} else {
			log.Println("Update index for", path)
			p.updateIndex()
		}
	} else if !slices.Contains(w.watcher.WatchList(), path) {
		fi, err := os.Stat(path)
		if err != nil {
			log.Printf("Cannot stat %s: %s", path, err)
			return
		}
		if fi.IsDir() {
			log.Println("Add watch for", path)
			w.watcher.Add(path)
		}
	}
}

// ignore is called at the end of functions that triggered additions to watches.files. These functions know that they
// handled the file so there's no need to handle the file again via watch. We have 1s before watchTimer is going to get
// called. Therefore, after 10ms, remove the file from the todo list.
func (w *Watches) ignore(path string) {
	timer := time.NewTimer(10*time.Millisecond)
	go func() {
		<-timer.C
		w.Lock()
		defer w.Unlock()
		delete(watches.files, path)
	}()
}
