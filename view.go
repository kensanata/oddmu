package main

import(
	"net/http"
	"os"
)

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
