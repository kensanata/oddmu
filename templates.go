package main

import (
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"slices"
	"strings"
)

// templateFiles are the various HTML template files used. These files must exist in the root directory for Oddmu to be
// able to generate HTML output. This always requires a template.
var templateFiles = []string{"edit.html", "add.html", "view.html",
	"diff.html", "search.html", "static.html", "upload.html", "feed.html"}

// templates are the parsed HTML templates used. See renderTemplate and loadTemplates. Subdirectories may contain their
// own templates which override the templates in the root directory.
var templates map[string]*template.Template

// loadTemplates loads the templates. These aren't always required. If the templates are required and cannot be loaded,
// this a fatal error and the program exits.
func loadTemplates() {
	if templates != nil {
		return
	}
	// walk the directory, load templates and add directories
	templates = make(map[string]*template.Template)
	filepath.Walk(".", loadTemplate)
}

// loadTemplate is used to walk the directory. It loads all the template files it finds, including the ones in
// subdirectories.
func loadTemplate(path string, info fs.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if strings.HasSuffix(path, ".html") &&
		slices.Contains(templateFiles, filepath.Base(path)) {
		t, err := template.ParseFiles(path)
		if err != nil {
			log.Println("Cannot parse template:", path, err)
			// ignore error
		} else {
			// log.Println("Parse template:", path)
			templates[path] = t
		}
	}
	return nil
}

func updateTemplate(path string) {
	if strings.HasSuffix(path, ".html") &&
		slices.Contains(templateFiles, filepath.Base(path)) {
		t, err := template.ParseFiles(path)
		if err != nil {
			log.Println("Template:", path, err)
		} else {
			templates[path] = t
			log.Println("Parsed", path)
		}
	}
}

// renderTemplate is the helper that is used to render the templates with data. If the templates cannot be found, that's
// fatal.
func renderTemplate(w http.ResponseWriter, tmpl string, data any) {
	loadTemplates()
	t := templates[tmpl+".html"]
	if t == nil {
		log.Println("Template not found:", tmpl)
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}
	err := t.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
