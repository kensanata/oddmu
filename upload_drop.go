package main

// The imaging library uses image.Decode internally. This function can use all image decoders available at that time.
// This is why we import goheif for side effects: HEIC files are read correctly.

import (
	_ "github.com/bashdrew/goheif"
	"github.com/disintegration/imaging"
	"github.com/edwvee/exiffix"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type upload struct {
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
	data := &upload{Dir: dir}
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
		data.Last = last
		data.Name, _ = next(last)
	}
	renderTemplate(w, dir, "upload", data)
}

// next returns the next name for a string matching lastRe. The last number in the given string is incremented by one
// ("a2b" → "a3b"). The second return value indicates whether such a replacement was made or not.
func next(s string) (string, bool) {
	m := lastRe.FindStringSubmatch(s)
	if m != nil {
		n, err := strconv.Atoi(m[2])
		if err == nil {
			return m[1] + strconv.Itoa(n+1) + m[3], true
		}
	}
	return s, false
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
	filename := filepath.Base(name)
	// no overwriting of hidden files or adding subdirectories
	if strings.HasPrefix(filename, ".") || filepath.Dir(name) != "." {
		http.Error(w, "no filename", http.StatusForbidden)
		return
	}
	// prepare for image encoding (saving) with the encoder based on the desired file name extensions
	var format imaging.Format
	quality := 75
	maxwidth := r.FormValue("maxwidth")
	mw := 0
	if len(maxwidth) > 0 {
		mw, err = strconv.Atoi(maxwidth)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		data.Add("maxwidth", maxwidth)
		// determine how the file will be written
		ext := strings.ToLower(filepath.Ext(filename))
		switch ext {
		case ".png":
			format = imaging.PNG
		case ".jpg", ".jpeg":
			q := r.FormValue("quality")
			if len(q) > 0 {
				quality, err = strconv.Atoi(q)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				data.Add("quality", q)
			}
			format = imaging.JPEG
		default:
			http.Error(w, "Resizing images requires a .png, .jpg or .jpeg extension for the filename", http.StatusBadRequest)
			return
		}
	}
	first := true
	for _, fhs := range r.MultipartForm.File["file"] {
		file, err := fhs.Open()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()
		if !first {
			s, ok := next(filename)
			if ok {
				filename = s
			} else {
				ext := filepath.Ext(s)
				filename = s[:len(s)-len(ext)] + "-1" + ext
			}
		}
		first = false
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
		if mw > 0 {
			// do not use imaging.Decode(file, imaging.AutoOrientation(true)) because that only works for JPEG files
			img, fmt, err := exiffix.Decode(file)
			if err != nil {
				http.Error(w, "The image could not be decoded (only PNG, JPG and HEIC formats are supported for resizing)", http.StatusBadRequest)
				return
			}
			log.Println("Decoded", fmt, "file")
			res := imaging.Resize(img, mw, 0, imaging.Lanczos) // preserve aspect ratio
			// imaging functions don't return errors but empty images…
			if !res.Rect.Empty() {
				img = res
			}
			// images are always reencoded, so image quality goes down
			err = imaging.Encode(dst, img, format, imaging.JPEGQuality(quality))
			if err != nil {
				log.Println(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
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
	}
	data.Set("last", filename) // has no slashes
	http.Redirect(w, r, "/upload/"+dir+"?"+data.Encode(), http.StatusFound)
}

// Today returns the date, as a string, for use in templates.
func (u *upload) Today() string {
	return time.Now().Format(time.DateOnly)
}
