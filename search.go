package main

import (
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Search is a struct containing the result of a search. Query is the
// query string and Items is the array of pages with the result.
// Currently there is no pagination of results! When a page is part of
// a search result, Body and Html are simple extracts.
type Search struct {
	Query    string
	Items    []*Page
	Previous int
	Page     int
	Next     int
	Last     int
	More     bool
	Results  bool
}

// sortNames returns a sort function that sorts in three stages: 1.
// whether the query string matches the page title; 2. descending if
// the page titles start with a digit; 3. otherwise ascending.
// Access to the index requires a read lock!
func sortNames(q string) func(a, b string) int {
	return func(a, b string) int {
		// If only one page contains the query string, it
		// takes precedence.
		ia := strings.Contains(index.titles[a], q)
		ib := strings.Contains(index.titles[b], q)
		if ia && !ib {
			return -1
		} else if !ia && ib {
			return 1
		}
		// If both page names start with a number (like an ISO date),
		// sort descending.
		ra, _ := utf8.DecodeRuneInString(a)
		rb, _ := utf8.DecodeRuneInString(b)
		if unicode.IsNumber(ra) && unicode.IsNumber(rb) {
			if a < b {
				return 1
			} else if a > b {
				return -1
			} else {
				return 0
			}
		}
		// Otherwise sort by title, ascending.
		if index.titles[a] < index.titles[b] {
			return -1
		} else if index.titles[a] > index.titles[b] {
			return 1
		} else {
			return 0
		}
	}
}

// load the pages named.
func load(names []string) []*Page {
	items := make([]*Page, len(names))
	for i, name := range names {
		p, err := loadPage(name)
		if err != nil {
			log.Printf("Error loading %s: %s", name, err)
		} else {
			items[i] = p
		}
	}
	return items
}

// itemsPerPage says how many items to print on a page of search
// results.
const itemsPerPage = 20

// search returns a sorted []Page where each page contains an extract
// of the actual Page.Body in its Page.Html. Page size is 20. The
// boolean return value indicates whether there are more results.
func search(q string, page int) ([]*Page, bool, int) {
	if len(q) == 0 {
		return make([]*Page, 0), false, 0
	}
	names := index.search(q)
	slices.SortFunc(names, sortNames(q))
	from := itemsPerPage * (page - 1)
	if from > len(names) {
		return make([]*Page, 0), false, 0
	}
	to := from + itemsPerPage
	if to > len(names) {
		to = len(names)
	}
	items := load(names[from:to])
	for _, p := range items {
		p.score(q)
		p.summarize(q)
	}
	return items, to < len(names), len(names)/itemsPerPage + 1
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
	items, more, last := search(q, page)
	s := &Search{Query: q, Items: items, Previous: page - 1, Page: page, Next: page + 1, Last: last,
		Results: len(items) > 0, More: more}
	renderTemplate(w, "search", s)
}
