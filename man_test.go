package main

import (
	"github.com/stretchr/testify/assert"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"testing"
)

// Does oddmu(1) link to all the other man pages?
func TestManPages(t *testing.T) {
	b, err := os.ReadFile("man/oddmu.1.txt")
	main := string(b)
	assert.NoError(t, err)
	count := 0
	filepath.Walk("man", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".txt") &&
			path != "man/oddmu.1.txt" {
			count++
			s := strings.TrimPrefix(path, "man/")
			s = strings.TrimSuffix(s, ".txt")
			i := strings.LastIndex(s, ".")
			ref := "_" + s[:i] + "_(" + s[i+1:] + ")"
			assert.Contains(t, main, ref, ref)
		}
		return nil
	})
	assert.Greater(t, count, 0, "no man pages were found")
}

// Does oddmu-templates(5) mention all the templates?
func TestManTemplates(t *testing.T) {
	b, err := os.ReadFile("man/oddmu-templates.5.txt")
	man := string(b)
	assert.NoError(t, err)
	count := 0
	filepath.Walk(".", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".html") {
			count++
			assert.Contains(t, man, path, path)
		}
		if path != "." && info.IsDir() {
			return filepath.SkipDir
		}
		return nil
	})
	assert.Greater(t, count, 0, "no templates were found")
}

// Does oddmu(1) mention all the actions? We're not going to parse the go file and make sure to catch them all. I tried
// it, and it's convoluted.
func TestManActions(t *testing.T) {
	b, err := os.ReadFile("man/oddmu.1.txt")
	assert.NoError(t, err)
	main := string(b)
	b, err = os.ReadFile("wiki.go")
	assert.NoError(t, err)
	wiki := string(b)
	count := 0
	// this doesn't match the root handler
	re := regexp.MustCompile(`http.HandleFunc\("(/[a-z]+/)", makeHandler\([a-z]+Handler, (true|false)\)\)`)
	for _, match := range re.FindAllStringSubmatch(wiki, -1) {
		count++
		var path string
		if match[2] == "true" {
			path = "_" + match[1] + "dir/name"
		} else {
			path = "_" + match[1] + "dir/"
		}
		assert.Contains(t, main, path, path)
	}
	assert.Greater(t, count, 0, "no handlers were found")
	// root handler is manual
	assert.Contains(t, main, "\n- _/_", "root")
}

// Does the README link to all the man pages and all the Go source files,
// excluding the command and test files?
func TestReadme(t *testing.T) {
	b, err := os.ReadFile("README.md")
	readme := string(b)
	assert.NoError(t, err)
	count := 0
	filepath.Walk("man", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".txt") {
			count++
			s := strings.TrimPrefix(path, "man/")
			s = strings.TrimSuffix(s, ".txt")
			i := strings.LastIndex(s, ".")
			ref := "[" + s[:i] + "(" + s[i+1:] + ")]"
			assert.Contains(t, readme, ref, ref)
		}
		return nil
	})
	assert.Greater(t, count, 0, "no man pages were found")
	count = 0
	filepath.Walk(".", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".go") &&
			!strings.HasSuffix(path, "_test.go") &&
			!strings.HasSuffix(path, "_cmd.go") {
			count++
			s := strings.TrimPrefix(path, "./")
			ref := "`" + s + "`"
			assert.Contains(t, readme, ref, ref)
		}
		return nil
	})
	assert.Greater(t, count, 0, "no source pages were found")
}

// Does the README document all the dependecies, checking all the all the packages with names containing a period?
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
				name := imp.Path.Value[1 : len(imp.Path.Value)-1]
				if strings.Contains(name, ".") && !slices.Contains(imports, name) {
					imports = append(imports, name)
				}
			}
		}
	}
	assert.Greater(t, len(imports), 0, "no imports found")
	sort.Slice(imports, func(i, j int) bool { return len(imports[i]) < len(imports[j]) })
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
