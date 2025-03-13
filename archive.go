package main

import (
	"archive/zip"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// archiveHandler serves a zip file. Directories starting with a period are skipped. Filenames starting with a period
// are skipped. If the environment variable ODDMU_FILTER is a regular expression that matches the starting directory,
// this is a "separate site"; if the regular expression does not match, this is the "main site" and page names must also
// not match the regular expression.
func archiveHandler(w http.ResponseWriter, r *http.Request, name string) {
	filter := os.Getenv("ODDMU_FILTER")
	re, err := regexp.Compile(filter)
	if err != nil {
		log.Println("ODDMU_FILTER does not compile:", filter, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	matches := re.MatchString(name)
	dir := filepath.Dir(filepath.FromSlash(name))
	z := zip.NewWriter(w)
	err = filepath.Walk(dir, func(fp string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if fp != "." && strings.HasPrefix(filepath.Base(fp), ".") {
				return filepath.SkipDir
			}
		} else if !strings.HasPrefix(filepath.Base(fp), ".") &&
			(matches || !re.MatchString(filepath.ToSlash(fp))) {
			zf, err := z.Create(fp)
			if err != nil {
				log.Println(err)
				return err
			}
			f, err := os.Open(fp)
			if err != nil {
				log.Println(err)
				return err
			}
			_, err = io.Copy(zf, f)
			if err != nil {
				log.Println(err)
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = z.Close()
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
