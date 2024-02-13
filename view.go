package main

import (
	"io"
	"net/http"
	"os"
	urlpath "path"
	"path/filepath"
	"strings"
	"time"
)

// rootHandler just redirects to /view/index. The root handler handles requests to the root path, and – implicity – all
// unhandled request. Thus, if the URL path is not "/", return a 404 NOT FOUND response.
func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
	} else {
		http.Redirect(w, r, "/view/index", http.StatusFound)
	}
}

// viewHandler serves pages. If the requested URL ends in ".rss" and the corresponding file ending with ".md" exists, a
// feed is generated and the "feed.html" template is used (it is used to generate a RSS 2.0 feed, even if the extension
// is ".html"). If the requested URL maps to a page name, the corresponding file (by appending ".md") is loaded and
// served using the "view.html" template. If the requested URL maps to an existing file, it is served (you can therefore
// request the ".md" files directly). If the requested URL maps to a directory, the browser is redirected to the index
// page. If none of the above, the browser is redirected to an edit page.
//
// Uploading files ending in ".rss" does not prevent RSS feed generation.
//
// Caching: a 304 NOT MODIFIED is returned if the request has an If-Modified-Since header that matches the file's
// modification time, truncated to one second. Truncation is required because the file's modtime has sub-second
// precision and the HTTP timestamp for the Last-Modified header has not.
func viewHandler(w http.ResponseWriter, r *http.Request, path string) {
	const (
		unknown = iota
		file
		page
		rss
		dir
	)
	t := unknown
	if strings.HasSuffix(path, ".rss") {
		path = path[:len(path)-4]
		t = rss
	}
	fp := filepath.FromSlash(path)
	fi, err := os.Stat(fp+".md")
	if err == nil {
		if fi.IsDir() {
			t = dir // directory ending in ".md"
		} else if t == unknown {
			t = page
		}
		// otherwise t == rss
	} else {
		if fp == "" {
			fp = "." // make sure Stat works
		}
		fi, err = os.Stat(fp)
		if err == nil {
			if fi.IsDir() {
				t = dir
			} else {
				t = file
			}
		}
	}
	// if nothing was found, offer to create it
	if t == unknown {
		http.Redirect(w, r, "/edit/"+path, http.StatusFound)
		return
	}
	// directories are redirected to the index page
	if t == dir {
		http.Redirect(w, r, urlpath.Join("/view", path, "index"), http.StatusFound)
		return
	}
	// if the page has not been modified, return (file, rss or page)
	h, ok := r.Header["If-Modified-Since"]
	if ok {
		ti, err := http.ParseTime(h[0])
		if err == nil && !fi.ModTime().Truncate(time.Second).After(ti) {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}
	// if only the headers were requested, return
	w.Header().Set("Last-Modified", fi.ModTime().UTC().Format(http.TimeFormat))
	if r.Method == http.MethodHead {
		w.WriteHeader(http.StatusOK)
		return
	}
	if t == file {
		file, err := os.Open(fp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = io.Copy(w, file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}
	p, err := loadPage(path)
	if err != nil {
		if t == rss {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Redirect(w, r, "/edit/"+path, http.StatusFound)
		return
	}
	p.handleTitle(true)
	if t == rss {
		it := feed(p, fi.ModTime())
		w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>`))
		renderTemplate(w, p.Dir(), "feed", it)
		return
	}
	p.renderHtml()
	renderTemplate(w, p.Dir(), "view", p)
}
