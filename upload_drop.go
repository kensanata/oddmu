package main

import (
	"github.com/anthonynsimon/bild/imgio"
	"github.com/anthonynsimon/bild/transform"
	"github.com/bashdrew/goheif"
	"image/jpeg"
	"image/png"
	"io"
	"log"
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

// uploadHandler uses the "upload.html" template to enable uploads. The file is saved using the dropHandler. URL
// parameters are used to copy name, maxwidth and quality from the previous upload. If the previous name contains a
// number, this is incremented by one.
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
	name := r.FormValue("filename")
	if name != "" {
		data.Name = name
	} else if last := r.FormValue("last"); last != "" {
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
	renderTemplate(w, dir, "upload", data)
}

// dropHandler takes the "name" form field and the "file" form file and saves the file under the given name. The browser
// is redirected to the view of that file. Some errors are for the users and some are for users and the admins. Those
// later errors are printed, too.
func dropHandler(w http.ResponseWriter, r *http.Request, dir string) {
	d := filepath.Dir(filepath.FromSlash(dir))
	// ensure the directory exists
	fi, err := os.Stat(d)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if !fi.IsDir() {
		http.Error(w, "directory does not exist", http.StatusBadRequest)
		return
	}
	data := url.Values{}
	name := r.FormValue("name")
	data.Set("last", name)
	filename := filepath.Base(name)
	// no overwriting of hidden files or adding subdirectories
	if strings.HasPrefix(filename, ".") || filepath.Dir(name) != "." {
		http.Error(w, "no filename", http.StatusForbidden)
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()
	path := filepath.Join(d, filename)
	watches.ignore(path)
	err = backup(path)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	dst, err := os.Create(path)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	// if a resize was requested
	maxwidth := r.FormValue("maxwidth")
	if len(maxwidth) > 0 {
		mw, err := strconv.Atoi(maxwidth)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		data.Add("maxwidth", maxwidth)
		// determine how the file will be written
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
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				data.Add("quality", quality)
			}
			encoder = imgio.JPEGEncoder(q)
		default:
			http.Error(w, "Resizing images requires a .png, .jpg or .jpeg extension for the filename", http.StatusBadRequest)
			return
		}
		// try and decode the data in various formats
		img, err := jpeg.Decode(file)
		if err != nil {
			img, err = png.Decode(file)
		}
		if err != nil {
			img, err = goheif.Decode(file)
		}
		if err != nil {
			http.Error(w, "The image could not be decoded (only PNG, JPG and HEIC formats are supported for resizing)", http.StatusBadRequest)
			return
		}
		rect := img.Bounds()
		width := rect.Max.X - rect.Min.X
		if width > mw {
			height := (rect.Max.Y - rect.Min.Y) * mw / width
			img = transform.Resize(img, mw, height, transform.Linear)
			if err := imgio.Save(path, img, encoder); err != nil {
				log.Println(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "The file is too small for this", http.StatusBadRequest)
			return
		}
	} else {
		// just copy the bytes
		n, err := io.Copy(dst, file)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// if zero bytes were copied, delete the file instead
		if n == 0 {
			err := os.Remove(path)
			if err != nil {
				log.Println(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			log.Println("Delete", path)
		}
	}
	username, _, ok := r.BasicAuth()
	if ok {
		log.Println("Save", path, "by", username)
	} else {
		log.Println("Save", path)
	}
	updateTemplate(path)
	http.Redirect(w, r, "/upload/"+dir+"?"+data.Encode(), http.StatusFound)
}
