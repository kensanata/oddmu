package main

import (
	"fmt"
	trigram "github.com/dgryski/go-trigram"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
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

// index is a struct containing the trigram index for search. It is
// generated at startup and updated after every page edit.
var index trigram.Index

// documents is a map, mapping document ids of the index to page
// names.
var documents map[trigram.DocID]string

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
	id := index.Add(string(p.Body))
	documents[id] = p.Name
	return nil
}

func loadIndex() error {
	index = make(trigram.Index)
	documents = make(map[trigram.DocID]string)
	err := filepath.Walk(".", indexAdd)
	if err != nil {
		fmt.Println("Indexing failed")
		index = nil
		documents = nil
	}
	return err
}

func (p *Page) updateIndex() {
	var id trigram.DocID
	for docId, name := range documents {
		if name == p.Name {
			id = docId
			break
		}
	}
	if id == 0 {
		id = index.Add(string(p.Body))
		documents[id] = p.Name
	} else {
		o, err := loadPage(p.Name)
		if err == nil {
			index.Delete(string(o.Body), id)
		}
		index.Insert(string(p.Body), id)
	}
}

// search returns a sorted []Page where each page contains an extract
// of the actual Page.Body in its Page.Html.
func search(q string) []Page {
	ids := index.Query(q)
	items := make([]Page, len(ids))
	for i, id := range ids {
		name := documents[id]
		p, err := loadPage(name)
		if err != nil {
			fmt.Printf("Error loading %s\n", name)
		} else {
			p.summarize(q)
			items[i] = *p
		}
	}
	fn := func(a, b Page) int {
		if a.Score < b.Score {
			return 1
		} else if a.Score > b.Score {
			return -1
		} else if a.Title < b.Title {
			return -1
		} else if a.Title > b.Title {
			return 1
		} else {
			return 0
		}
	}
	slices.SortFunc(items, fn)
	return items
}
