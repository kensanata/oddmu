package main

import (
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

// rootHandler just redirects to /view/index.
func rootHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/view/index", http.StatusFound)
}

// viewHandler serves pages. If the requested URL maps to an existing file, it is served. If the requested URL maps to a
// directory, the browser is redirected to the index page. If the requested URL ends in ".rss" and the corresponding
// file ending with ".md" exists, a feed is generated and the "feed.html" template is used (it is used to generate a RSS
// 2.0 feed, no matter what the template's extension is). If the requested URL maps to a page name, the corresponding
// file (ending in ".md") is loaded and served using the "view.html" template. If none of the above, the browser is
// redirected to an edit page.
//
// Caching: a 304 NOT MODIFIED is returned if the request has an If-Modified-Since header that matches the file's
// modification time, truncated to one second. Truncation is required because the file's modtime has sub-second
// precision and the HTTP timestamp for the Last-Modified header has not.
func viewHandler(w http.ResponseWriter, r *http.Request, name string) {
	file := true
	rss := false
	if name == "" {
		name = "."
	}
	fn := name
	fi, err := os.Stat(fn)
	if err != nil {
		file = false
		if strings.HasSuffix(fn, ".rss") {
			rss = true
			name = fn[0 : len(fn)-4]
			fn = name
		}
		fn += ".md"
		fi, err = os.Stat(fn)
	} else if fi.IsDir() {
		http.Redirect(w, r, path.Join("/view", name, "index"), http.StatusFound)
		return
	}
	if err == nil {
		h, ok := r.Header["If-Modified-Since"]
		if ok {
			ti, err := http.ParseTime(h[0])
			if err == nil && !fi.ModTime().Truncate(time.Second).After(ti) {
				w.WriteHeader(http.StatusNotModified)
				return
			}
		}
		w.Header().Set("Last-Modified", fi.ModTime().UTC().Format(http.TimeFormat))
	}
	if r.Method == http.MethodHead {
		if err == nil {
			return
		}
		http.Redirect(w, r, "/edit/"+name, http.StatusFound)
	}
	if file {
		body, err := os.ReadFile(fn)
		if err != nil {
			// This is an internal error because os.Stat
			// says there is a file. Non-existent files
			// are treated like pages.
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		w.Write(body)
		return
	}
	p, err := loadPage(name)
	if err != nil {
		http.Redirect(w, r, "/edit/"+name, http.StatusFound)
		return
	}
	p.handleTitle(true)
	if rss {
		it := feed(p, fi.ModTime())
		w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>`))
		renderTemplate(w, "feed", it)
		return
	}
	p.renderHtml()
	renderTemplate(w, "view", p)
}
