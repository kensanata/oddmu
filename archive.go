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
func archiveHandler(w http.ResponseWriter, r *http.Request, path string) {
	filter := os.Getenv("ODDMU_FILTER")
	re, err := regexp.Compile(filter)
	if err != nil {
		log.Println("ODDMU_FILTER does not compile:", filter, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	matches := re.MatchString(path)
	dir := filepath.Dir(filepath.FromSlash(path))
	z := zip.NewWriter(w)
	err = filepath.Walk(dir, func (path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				if path != "." && strings.HasPrefix(filepath.Base(path), ".") {
					return filepath.SkipDir
				}
			} else if !strings.HasPrefix(filepath.Base(path), ".") &&
				(matches || !re.MatchString(path)) {
				zf, err := z.Create(path)
				if err != nil {
					log.Println(err)
					return err
				}
				file, err := os.Open(path)
				if err != nil {
					log.Println(err)
					return err
				}
				_, err = io.Copy(zf, file)
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
