package main

// The imaging library uses image.Decode internally. This function can use all image decoders available at that time.
// This is why we import heic for side effects. For writing, the particular encoders have to be imported.

import (
	"errors"
	"fmt"
	_ "github.com/gen2brain/heic"
	"github.com/disintegration/imaging"
	"github.com/edwvee/exiffix"
	"github.com/gen2brain/webp"
	"image/png"
	"image/jpeg"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
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
	Actual   []string
}

var lastRe = regexp.MustCompile(`^(.*?)([0-9]+)([^0-9]*)$`)
var baseRe = regexp.MustCompile(`^(.*?)-[0-9]+$`)

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
	var err error
	if name != "" {
		data.Name, err = next(dir, name, 0)
	} else if last := r.FormValue("last"); last != "" {
		data.Last = last
		mimeType := mime.TypeByExtension(filepath.Ext(last))
		data.Image = strings.HasPrefix(mimeType, "image/")
		data.Name, err = next(dir, last, 1)
		data.Actual = r.Form["actual"]
	}
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	renderTemplate(w, dir, "upload", data)
}

// next returns the next filename for a filename containing a number. The last number is identified using lastRe. This
// number is increased by the second argument. Then, for as long as a file with that number exists, the number is
// increased by one. Thus, when called with "image-1.jpg", 0 the string returned will be "image-1.jpg" if no such file
// exists. If "image-1.jpg" exists but "image-2.jpg" does not, then that is returned. When called with "image.jpg"
// (containing no number) and the file does not exist, it is returned unchanged. If it exists, "image-1.jpg" is assumed
// and the algorithm described previously is used to find the next unused filename.
func next(dir, fn string, i int) (string, error) {
	m := lastRe.FindStringSubmatch(fn)
	if m == nil {
		_, err := os.Stat(filepath.Join(dir, fn))
		if err != nil {
			return fn, nil
		}
		ext := filepath.Ext(fn)
		// faking it
		m = []string{"", fn[:len(fn)-len(ext)] + "-", "0", ext}
	}
	n, err := strconv.Atoi(m[2])
	if err == nil {
		n += i
		for {
			s := m[1] + strconv.Itoa(n) + m[3]
			_, err = os.Stat(filepath.Join(dir, s))
			if err != nil {
				return s, nil
			}
			n += 1
		}
	}
	return fn, fmt.Errorf("unable to find next filename after %s", fn)
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
		http.Error(w, "not a directory", http.StatusBadRequest)
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
	// Quality is a number. If no quality is set and a quality is required, 75 is used.
	q := 75
	quality := r.FormValue("quality")
	if len(quality) > 0 {
		q, err = strconv.Atoi(quality)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		data.Set("quality", quality) // remember for the next request
	}
	// maxwidth is a number. If no maxwidth is set, no resizing is done.
	mw := 0
	maxwidth := r.FormValue("maxwidth")
	if len(maxwidth) > 0 {
		mw, err = strconv.Atoi(maxwidth)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		data.Set("maxwidth", maxwidth) // remember for the next request
	}
	// the destination image format is determined by the extension
	to := strings.ToLower(filepath.Ext(filename))
	first := true
	for _, fhs := range r.MultipartForm.File["file"] {
		file, err := fhs.Open()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()
		// the first filename overwrites!
		if !first {
			filename, err = next(d, filename, 1)
			if err != nil {
				log.Println(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
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
		// the source image format is determined by the extension
		from := strings.ToLower(filepath.Ext(fhs.Filename))
		if q != 75 || mw > 0 || from != to {
			// do not use imaging.Decode(file, imaging.AutoOrientation(true)) because that only works for JPEG files
			img, fmt, err := exiffix.Decode(file)
			if err != nil {
				http.Error(w, "The image could not be decoded from " + from + " format", http.StatusBadRequest)
				return
			}
			log.Println("Decoded", fmt, "file")
			if mw > 0 {
				res := imaging.Resize(img, mw, 0, imaging.Lanczos) // preserve aspect ratio
				// imaging functions don't return errors but empty images…
				if !res.Rect.Empty() {
					img = res
				}
			}
			// images are always reencoded, so image quality goes down
			switch (to) {
			case ".png":
				err = png.Encode(dst, img)
			case ".jpg", ".jpeg":
				err = jpeg.Encode(dst, img, &jpeg.Options{Quality: q})
			case ".webp":
				err = webp.Encode(dst, img, webp.Options{Quality: q}) // Quality of 100 implies Lossless.
			default:
				err = errors.New("Unsupported destination format for image conversion: " + to)
			}
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
		data.Add("actual", filename)
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

// Base returns a page name matching the first uploaded file: no extension and no appended number. If the name
// refers to a directory, returns "index". This is used to create the form target in "upload.html", for example.
func (u *upload) Base() string {
	n := u.Name[:strings.LastIndex(u.Name, ".")]
	m := baseRe.FindStringSubmatch(n)
	if m != nil {
		return m[1]
	}
	if n == "." {
		return "index"
	}
	return n
}

// Title returns the title of the matching page, if it exists.
func (u *upload) Title() string {
	index.RLock()
	defer index.RUnlock()
	name := path.Join(u.Dir, u.Base())
	title, ok := index.titles[name]
	if ok {
		return title
	}
	return name
}

// Today returns the date, as a string, for use in templates.
func (u *upload) Today() string {
	return time.Now().Format(time.DateOnly)
}
