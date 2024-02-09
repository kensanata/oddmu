// Read Artem Krylysov's blog post on full text search as an
// introduction.
// https://artem.krylysov.com/blog/2020/07/28/lets-build-a-full-text-search-engine/

package main

import (
	"golang.org/x/exp/constraints"
	"io/fs"
	"log"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

type docid uint

// Index contains the two maps used for search. Make sure to lock and
// unlock as appropriate.
type Index struct {
	sync.RWMutex

	// next_id is the number of the next document added to the index
	next_id docid

	// index is an inverted index mapping tokens to document ids.
	token map[string][]docid

	// documents is a map, mapping document ids to page names.
	documents map[docid]string

	// titles is a map, mapping page names to titles.
	titles map[string]string
}

var index Index

func init() {
	index.reset()
}

// reset the index. This assumes that the index is locked. It's useful for tests.
func (idx *Index) reset() {
	idx.next_id = 0
	idx.token = make(map[string][]docid)
	idx.documents = make(map[docid]string)
	idx.titles = make(map[string]string)
}

// addDocument adds the text as a new document. This assumes that the index is locked!
func (idx *Index) addDocument(text []byte) docid {
	id := idx.next_id
	idx.next_id++
	for _, token := range hashtags(text) {
		ids := idx.token[token]
		// Don't add same ID more than once. Checking the last
		// position of the []docid works because the id is
		// always a new one, i.e. the last one, if at all.
		if len(ids) > 0 && ids[len(ids)-1] == id {
			continue
		}
		idx.token[token] = append(ids, id)
	}
	return id
}

// deleteDocument deletes all references to the id. The id can no longer be used. This assumes that the index is locked.
func (idx *Index) deleteDocument(id docid) {
	// Looping through all tokens makes sense if there are few tokens (like hashtags). It doesn't make sense if the
	// number of tokens is large (like for full-text search or a trigram index).
	for token, ids := range idx.token {
		// If the token appears only in this document, remove the whole entry.
		if len(ids) == 1 && ids[0] == id {
			delete(idx.token, token)
			continue
		}
		// Otherwise, remove the token from the index.
		i := sort.Search(len(ids), func(i int) bool { return ids[i] >= id })
		if i != -1 && i < len(ids) && ids[i] == id {
			copy(ids[i:], ids[i+1:])
			idx.token[token] = ids[:len(ids)-1]
			continue
		}
	}
}

// deletePageName determines the document id based on the page name and calls deleteDocument to delete all references.
func (idx *Index) deletePageName(name string) {
	idx.Lock()
	defer idx.Unlock()
	var id docid
	// Reverse lookup! At least it's in memory.
	for key, value := range idx.documents {
		if value == name {
			id = key
			break
		}
	}
	if id == 0 {
		log.Printf("Page %s is not indexed", name)
		return
	}
	delete(idx.documents, id)
	delete(idx.titles, name)
	idx.deleteDocument(id)
}

// remove the page from the index. Do this when deleting a page.
func (idx *Index) remove(p *Page) {
	idx.deletePageName(p.Name)
}

// add reads a file and adds it to the index. This must happen while the idx is locked.
func (idx *Index) add(path string, info fs.FileInfo, err error) error {
	if err != nil {
		return err
	}
	filename := path
	if info.IsDir() ||
		strings.HasPrefix(filepath.Base(filename), ".") ||
		!strings.HasSuffix(filename, ".md") {
		return nil
	}
	name := strings.TrimSuffix(filename, ".md")
	p, err := loadPage(name)
	if err != nil {
		return err
	}
	p.handleTitle(false)

	id := idx.addDocument(p.Body)
	idx.documents[id] = p.Name
	idx.titles[p.Name] = p.Title
	return nil
}

// load loads all the pages and indexes them. This takes a while. It returns the number of pages indexed.
func (idx *Index) load() (int, error) {
	idx.Lock()
	defer idx.Unlock()
	err := filepath.Walk(".", idx.add)
	if err != nil {
		return 0, err
	}
	n := len(idx.documents)
	return n, nil
}

// dump prints the index to the log for debugging.
func (idx *Index) dump() {
	idx.RLock()
	defer idx.RUnlock()
	for token, ids := range idx.token {
		log.Printf("%s: %v", token, ids)
	}
}

// updateIndex updates the index for a single page.
func (p *Page) updateIndex() {
	index.Lock()
	defer index.Unlock()
	var id docid
	// Reverse lookup! At least it's in memory.
	for key, value := range index.documents {
		if value == p.Name {
			id = key
			break
		}
	}
	if id == 0 {
		id = index.addDocument(p.Body)
		index.documents[id] = p.Name
		index.titles[p.Name] = p.Title
	} else {
		index.deleteDocument(id)
		// Do not reuse the old id. We need a new one for indexing to work.
		id = index.addDocument(p.Body)
		// The page name stays the same but the title may have changed.
		index.documents[id] = p.Name
		p.handleTitle(false)
		index.titles[p.Name] = p.Title
	}
}

// search searches the index for a query string and returns page
// names.
func (idx *Index) search(q string) []string {
	idx.RLock()
	defer idx.RUnlock()
	names := make([]string, 0)
	hashtags := hashtags([]byte(q))
	if len(hashtags) > 0 {
		var r []docid
		for _, token := range hashtags {
			if ids, ok := idx.token[token]; ok {
				if r == nil {
					r = ids
				} else {
					r = intersection(r, ids)
				}
			} else {
				// Token doesn't exist therefore abort search.
				return nil
			}
		}
		for _, id := range r {
			names = append(names, idx.documents[id])
		}
	} else {
		for _, name := range idx.documents {
			names = append(names, name)
		}
	}
	return names
}

// intersection returns the set intersection between a and b.
// a and b have to be sorted in ascending order and contain no duplicates.
func intersection[T constraints.Ordered](a []T, b []T) []T {
	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}
	r := make([]T, 0, maxLen)
	var i, j int
	for i < len(a) && j < len(b) {
		if a[i] < b[j] {
			i++
		} else if a[i] > b[j] {
			j++
		} else {
			r = append(r, a[i])
			i++
			j++
		}
	}
	return r
}
