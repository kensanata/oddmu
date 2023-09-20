package main

import (
	"net/http"
	"os"
	"time"
)

// rootHandler just redirects to /view/index.
func rootHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/view/index", http.StatusFound)
}

// viewHandler serves existing files (including markdown files with
// the .md extension). If the requested file does not exist, a page
// with the same name is loaded. This means adding the .md extension
// and using the "view.html" template to render the HTML. Both
// attempts fail, the browser is redirected to an edit page. As far as
// caching goes: we respond with a 304 NOT MODIFIED if the request has
// an If-Modified-Since header that matches the file's modification
// time, truncated to one second, because the file's modtime has
// sub-second precision and the HTTP timestamp for the Last-Modified
// header has not.
func viewHandler(w http.ResponseWriter, r *http.Request, name string) {
	file := true
	fn := name
	fi, err := os.Stat(fn)
	if err != nil {
		file = false
		fn += ".md"
		fi, err = os.Stat(fn)
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
	p.renderHtml()
	renderTemplate(w, "view", p)
}
