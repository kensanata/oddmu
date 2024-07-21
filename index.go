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
	"html/template"
)

type docid uint

// ImageData holds the data used to search for images using the alt-text. Title is the alt-text; Name is the complete
// URL including path (which is important since the image link itself only has the URL relative to the page in which it
// is found; and Html is a copy of the Title with highlighting of a term as applied when searching. This is temporary.
// It depends on the fact that Title is always plain text.
type ImageData struct {
	Title, Name string
	Html template.HTML
}

// indexStore controls access to the maps used for search. Make sure to lock and unlock as appropriate.
type indexStore struct {
	sync.RWMutex

	// next_id is the number of the next document added to the index
	next_id docid

	// index is an inverted index mapping tokens to document ids.
	token map[string][]docid

	// documents is a map, mapping document ids to page names.
	documents map[docid]string

	// titles is a map, mapping page names to titles.
	titles map[string]string

	// images is a map, mapping pages names to alt text to an array of image data.
	images map[string][]ImageData
}

var index indexStore

func init() {
	index.reset()
}

// reset the index. This assumes that the index is locked. It's useful for tests.
func (idx *indexStore) reset() {
	idx.next_id = 0
	idx.token = make(map[string][]docid)
	idx.documents = make(map[docid]string)
	idx.titles = make(map[string]string)
	idx.images = make(map[string][]ImageData)
}

// addDocument adds the text as a new document. This assumes that the index is locked!
func (idx *indexStore) addDocument(text []byte) docid {
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
func (idx *indexStore) deleteDocument(id docid) {
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
// This assumes that the index is unlocked.
func (idx *indexStore) deletePageName(name string) {
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
	if id != 0 {
		idx.deleteDocument(id)
		delete(idx.documents, id)
	}
	delete(idx.titles, name)
	delete(idx.images, name)
}

// remove the page from the index. Do this when deleting a page. This assumes that the index is unlocked.
func (idx *indexStore) remove(p *Page) {
	idx.deletePageName(p.Name)
}

// load loads all the pages and indexes them. This takes a while. It returns the number of pages indexed.
func (idx *indexStore) load() (int, error) {
	idx.Lock()
	defer idx.Unlock()
	err := filepath.Walk(".", idx.walk)
	if err != nil {
		return 0, err
	}
	n := len(idx.documents)
	return n, nil
}

// walk reads a file and adds it to the index. This assumes that the index is locked.
func (idx *indexStore) walk(path string, info fs.FileInfo, err error) error {
	if err != nil {
		return err
	}
	// skip hidden directories and files
	if path != "." && strings.HasPrefix(filepath.Base(path), ".") {
		if info.IsDir() {
			return filepath.SkipDir
		} else {
			return nil
		}
	}
	// skipp all but page files
	if !strings.HasSuffix(path, ".md") {
		return nil
	}
	p, err := loadPage(strings.TrimSuffix(path, ".md"))
	if err != nil {
		return err
	}
	p.handleTitle(false)
	idx.addPage(p)
	return nil
}

// addPage adds a page to the index. This assumes that the index is locked.
func (idx *indexStore) addPage(p *Page) {
	id := idx.addDocument(p.Body)
	idx.documents[id] = p.Name
	p.handleTitle(false)
	idx.titles[p.Name] = p.Title
	idx.images[p.Name] = p.images()
}

// add a page to the index. This assumes that the index is unlocked.
func (idx *indexStore) add(p *Page) {
	idx.Lock()
	defer idx.Unlock()
	idx.addPage(p)
}

// dump prints the index to the log for debugging.
func (idx *indexStore) dump() {
	idx.RLock()
	defer idx.RUnlock()
	for token, ids := range idx.token {
		log.Printf("%s: %v", token, ids)
	}
}

// updateIndex updates the index for a single page.
func (idx *indexStore) update(p *Page) {
	idx.remove(p)
	idx.add(p)
}

// search searches the index for a query string and returns page
// names.
func (idx *indexStore) search(q string) []string {
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
