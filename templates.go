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

// loadTemplates loads the templates. If templates have already been loaded, return immediately.
func loadTemplates() {
	if templates != nil {
		return
	}
	// walk the directory, load templates and add directories
	templates = make(map[string]*template.Template)
	filepath.Walk(".", loadTemplate)
	log.Println(len(templates), "templates loaded")
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

// updateTemplate checks whether this is a valid template file and if so, reloads it.
func updateTemplate(path string) {
	if strings.HasSuffix(path, ".html") &&
		slices.Contains(templateFiles, filepath.Base(path)) {
		t, err := template.ParseFiles(path)
		if err != nil {
			log.Println("Template:", path, err)
		} else {
			templates[path] = t
			log.Println("Parse template:", path)
		}
	}
}

// renderTemplate is the helper that is used to render the templates with data.
// A template in the same directory is preferred, if it exists.
func renderTemplate(w http.ResponseWriter, dir, tmpl string, data any) {
	loadTemplates()
	base := tmpl + ".html"
	t := templates[filepath.Join(dir, base)]
	if t == nil {
		t = templates[base]
	}
	if t == nil {
		log.Println("Template not found:", base)
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}
	err := t.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
