package main

import (
	"github.com/stretchr/testify/assert"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestManPages(t *testing.T) {
	b, err := os.ReadFile("man/oddmu.1.txt")
	main := string(b)
	assert.NoError(t, err)
	filepath.Walk("man", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".txt") &&
			path != "man/oddmu.1.txt" {
			s := strings.TrimPrefix(path, "man/")
			s = strings.TrimSuffix(s, ".txt")
			i := strings.LastIndex(s, ".")
			ref := "_" + s[:i] + "_(" + s[i+1:] + ")"
			assert.Contains(t, main, ref, ref)
		}
		return nil
	})
}

func TestReadme(t *testing.T) {
	b, err := os.ReadFile("README.md")
	main := string(b)
	assert.NoError(t, err)
	filepath.Walk("man", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".txt") {
			s := strings.TrimPrefix(path, "man/")
			s = strings.TrimSuffix(s, ".txt")
			i := strings.LastIndex(s, ".")
			ref := "[" + s[:i] + "(" + s[i+1:] + ")]"
			assert.Contains(t, main, ref, ref)
		}
		return nil
	})
}
