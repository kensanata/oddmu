package main

import (
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// ListItem is used to display the list of files.
type File struct {
	Name, Title string
	IsDir, IsUp bool
	// Date is the last modification date of the file storing the page. As the pages used by Oddmu are plain
	// Markdown files, they don't contain any metadata. Instead, the last modification date of the file is used.
	// This makes it work well with changes made to the files outside of Oddmu.
	Date string
}

type List struct {
	Dir string
	Files []File
}

// listHandler uses the "list.html" template to enable file management in a particular directory.
func listHandler(w http.ResponseWriter, r *http.Request, name string) {
	files := []File{}
	d := filepath.FromSlash(name)
	if d == "" {
		d = "."
	} else if !strings.HasSuffix(d, "/") {
		http.Redirect(w, r, "/list/" + nameEscape(name) + "/", http.StatusFound)
		return
	} else {
		it := File{Name: "..", IsUp: true, IsDir: true }
		files = append(files, it)
	}
	err := filepath.Walk(d, func (fp string, fi fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		isDir := false
		if fi.IsDir() {
			if d == fp {
				return nil
			}
			isDir = true
		}
		name := filepath.ToSlash(fp)
		base := filepath.Base(fp)
		title := ""
		if !isDir && strings.HasSuffix(name, ".md") {
			index.RLock()
			defer index.RUnlock()
			title = index.titles[name[:len(name)-3]]
		} else if isDir {
			// even on Windows, this looks like a Unix directory
			base += "/"
		}
		it := File{Name: base, Title: title, Date: fi.ModTime().Format(time.DateTime), IsDir: isDir }
		files = append(files, it)
		if isDir {
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	renderTemplate(w, d, "list", &List{Dir: name, Files: files})
}


// deleteHandler deletes the named file and then redirects back to the list
func deleteHandler(w http.ResponseWriter, r *http.Request, name string) {
	fn := filepath.FromSlash(name)
	err := os.RemoveAll(fn) // and all its children!
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/list/" + nameEscape(path.Dir(name)) + "/", http.StatusFound)
}

// renameHandler renames the named file and then redirects back to the list
func renameHandler(w http.ResponseWriter, r *http.Request, name string) {
	fn := filepath.FromSlash(name)
	dir := path.Dir(name)
	target := path.Join(dir, r.FormValue("name"))
	if (isHiddenName(target)) {
		http.Error(w, "the target file would be hidden", http.StatusForbidden)
		return
	}
	err := os.Rename(fn, filepath.FromSlash(target))
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/list/" + nameEscape(path.Dir(filepath.ToSlash(target))) + "/", http.StatusFound)
}
