package main

import (
	"log"
	"net/http"
)

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
	renderTemplate(w, p.Dir(), "edit", p)
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
	if r.FormValue("notify") == "on" {
		err = p.notify()
		if err != nil {
			log.Println("notify:", err)
		}
	}
	http.Redirect(w, r, "/view/"+name, http.StatusFound)
}
