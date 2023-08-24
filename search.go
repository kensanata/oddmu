package main

import (
	trigram "github.com/dgryski/go-trigram"
	"path/filepath"
	"strings"
	"io/fs"
	"fmt"
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
	fmt.Printf("Indexing %s\n", name)
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

func updateIndex(p *Page) {
	var id trigram.DocID
	for docId, name := range documents {
		if name == p.Name {
			id = docId
			break
		}
	}
	s := string(p.Body)
	if id == 0 {
		id = index.Add(s)
		documents[id] = p.Name
	} else {
		index.Delete(s, id)
		index.Insert(s, id)
	}
}
