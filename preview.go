package main

import (
	"net/http"
	"strings"
)

// previewHandler is a bit like saveHandler and viewHandler. Instead of saving the date to a page, we create a synthetic
// Page and render it. Note that when saving, the carriage returns (\r) are removed. We need to do this as well,
// otherwise the rendered template has garbage bytes at the end. Note also that we need to remove the title from the
// page so that the preview works as intended (and much like the "view.html" template) where as the editing requires the
// page content including the header… which is why it needs to be added in the "preview.html" template. This makes me
// sad.
func previewHandler(w http.ResponseWriter, r *http.Request, path string) {
	body := strings.ReplaceAll(r.FormValue("body"), "\r", "")
	p := &Page{Name: path, Body: []byte(body)}
	p.handleTitle(true)
	p.renderHtml()
	renderTemplate(w, p.Dir(), "preview", p)
}
