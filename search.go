package main

import (
	"log"
	"net/http"
	"regexp"
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
	More     bool
	Results  bool
}

// sortNames returns a sort function that sorts in three stages: 1.
// whether the query string matches the page title; 2. descending if
// the page titles start with a digit; 3. otherwise ascending.
// Access to the index requires a read lock!
func sortNames(tokens []string) func(a, b string) int {
	return func(a, b string) int {
		// If only one page contains the query string, it
		// takes precedence.
		ia := false
		ib := false
		for _, token := range tokens {
			if !ia && strings.Contains(index.titles[a], token) {
				ia = true
			}
			if !ib && strings.Contains(index.titles[b], token) {
				ib = true
			}
		}
		if ia && !ib {
			return -1
		} else if !ia && ib {
			return 1
		}
		// Page names starting with a number come first. If
		// both page names start with a number (like an ISO
		// date), sort by page name, descending.
		ra, _ := utf8.DecodeRuneInString(a)
		na := unicode.IsNumber(ra)
		rb, _ := utf8.DecodeRuneInString(b)
		nb := unicode.IsNumber(rb)
		if na && !nb {
			return -1
		} else if !na && nb {
			return 1
		} else if na && nb {
			if a < b {
				return 1
			} else if a > b {
				return -1
			}
		}
		// Otherwise sort by title, ascending.
		if index.titles[a] < index.titles[b] {
			return -1
		} else if index.titles[a] > index.titles[b] {
			return 1
		}
		// Either the titles are equal or the index isn't
		// initialized.
		if a < b {
			return -1
		} else if a > b {
			return 1
		}
		return 0
	}
}

// itemsPerPage says how many items to print on a page of search
// results.
const itemsPerPage = 20

// search returns a sorted []Page where each page contains an extract
// of the actual Page.Body in its Page.Html. Page size is 20. The
// boolean return value indicates whether there are more results.
func search(q string, page int) ([]*Page, bool) {
	if len(q) == 0 {
		return make([]*Page, 0), false
	}
	names := index.search(q) // hashtags or all names
	predicates, terms := predicatesAndTokens(q)
	names = filterNames(names, predicates)
	slices.SortFunc(names, sortNames(terms))
	from := itemsPerPage * (page - 1)
	to := from + itemsPerPage - 1
	items, more := grep(terms, names, from, to)
	for _, p := range items {
		p.score(q)
		p.summarize(q)
	}
	return items, more
}

// filterNames filters the names by all the predicats such as
// "title:foo" or "blog:true".
func filterNames(names, predicates []string) []string {
	if len(predicates) == 0 {
		return names
	}
	// the intersection requires sorted lists
	slices.Sort(names)
	index.RLock()
	defer index.RUnlock()
	for _, predicate := range predicates {
		r := make([]string, 0)
		if strings.HasPrefix(predicate, "title:") {
			token := predicate[6:]
			for _, name := range names {
				if strings.Contains(strings.ToLower(index.titles[name]), token) {
					r = append(r, name)
				}
			}
		} else if predicate == "blog:true" || predicate == "blog:false" {
			blog := predicate == "blog:true"
			re := regexp.MustCompile(`(^|/)\d\d\d\d-\d\d-\d\d`)
			for _, name := range names {
				match := re.MatchString(name)
				if blog && match || !blog && !match {
					r = append(r, name)
				}
			}
		} else {
			log.Printf("Unsupported predicate: %s", predicate)
		}
		names = intersection(names, r)
	}
	return names
}

// grep searches the files for matches to all the tokens. It returns
// just a single page of results based [from:to-1] and returns if
// there are more results.
func grep(tokens, names []string, from, to int) ([]*Page, bool) {
	pages := make([]*Page, 0)
	i := 0
NameLoop:
	for _, name := range names {
		p, err := loadPage(name)
		if err != nil {
			log.Printf("grep: cannot load %s: %s", name, err)
			panic(err)
		}
		body := strings.ToLower(string(p.Body))
		for _, token := range tokens {
			if !strings.Contains(body, token) {
				continue NameLoop
			}
		}
		i++
		if i > from {
			pages = append(pages, p)
		}
		if i > to {
			return pages, true
		}
	}
	return pages, false
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
	s := &Search{Query: q, Items: items, Previous: page - 1, Page: page, Next: page + 1,
		Results: len(items) > 0, More: more}
	renderTemplate(w, "search", s)
}
