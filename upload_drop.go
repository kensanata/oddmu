package main

import (
	"github.com/anthonynsimon/bild/imgio"
	"github.com/anthonynsimon/bild/transform"
	"image/jpeg"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type Upload struct {
	Dir      string
	Name     string
	Last     string
	Image    bool
	MaxWidth string
	Quality  string
}

var lastRe = regexp.MustCompile(`^(.*)([0-9]+)(.*)$`)

// uploadHandler uses the "upload.html" template to enable uploads.
// The file is saved using the saveUploadHandler. URL parameter are
// used to copy name, maxwidth and quality from the previous upload.
// If the previous name contains a number, this is incremented by
// one.
func uploadHandler(w http.ResponseWriter, r *http.Request, dir string) {
	data := &Upload{Dir: dir}
	maxwidth := r.FormValue("maxwidth")
	if maxwidth != "" {
		data.MaxWidth = maxwidth
	}
	quality := r.FormValue("quality")
	if quality != "" {
		data.Quality = quality
	}
	last := r.FormValue("last")
	if last != "" {
		ext := strings.ToLower(filepath.Ext(last))
		switch ext {
		case ".png", ".jpg", ".jpeg":
			data.Image = true
		}
		data.Last = path.Join(dir, last)
		m := lastRe.FindStringSubmatch(last)
		if m != nil {
			n, err := strconv.Atoi(m[2])
			if err == nil {
				data.Name = m[1] + strconv.Itoa(n+1) + m[3]
			}
		}
	}
	renderTemplate(w, "upload", data)
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
	data := url.Values{}
	name := r.FormValue("name")
	data.Set("last", name)
	filename := filepath.Base(name)
	if filename == "." || filepath.Dir(name) != "." {
		http.Error(w, "no filename", http.StatusInternalServerError)
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()
	// backup an existing file with the same name
	_, err = os.Stat(filename)
	if err != nil {
		os.Rename(filename, filename+"~")
	}
	// create the new file
	path := d + "/" + filename
	dst, err := os.Create(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// if a resize was requested
	maxwidth := r.FormValue("maxwidth")
	if len(maxwidth) > 0 {
		mw, err := strconv.Atoi(maxwidth)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data.Add("maxwidth", maxwidth)
		ext := strings.ToLower(filepath.Ext(path))
		var encoder imgio.Encoder
		switch ext {
		case ".png":
			encoder = imgio.PNGEncoder()
		case ".jpg", ".jpeg":
			q := jpeg.DefaultQuality
			quality := r.FormValue("quality")
			if len(quality) > 0 {
				q, err = strconv.Atoi(quality)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				data.Add("quality", quality)
			}
			encoder = imgio.JPEGEncoder(q)
		default:
			http.Error(w, "only .png, .jpg, or .jpeg files are supported", http.StatusInternalServerError)
			return
		}
		img, err := imgio.Open(path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		rect := img.Bounds()
		width := rect.Max.X - rect.Min.X
		if width > mw {
			height := (rect.Max.Y - rect.Min.Y) * mw / width
			img = transform.Resize(img, mw, height, transform.Linear)
			if err := imgio.Save(path, img, encoder); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

	}
	http.Redirect(w, r, "/upload/"+d+"/?"+data.Encode(), http.StatusFound)
}
