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

type Upload struct {
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
	data := &Upload{Dir: pathEncode(dir)}
	maxwidth := r.FormValue("maxwidth")
	if maxwidth != "" {
		data.MaxWidth = maxwidth
	}
	quality := r.FormValue("quality")
	if quality != "" {
		data.Quality = quality
	}
	name := r.FormValue("filename")
	if isHiddenName(name) {
		http.Error(w, "the file would be hidden", http.StatusForbidden)
		return
	}
	var err error
	if name != "" {
		data.Name, err = next(filepath.FromSlash(dir), name, 0)
	} else if last := r.FormValue("last"); last != "" {
		data.Last = last
		mimeType := mime.TypeByExtension(path.Ext(last))
		data.Image = strings.HasPrefix(mimeType, "image/")
		data.Name, err = next(filepath.FromSlash(dir), last, 1)
		data.Actual = make([]string, len(r.Form["actual"]))
		for i, s := range r.Form["actual"] {
			data.Actual[i] = pathEncode(s)
		}
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
		// faking a match as if "-0" had been matched
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
	dir = filepath.FromSlash(dir)
	// ensure the directory exists and that "" results in "."
	fi, err := os.Stat(filepath.Clean(dir))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if !fi.IsDir() {
		http.Error(w, "not a directory", http.StatusBadRequest)
		return
	}
	data := url.Values{}
	fn := r.FormValue("name")
	// This is like the id query parameter: it may not contain any slashes, so it's a path and a filepath.
	if strings.Contains(fn, "/") {
		http.Error(w, "the file may not contain slashes", http.StatusBadRequest)
		return
	}
	if isHiddenName(fn) {
		http.Error(w, "the file would be hidden", http.StatusForbidden)
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
	to := strings.ToLower(path.Ext(fn))
	first := true
	for _, fhs := range r.MultipartForm.File["file"] {
		log.Println("Reading", fhs.Filename)
		file, err := fhs.Open()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()
		// the first filename overwrites!
		if !first {
			fn, err = next(dir, fn, 1)
			if err != nil {
				log.Println(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		first = false
		fp := filepath.Join(dir, fn)
		watches.ignore(fp)
		err = backup(fp)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Println("Creating", fp)
		dst, err := os.Create(fp)
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
			log.Println("Encoded", to, fp)
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
				err := os.Remove(fp)
				if err != nil {
					log.Println(err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				log.Println("Deleted", fp)
			} else {
				log.Println("Copied", fp)
			}
		}
		data.Add("actual", fn)
		username, _, ok := r.BasicAuth()
		if ok {
			log.Println("Saved", filepath.ToSlash(fp), "by", username)
		} else {
			log.Println("Saved", filepath.ToSlash(fp))
		}
		updateTemplate(fp)
	}
	data.Set("last", fn) // has no slashes
	http.Redirect(w, r, "/upload/" + nameEscape(dir) + "?" + data.Encode(), http.StatusFound)
}

// Base returns a page name matching the first uploaded file: no extension and no appended number. If the name refers to
// a directory, returns "index".
func (u *Upload) Base() string {
	s := u.Name
	n := s[:strings.LastIndex(s, ".")]
	m := baseRe.FindStringSubmatch(n)
	if m != nil {
		return m[1]
	}
	if n == "." {
		return "index"
	}
	return n
}

// PagePath returns the Upload.Base(), percent-escaped except for the slashes.
func (u *Upload) PagePath() string {
	s := u.Name
	n := s[:strings.LastIndex(s, ".")]
	m := baseRe.FindStringSubmatch(n)
	if m != nil {
		return pathEncode(m[1])
	}
	if n == "." {
		return "index"
	}
	return pathEncode(n)
}

// Title returns the title of the matching page, if it exists.
func (u *Upload) Title() string {
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
func (u *Upload) Today() string {
	return time.Now().Format(time.DateOnly)
}

// LastPath returns the LastName with some characters escaped because html/template doesn't escape those. This is
// suitable for use in HTML templates.
func (u *Upload) LastPath() string {
	return pathEncode(u.Last)
}
