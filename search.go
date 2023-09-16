package main

import (
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"unicode"
	"unicode/utf8"
)

// Search is a struct containing the result of a search. Query is the
// query string and Items is the array of pages with the result.
// Currently there is no pagination of results! When a page is part of
// a search result, Body and Html are simple extracts.
type Search struct {
	Query   string
	Items   []*Page
	Previous int
	Page    int
	Next    int
	More    bool
	Results bool
}

func sortItems(a, b *Page) int {
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

// load the pages named.
func load(names []string) []*Page {
	items := make([]*Page, len(names))
	for i, name := range names {
		p, err := loadPage(name)
		if err != nil {
			fmt.Printf("Error loading %s\n", name)
		} else {
			items[i] = p
		}
	}
	return items
}

// search returns a sorted []Page where each page contains an extract
// of the actual Page.Body in its Page.Html. Page size is 20. The
// boolean return value indicates whether there are more results.
func search(q string, page int) ([]*Page, bool) {
	if len(q) == 0 {
		return make([]*Page, 0), false
	}
	names := searchDocuments(q)
	items := load(names)
	for _, p := range items {
		p.score(q)
	}
	slices.SortFunc(items, sortItems)
	from := 20*(page-1)
	if from > len(names) {
		return make([]*Page, 0), false
	}
	to := from + 20
	if to > len(names) {
		to = len(names)
	}
	items = items[from:to]
	for _, p := range items {
		p.summarize(q)
	}
	return items, to < len(names)
}

// searchHandler presents a search result. It uses the query string in
// the form parameter "q" and the template "search.html". For each
// page found, the HTML is just an extract of the actual body.
func searchHandler(w http.ResponseWriter, r *http.Request) {
	q := r.FormValue("q")
	page, err := strconv.Atoi(r.FormValue("page"))
	if err != nil {
		page = 1
	}
	items, more := search(q, page)
	s := &Search{Query: q, Items: items, Previous: page-1, Page: page, Next: page+1, Results: len(items) > 0, More: more}
	renderTemplate(w, "search", s)
}
