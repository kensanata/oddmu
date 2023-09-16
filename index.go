package main

import (
	trigram "github.com/dgryski/go-trigram"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"
)

// Index contains the two maps used for search. Make sure to lock and
// unlock as appropriate.
type Index struct {
	sync.RWMutex

	// index is a struct containing the trigram index for search.
	// It is generated at startup and updated after every page
	// edit. The index is case-insensitive.
	index trigram.Index

	// documents is a map, mapping document ids of the index to
	// page names.
	documents map[trigram.DocID]string

	// names is a map, mapping page names to titles.
	titles map[string]string
}

// idx is the global Index per wiki.
var index Index

// reset resets the Index. This assumes that the index is locked!
func (idx *Index) reset() {
	idx.index = nil
	idx.documents = nil
	idx.titles = nil
}

// add reads a file and adds it to the index. This must happen while
// the idx is locked, which is true when called from loadIndex.
func (idx *Index) add(path string, info fs.FileInfo, err error) error {
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
	p.handleTitle(false)
	id := idx.index.Add(strings.ToLower(string(p.Body)))
	idx.documents[id] = p.Name
	idx.titles[p.Name] = p.Title
	return nil
}

// load loads all the pages and indexes them. This takes a while.
// It returns the number of pages indexed.
func (idx *Index) load() (int, error) {
	idx.Lock()
	defer idx.Unlock()
	idx.index = make(trigram.Index)
	idx.documents = make(map[trigram.DocID]string)
	idx.titles = make(map[string]string)
	err := filepath.Walk(".", idx.add)
	if err != nil {
		idx.reset()
		return 0, err
	}
	n := len(idx.documents)
	return n, nil
}

// updateIndex updates the index for a single page. The old text is
// loaded from the disk and removed from the index first, if it
// exists.
func (p *Page) updateIndex() {
	index.Lock()
	defer index.Unlock()
	var id trigram.DocID
	// This function does not rely on files actually existing, so
	// let's quickly find the document id.
	for docId, name := range index.documents {
		if name == p.Name {
			id = docId
			break
		}
	}
	if id == 0 {
		id = index.index.Add(strings.ToLower(string(p.Body)))
		index.documents[id] = p.Name
	} else {
		o, err := loadPage(p.Name)
		if err == nil {
			index.index.Delete(strings.ToLower(string(o.Body)), id)
			o.handleTitle(false)
			delete(index.titles, o.Title)
		}
		index.index.Insert(strings.ToLower(string(p.Body)), id)
		p.handleTitle(false)
		index.titles[p.Name] = p.Title
	}
}

// searchDocuments searches the index for a string. This requires the
// index to be locked.
func searchDocuments(q string) []string {
	words := strings.Fields(strings.ToLower(q))
	var trigrams []trigram.T
	for _, word := range words {
		trigrams = trigram.Extract(word, trigrams)
	}
	ids := index.index.QueryTrigrams(trigrams)
	names := make([]string, len(ids))
	for i, id := range ids {
		names[i] = index.documents[id]
	}
	return names
}
