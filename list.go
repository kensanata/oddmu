package main

import (
	"io/fs"
	"log"
	"net/http"
	"os"
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
func listHandler(w http.ResponseWriter, r *http.Request, dir string) {
	files := []File{}
	d := filepath.FromSlash(dir)
	if d == "" {
		d = "."
	} else if !strings.HasSuffix(d, "/") {
		http.Redirect(w, r, "/list/"+d+"/", http.StatusFound)
		return
	} else {
		it := File{Name: "..", IsUp: true, IsDir: true }
		files = append(files, it)
	}
	err := filepath.Walk(d, func (path string, fi fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		isDir := false
		if fi.IsDir() {
			if d == path {
				return nil
			}
			isDir = true
		}
		name := filepath.ToSlash(path)
		base := filepath.Base(name)
		title := ""
		if !isDir && strings.HasSuffix(name, ".md") {
			index.RLock()
			defer index.RUnlock()
			title = index.titles[name[:len(name)-3]]
		}
		if isDir {
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
	renderTemplate(w, dir, "list", &List{Dir: dir, Files: files})
}


// deleteHandler deletes the named file and then redirects back to the list
func deleteHandler(w http.ResponseWriter, r *http.Request, path string) {
	fn := filepath.FromSlash(path)
	fi, err := os.Stat(fn)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = os.RemoveAll(fn) // and all its children!
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if fi.IsDir() {
		fn = filepath.Dir(fn) // net result is that the redirect goes to the parent
	}
	http.Redirect(w, r, "/list/"+filepath.Dir(fn)+"/", http.StatusFound)
}
