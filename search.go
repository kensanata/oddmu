package main

import (
	"fmt"
	trigram "github.com/dgryski/go-trigram"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
)

// Search is a struct containing the result of a search. Query is the
// query string and Items is the array of pages with the result.
// Currently there is no pagination of results! When a page is part of
// a search result, Body and Html are simple extracts.
type Search struct {
	Query   string
	Items   []Page
	Results bool
}

// idx contains the two maps used for search. Make sure to lock and
// unlock as appropriate.
var idx = struct {
	sync.RWMutex

	// index is a struct containing the trigram index for search. It is
	// generated at startup and updated after every page edit. The index
	// is case-insensitive.
	index trigram.Index

	// documents is a map, mapping document ids of the index to page
	// names.
	documents map[trigram.DocID]string
}{}

// indexAdd reads a file and adds it to the index. This must happen
// while the idx is locked, which is true when called from loadIndex.
func indexAdd(path string, info fs.FileInfo, err error) error {
	if err != nil {
		return err
	}
	filename := path
	if info.IsDir() || strings.HasPrefix(filename, ".") || !strings.HasSuffix(filename, ".md") {
		return nil
	}
	name := strings.TrimSuffix(filename, ".md")
	p, err := loadPage(name)
	if err != nil {
		return err
	}
	id := idx.index.Add(strings.ToLower(string(p.Body)))
	idx.documents[id] = p.Name
	return nil
}

// loadIndex loads all the pages and indexes them. This takes a while.
// It returns the number of pages indexed.
func loadIndex() (int, error) {
	idx.Lock()
	defer idx.Unlock()
	idx.index = make(trigram.Index)
	idx.documents = make(map[trigram.DocID]string)
	err := filepath.Walk(".", indexAdd)
	if err != nil {
		idx.index = nil
		idx.documents = nil
		return 0, err
	}
	n := len(idx.documents)
	return n, nil
}

// updateIndex updates the index for a single page. The old text is
// loaded from the disk and removed from the index first, if it
// exists.
func (p *Page) updateIndex() {
	idx.Lock()
	defer idx.Unlock()
	var id trigram.DocID
	// This function does not rely on files actually existing, so
	// let's quickly find the document id.
	for docId, name := range idx.documents {
		if name == p.Name {
			id = docId
			break
		}
	}
	if id == 0 {
		id = idx.index.Add(strings.ToLower(string(p.Body)))
		idx.documents[id] = p.Name
	} else {
		o, err := loadPage(p.Name)
		if err == nil {
			idx.index.Delete(strings.ToLower(string(o.Body)), id)
		}
		idx.index.Insert(strings.ToLower(string(p.Body)), id)
	}
}

func sortItems(a, b Page) int {
	// Sort by score
	if a.Score < b.Score {
		return 1
	} else if a.Score > b.Score {
		return -1
	}
	// If the score is the same and both page names start
	// with a number (like an ISO date), sort descending.
	ra, _ := utf8.DecodeRuneInString(a.Title)
	rb, _ := utf8.DecodeRuneInString(b.Title)
	if unicode.IsNumber(ra) && unicode.IsNumber(rb) {
		if a.Title < b.Title {
			return 1
		} else if a.Title > b.Title {
			return -1
		} else {
			return 0
		}
	}
	// Otherwise sort ascending.
	if a.Title < b.Title {
		return -1
	} else if a.Title > b.Title {
		return 1
	} else {
		return 0
	}
}

// loadAndSummarize loads the pages named and summarizes them for the
// query give.
func loadAndSummarize(names []string, q string) []Page {
	// Load and summarize the items.
	items := make([]Page, len(names))
	for i, name := range names {
		p, err := loadPage(name)
		if err != nil {
			fmt.Printf("Error loading %s\n", name)
		} else {
			p.summarize(q)
			items[i] = *p
		}
	}
	return items
}

// search returns a sorted []Page where each page contains an extract
// of the actual Page.Body in its Page.Html.
func search(q string) []Page {
	if len(q) == 0 {
		return make([]Page, 0)
	}
	words := strings.Fields(strings.ToLower(q))
	var trigrams []trigram.T
	for _, word := range words {
		trigrams = trigram.Extract(word, trigrams)
	}
	// Keep the read lock for a short as possible. Make a list of
	// the names we need to load and summarize.
	idx.RLock()
	ids := idx.index.QueryTrigrams(trigrams)
	names := make([]string, len(ids))
	for i, id := range ids {
		names[i] = idx.documents[id]
	}
	idx.RUnlock()
	items := loadAndSummarize(names, q)
	slices.SortFunc(items, sortItems)
	return items
}
