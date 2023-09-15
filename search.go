package main

import (
	"fmt"
	"net/http"
	"slices"
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
	names := searchDocuments(q)
	items := loadAndSummarize(names, q)
	slices.SortFunc(items, sortItems)
	return items
}

// searchHandler presents a search result. It uses the query string in
// the form parameter "q" and the template "search.html". For each
// page found, the HTML is just an extract of the actual body.
func searchHandler(w http.ResponseWriter, r *http.Request) {
	q := r.FormValue("q")
	items := search(q)
	s := &Search{Query: q, Items: items, Results: len(items) > 0}
	renderTemplate(w, "search", s)
}
