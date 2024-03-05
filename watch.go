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

// watchStore controls access to the maps used by the filesystem watches. Make sure to lock and unlock as appropriate.
// The maps are used to control a sort of queue for files that need reloading (if a template) or reindexing (if a page).
// File system notifications add files to the queue in order to handle changes made without Oddmu, while Oddmu is
// running.
type watchStore struct {
	sync.RWMutex

	// files contains the filenames that have been queued for reloading (if a template) or reindexing (if a page)
	// and the exact time at which they have been added. When the same file is added multiple times, such as when
	// the watchStore function sees multiple Write events for the same file, the time keeps getting updated so that
	// when the watchTimer runs, it only acts on files that haven't been updated in the last second.
	files map[string]time.Time

	// ignores contains the files that some code intends to change, knowing that subsequent writes events would
	// result in file system notifications that would end up adding the filenames to the queue for reloading (if a
	// template) or reindexing (if a page). When Oddmu is making the changes, it can ignore the corresponding
	// notifications by the file system. Those notifications are consequences of Oddmu doing its job. In other
	// words, Oddmu does not rely on file system notification even it is Oddmu doing the changes. This avoids a 1s
	// when changing templates, for example.
	ignores map[string]time.Time

	// watcher is the pointer to the actual watcher doing the file system watching. It watches a set of paths.
	// Whenever Oddmu creates a new subdirectory, it adds the path for this subdirectory to the watcher.
	watcher *fsnotify.Watcher
}

var watches watchStore

func init() {
	watches.ignores = make(map[string]time.Time)
	watches.files = make(map[string]time.Time)
}

// install initializes watches and installs watchers for all directories and subdirectories.
func (w *watchStore) install() (int, error) {
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

// add installs a watch for every directory that isn't hidden. Note that the root directory (".") is not skipped.
func (w *watchStore) add(path string, info fs.FileInfo, err error) error {
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
func (w *watchStore) watch() {
	for {
		select {
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			log.Println("Watcher:", err)
		case e, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			w.watchHandle(e)
		}
	}
}

// watchHandle is called for every fsnotify.Event. It handles template updates, page updates (both on a 1s timer), and
// the creation of pages and directories (immediately). Files and directories starting with a dot are skipped.
// Incidentally, this also prevents rsync updates from generating activity ("stat ./.index.md.tTfPFg: no such file or
// directory"). Note the painful details: If moving a file into a watched directory, a Create event is received. If a
// new file is created in a watched directory, a Create event and one or more Write events is received.
func (w *watchStore) watchHandle(e fsnotify.Event) {
	path := strings.TrimPrefix(e.Name, "./")
	if strings.HasPrefix(filepath.Base(path), ".") {
		return
	}
	// log.Println(e)
	w.Lock()
	defer w.Unlock()
	if e.Op.Has(fsnotify.Create|fsnotify.Write) &&
		(strings.HasSuffix(path, ".html") &&
			slices.Contains(templateFiles, filepath.Base(path)) ||
			strings.HasSuffix(path, ".md")) {
		w.files[path] = time.Now()
		timer := time.NewTimer(time.Second)
		go func() {
			<-timer.C
			w.Lock()
			defer w.Unlock()
			w.watchTimer(path)
		}()
	} else if e.Op.Has(fsnotify.Rename | fsnotify.Remove) {
		w.watchDoRemove(path)
	} else if e.Op.Has(fsnotify.Create) &&
		!slices.Contains(w.watcher.WatchList(), path) {
		fi, err := os.Stat(path)
		if err != nil {
			log.Println(err)
		} else if fi.IsDir() {
			log.Println("Add watch for", path)
			w.watcher.Add(path)
		}
	}
}

// watchTimer checks if the file hasn't been updated in 1s and if so, it calls watchDoUpdate. If another write has
// updated the file, do nothing because another watchTimer will run at the appropriate time and check again.
func (w *watchStore) watchTimer(path string) {
	t, ok := w.files[path]
	if ok && t.Add(time.Second).Before(time.Now().Add(time.Nanosecond)) {
		delete(w.files, path)
		w.watchDoUpdate(path)
	}
}

// Do the right thing right now. For Create events such as directories being created or files being moved into a watched
// directory, this is the right thing to do. When a file is being written to, watchHandle will have started a timer and
// will call this function after 1s of no more writes. If, however, the path is in the ignores map, do nothing.
func (w *watchStore) watchDoUpdate(path string) {
	_, ignored := w.ignores[path]
	if ignored {
		return
	} else if strings.HasSuffix(path, ".html") {
		updateTemplate(path)
	} else if strings.HasSuffix(path, ".md") {
		p, err := loadPage(path[:len(path)-3]) // page name without ".md"
		if err != nil {
			log.Println("Cannot load page", path)
		} else {
			log.Println("Update index for", path)
			index.update(p)
		}
	} else if !slices.Contains(w.watcher.WatchList(), path) {
		fi, err := os.Stat(path)
		if err != nil {
			log.Println(err)
			return
		}
		if fi.IsDir() {
			log.Println("Add watch for", path)
			w.watcher.Add(path)
		}
	}
}

// watchDoRemove removes files from the index or discards templates. If the path in question is in the ignores map, do
// nothing.
func (w *watchStore) watchDoRemove(path string) {
	_, ignored := w.ignores[path]
	if ignored {
		return
	} else if strings.HasSuffix(path, ".html") {
		removeTemplate(path)
	} else if strings.HasSuffix(path, ".md") {
		_, err := os.Stat(path)
		if err == nil {
			log.Println("Cannot remove existing page from the index", path)
		} else {
			log.Println("Deindex", path)
			index.deletePageName(path[:len(path)-3]) // page name without ".md"
		}
	}
}

// ignore is before code that is known suspected save files and trigger watchHandle eventhough the code already handles
// this. This is achieved by adding the path to the ignores map for 1s.
func (w *watchStore) ignore(path string) {
	w.Lock()
	defer w.Unlock()
	w.ignores[path] = time.Now()
	timer := time.NewTimer(time.Second)
	go func() {
		<-timer.C
		w.Lock()
		defer w.Unlock()
		t := w.ignores[path]
		if t.Add(time.Second).Before(time.Now().Add(time.Nanosecond)) {
			delete(w.ignores, path)
		}
	}()
}
