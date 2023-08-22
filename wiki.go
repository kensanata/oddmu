package main

import (
	"os"
	"fmt"
	"github.com/microcosm-cc/bluemonday"
	"github.com/gomarkdown/markdown"
	"html/template"
	"net/http"
	"strings"
	"regexp"
)

var templates = template.Must(template.ParseFiles("edit.html", "view.html"))

var validPath = regexp.MustCompile("^/(edit|save|view)/(([a-z]+/)?[^/]+)$")
var titleRegexp = regexp.MustCompile("(?m)^#\\s*(.*)\n+")

type Page struct {
	Title string
	Name  string
	Body  []byte
	Html  template.HTML
}

func (p *Page) save() error {
	filename := p.Name + ".md"
	return os.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := title + ".md"
	body, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	name := title
	s := string(body)
	m := titleRegexp.FindStringSubmatch(s)
	if m != nil {
		title = m[1]
		body = []byte(strings.Replace(s, m[0], "", 1))
	}
	return &Page{Title: title, Name: name, Body: body}, nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/view/index", http.StatusFound)
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	// Short cut for text files
	if (strings.HasSuffix(title, ".txt")) {
		body, err := os.ReadFile(title)
		if err == nil {
			w.Write(body)
			return
		}
	}
	// Attempt to load Markdown page; edit it if this fails
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	// Render the Markdown to HTML, sanitizing it
	maybeUnsafeHTML := markdown.ToHTML(p.Body, nil, nil)
	html := bluemonday.UGCPolicy().SanitizeBytes(maybeUnsafeHTML)
	p.Html = template.HTML(html);
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func makeHandler(fn func (http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func getPort() string {
	port := os.Getenv("ODDMU_PORT")
	if port == "" {
		port = "8080"
	}
	return port
}

func main() {
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))

	port := getPort()
	fmt.Println("Serving a wiki on port " + port)
	http.ListenAndServe(":" + port, nil)
}
