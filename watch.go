package main

import (
 	"github.com/fsnotify/fsnotify"
	"io/fs"
 	"log"
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
	w.files = make(map[string]time.Time)
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
					w.Lock()
					defer w.Unlock()
					for f, t := range w.files {
						if t.Add(time.Second).Before(time.Now().Add(time.Nanosecond)) {
							delete(w.files, f)
							if strings.HasSuffix(f, ".html") {
								updateTemplate(f)
								log.Println("Watched updated template", f)
							} else if strings.HasSuffix(f, ".md") {
								p, err := loadPage(f[:len(f)-3]) // page name without ".md"
								if err != nil {
									log.Println("Cannot load page", f)
								} else {
									p.updateIndex()
									log.Println("Watched updated index for", f)
								}
							}
						}
					}
				}()
			}
		}
	}
}

// ignore is called at the end of functions that triggered additions to watches.files. These functions know that they
// handled the file so there's no need to handle the file again via watch.
func (w *Watches) ignore(path string) {
	w.Lock()
	defer w.Unlock()
	delete(watches.files, path)
}
