package main

import (
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
	"github.com/microcosm-cc/bluemonday"
	"html/template"
	"net/http"
	"strings"
	"regexp"
	"bytes"
	"fmt"
	"os"
)

// Templates are parsed at startup.
var templates = template.Must(template.ParseFiles("edit.html", "view.html", "search.html"))

// validPath is a regular expression where the second group matches a
// page, so when the handler for "/edit/" is called, a URL path of
// "/edit/foo" results in the editHandler being called with title
// "foo". The regular expression doesn't define the handlers (this
// happens in the main function).
var validPath = regexp.MustCompile("^/([^/]+)/(.+)$")

// titleRegexp is a regular expression matching a level 1 header line
// in a Markdown document. The first group matches the actual text and
// is used to provide an title for pages. If no title exists in the
// document, the page name is used instead.
var titleRegexp = regexp.MustCompile("(?m)^#\\s*(.*)\n+")

// Page is a struct containing information about a single page. Title
// is the title extracted from the page content using titleRegexp.
// Name is the filename without extension (so a filename of "foo.md"
// results in the Name "foo"). Body is the Markdown content of the
// page and Html is the rendered HTML for that Markdown.
type Page struct {
	Title string
	Name  string
	Body  []byte
	Html  template.HTML
}

// save saves a Page. The filename is based on the Page.Name and gets
// the ".md" extension. Page.Body is saved, without any carriage
// return characters ("\r"). The file permissions used are readable
// and writeable for the current user, i.e. u+rw or 0600.
func (p *Page) save() error {
	filename := p.Name + ".md"
	updateIndex(p)
	return os.WriteFile(filename, bytes.ReplaceAll(p.Body, []byte{'\r'}, []byte{}), 0600)
}

// loadPage loads a Page given a name. The filename loaded is that
// Page.Name with the ".md" extension. The Page.Title is set to the
// Page.Name (and possibly changed, later). The Page.Body is set to
// the file content. The Page.Html remains undefined (there is no
// caching).
func loadPage(name string) (*Page, error) {
	filename := name + ".md"
	body, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: name, Name: name, Body: body}, nil
}

// handleTitle extracts the title from a Page and sets Page.Title, if
// any.
func (p* Page) handleTitle() {
	s := string(p.Body)
	m := titleRegexp.FindStringSubmatch(s)
	if m != nil {
		p.Title = m[1]
		p.Body = []byte(strings.Replace(s, m[0], "", 1))
	}
}

// renderHtml renders the Page.Body to HTML and sets Page.Html.
func (p* Page) renderHtml() {
	maybeUnsafeHTML := markdown.ToHTML(p.Body, nil, nil)
	html := bluemonday.UGCPolicy().SanitizeBytes(maybeUnsafeHTML)
	p.Html = template.HTML(html);
}

// plainText renders the Page.Body to plain text and returns it,
// ignoring all the Markdown.
func (p* Page) plainText() string {
	parser := parser.New()
	doc := markdown.Parse(p.Body, parser)
	text := []byte("")
	ast.WalkFunc(doc, func(node ast.Node, entering bool) ast.WalkStatus {
		if entering && node.AsLeaf() != nil {
			text = append(text, node.AsLeaf().Literal...)
			text = append(text, []byte(" ")...)
		}
		return ast.GoToNext
	})
	return strings.ReplaceAll(string(text), "\n", " ")
}

func renderTemplate(w http.ResponseWriter, tmpl string, data any) {
	err := templates.ExecuteTemplate(w, tmpl+".html", data)
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
	p.handleTitle()
	p.renderHtml()
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title, Name: title}
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Name: title, Body: []byte(body)}
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
		if m != nil {
			fn(w, r, m[2])
		} else {
			http.NotFound(w, r)
		}
	}
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	q := r.FormValue("q")
	ids := index.Query(q)
	items := make([]Page, len(ids))
	for i, id := range ids {
		name := documents[id]
		p, err := loadPage(name)
		if err != nil {
			fmt.Printf("Error loading %s\n", name)
		} else {
			p.handleTitle()
			extract := []byte(snippets(q, p.plainText()))
			html := bluemonday.UGCPolicy().SanitizeBytes(extract)
			p.Html = template.HTML(html)
			items[i] = *p
		}
	}
	s := &Search{Query: q, Items: items, Results: len(items) > 0}
	renderTemplate(w, "search", s)
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
	http.HandleFunc("/search", searchHandler)
	loadIndex()
	port := getPort()
	fmt.Printf("Serving a wiki on port %s\n", port)
	http.ListenAndServe(":" + port, nil)
}
