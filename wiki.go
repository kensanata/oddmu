package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"regexp"
)

// Templates are parsed at startup.
var templates = template.Must(
	template.ParseFiles("edit.html", "add.html", "view.html", "search.html"))

// validPath is a regular expression where the second group matches a
// page, so when the editHandler is called, a URL path of "/edit/foo"
// results in the editHandler being called with title "foo". The
// regular expression doesn't define the handlers (this happens in the
// main function).
var validPath = regexp.MustCompile("^/([^/]+)/(.+)$")

// titleRegexp is a regular expression matching a level 1 header line
// in a Markdown document. The first group matches the actual text and
// is used to provide an title for pages. If no title exists in the
// document, the page name is used instead.
var titleRegexp = regexp.MustCompile("(?m)^#\\s*(.*)\n+")

// renderTemplate is the helper that is used render the templates with
// data.
func renderTemplate(w http.ResponseWriter, tmpl string, data any) {
	err := templates.ExecuteTemplate(w, tmpl+".html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// rootHandler just redirects to /view/index.
func rootHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/view/index", http.StatusFound)
}

// viewHandler serves existing files (including markdown files with
// the .md extension). If the requested file does not exist, a page
// with the same name is loaded. This means adding the .md extension
// and using the "view.html" template to render the HTML. Both
// attempts fail, the browser is redirected to an edit page.
func viewHandler(w http.ResponseWriter, r *http.Request, name string) {
	body, err := os.ReadFile(name)
	if err == nil {
		w.Write(body)
		return
	}
	p, err := loadPage(name)
	if err == nil {
		p.handleTitle(true)
		p.renderHtml()
		renderTemplate(w, "view", p)
		return
	}
	http.Redirect(w, r, "/edit/"+name, http.StatusFound)
}

// editHandler uses the "edit.html" template to present an edit page.
// When editing, the page title is not overriden by a title in the
// text. Instead, the page name is used. The edit is saved using the
// saveHandler.
func editHandler(w http.ResponseWriter, r *http.Request, name string) {
	p, err := loadPage(name)
	if err != nil {
		p = &Page{Title: name, Name: name}
	} else {
		p.handleTitle(false)
	}
	renderTemplate(w, "edit", p)
}

// saveHandler takes the "body" form parameter and saves it. The
// browser is redirected to the page view.
func saveHandler(w http.ResponseWriter, r *http.Request, name string) {
	body := r.FormValue("body")
	p := &Page{Name: name, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+name, http.StatusFound)
}

// addHandler uses the "add.html" template to present an empty edit
// page. What you type there is appended to the page using the
// appendHandler.
func addHandler(w http.ResponseWriter, r *http.Request, name string) {
	p, err := loadPage(name)
	if err != nil {
		p = &Page{Title: name, Name: name}
	} else {
		p.handleTitle(false)
	}
	renderTemplate(w, "add", p)
}

// appendHandler takes the "body" form parameter and appends it. The
// browser is redirected to the page view.
func appendHandler(w http.ResponseWriter, r *http.Request, name string) {
	body := r.FormValue("body")
	p, err := loadPage(name)
	if err != nil {
		p = &Page{Title: name, Name: name, Body: []byte(body)}
	} else {
		p.handleTitle(false)
		p.Body = append(p.Body, []byte(body)...)
	}
	err = p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+name, http.StatusFound)
}

// makeHandler returns a handler that uses the URL path without the
// first path element as its argument, e.g. if the URL path is
// /edit/foo/bar, the editHandler is called with "foo/bar" as its
// argument. This uses the second group from the validPath regular
// expression.
func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m != nil {
			fn(w, r, m[2])
		} else {
			http.NotFound(w, r)
		}
	}
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

// getPort returns the environment variable ODDMU_PORT or the default
// port, "8080".
func getPort() string {
	port := os.Getenv("ODDMU_PORT")
	if port == "" {
		port = "8080"
	}
	return port
}

// scheduleLoadIndex calls loadIndex and prints some messages before
// and after. For testing, call loadIndex directly and skip the
// messages.
func scheduleLoadIndex() {
	fmt.Print("Indexing pages\n")
	n, err := loadIndex()
	if err == nil {
		fmt.Printf("Indexed %d pages\n", n)
	} else {
		fmt.Println("Indexing failed")
	}
}

// scheduleLoadLanguages calls loadLanguages and prints some messages before
// and after. For testing, call loadLanguages directly and skip the
// messages.
func scheduleLoadLanguages(){
	fmt.Print("Loading languages\n")
	n := loadLanguages()
	fmt.Printf("Loaded %d languages\n", n)
}

func main() {
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.HandleFunc("/add/", makeHandler(addHandler))
	http.HandleFunc("/append/", makeHandler(appendHandler))
	http.HandleFunc("/search", searchHandler)
	go scheduleLoadIndex()
	go scheduleLoadLanguages()
	port := getPort()
	fmt.Printf("Serving a wiki on port %s\n", port)
	http.ListenAndServe(":"+port, nil)
}
