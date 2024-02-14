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

func archiveHandler(w http.ResponseWriter, r *http.Request, path string) {
	filter := os.Getenv("ODDMU_FILTER")
	re, err := regexp.Compile(filter)
	if err != nil {
		log.Println("ODDMU_FILTER does not compile:", filter, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	mustMatch := re.MatchString(path)
	z := zip.NewWriter(w)
	err = filepath.Walk(filepath.Dir(filepath.FromSlash(path)),
		func (path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				// skip hidden directories
				if path != "." && strings.HasPrefix(filepath.Base(path), ".") {
					return filepath.SkipDir
				}
			} else if !strings.HasPrefix(filepath.Base(path), ".") &&
				// skip filtered files
				(mustMatch && re.MatchString(path) ||
					!mustMatch && !re.MatchString(path)) {
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
