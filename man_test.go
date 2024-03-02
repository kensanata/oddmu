package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sort"
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
	filepath.Walk(".", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".go") &&
			!strings.HasSuffix(path, "_test.go") &&
			!strings.HasSuffix(path, "_cmd.go") {
			s := strings.TrimPrefix(path, "./")
			ref := "`" + s + "`"
			assert.Contains(t, main, ref, ref)
		}
		return nil
	})
}

func TestDocumentDependencies(t *testing.T) {
	b, err := os.ReadFile("README.md")
	readme := string(b)
	assert.NoError(t, err)
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", nil, parser.ImportsOnly)
	assert.NoError(t, err)
	imports := []string{}
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, imp := range file.Imports {
				name := imp.Path.Value[1:len(imp.Path.Value)-1]
				if strings.Contains(name, ".") && !slices.Contains(imports, name) {
					imports = append(imports, name)
				}
			}
		}
	}
	sort.Slice(imports, func(i, j int) bool { return len(imports[i]) < len(imports[j]) })
	fmt.Println(imports)
IMPORT:
	for _, name := range imports {
		for _, other := range imports {
			if strings.HasPrefix(name, other) && name != other {
				continue IMPORT
			}
		}
		ok := strings.Contains(readme, name)
		assert.True(t, ok, name)
	}
}
