package main

import (
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"slices"
	"strings"
	"sync"
)

// templateFiles are the various HTML template files used. These files must exist in the root directory for Oddmu to be
// able to generate HTML output. This always requires a template.
var templateFiles = []string{"edit.html", "add.html", "view.html", "preview.html",
	"diff.html", "search.html", "static.html", "upload.html", "feed.html",
	"list.html"}

// templateStore controls access to map of parsed HTML templates. Make sure to lock and unlock as appropriate. See
// renderTemplate and loadTemplates.
type templateStore struct {
	sync.RWMutex

	// template is a map of parsed HTML templates. The key is their filepath name. By default, the map only contains
	// top-level templates like "view.html". Subdirectories may contain their own templates which override the
	// templates in the root directory. If so, they are filepaths like "dir/view.html".
	template map[string]*template.Template
}

var templates templateStore

// loadTemplates loads the templates. If templates have already been loaded, return immediately.
func loadTemplates() {
	if templates.template != nil {
		return
	}
	templates.Lock()
	defer templates.Unlock()
	// walk the directory, load templates and add directories
	templates.template = make(map[string]*template.Template)
	filepath.Walk(".", loadTemplate)
	log.Println(len(templates.template), "templates loaded")
}

// loadTemplate is used to walk the directory. It loads all the template files it finds, including the ones in
// subdirectories. This is called with templates already locked.
func loadTemplate(fp string, info fs.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if strings.HasSuffix(fp, ".html") &&
		slices.Contains(templateFiles, filepath.Base(fp)) {
		t, err := template.ParseFiles(fp)
		if err != nil {
			log.Println("Cannot parse template:", fp, err)
			// ignore error
		} else {
			templates.template[fp] = t
		}
	}
	return nil
}

// updateTemplate checks whether this is a valid template file and if so, reloads it.
func updateTemplate(fp string) {
	if strings.HasSuffix(fp, ".html") &&
		slices.Contains(templateFiles, filepath.Base(fp)) {
		t, err := template.ParseFiles(fp)
		if err != nil {
			log.Println("Template:", fp, err)
		} else {
			templates.Lock()
			defer templates.Unlock()
			templates.template[fp] = t
			log.Println("Parse template:", fp)
		}
	}
}

// removeTemplate removes a template unless it's a root template because that would result in the site being unusable.
func removeTemplate(fp string) {
	if slices.Contains(templateFiles, filepath.Base(fp)) &&
		filepath.Dir(fp) != "." {
		templates.Lock()
		defer templates.Unlock()
		delete(templates.template, fp)
		log.Println("Discard template:", fp)
	}
}

// renderTemplate is the helper that is used to render the templates with data.
// A template in the same directory is preferred, if it exists.
func renderTemplate(w http.ResponseWriter, dir, tmpl string, data any) {
	loadTemplates()
	base := tmpl + ".html"
	templates.RLock()
	defer templates.RUnlock()
	t := templates.template[filepath.Join(dir, base)]
	if t == nil {
		t = templates.template[base]
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
