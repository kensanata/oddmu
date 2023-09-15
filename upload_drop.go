package main

import (
	"io"
	"net/http"
	"os"
	"path"
)

// uploadHandler uses the "upload.html" template to enable uploads.
// The file is saved using the saveUploadHandler.
func uploadHandler(w http.ResponseWriter, r *http.Request, dir string) {
	renderTemplate(w, "upload", dir)
}

// dropHandler takes the "name" form field and the "file" form
// file and saves the file under the given name. The browser is
// redirected to the view of that file.
func dropHandler(w http.ResponseWriter, r *http.Request, dir string) {
	d := path.Dir(dir)
	// ensure the directory exists
	fi, err := os.Stat(d)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !fi.IsDir() {
		http.Error(w, "file exists", http.StatusInternalServerError)
		return
	}
	filename := r.FormValue("name")
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()
	// backup an existing file with the same name
	_, err = os.Stat(filename)
	if err != nil {
		os.Rename(filename, filename + "~")
	}
	// create the new file
	dst, err := os.Create(d + "/" + filename)
	if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
	http.Redirect(w, r, "/view/"+filename, http.StatusFound)
}

