package main

import (
	"archive/zip"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

func TestArchive(t *testing.T) {
	cleanup(t, "testdata/archive")
	assert.NoError(t, os.MkdirAll("testdata/archive/public", 0755))
	assert.NoError(t, os.MkdirAll("testdata/archive/secret", 0755))
	assert.NoError(t, os.WriteFile("testdata/archive/public/index.md", []byte("# Public\nChurch tower bells ringing\nA cold wind biting my ears\nWalk across the square"), 0644))
	assert.NoError(t, os.WriteFile("testdata/archive/secret/index.md", []byte("# Secret\nMany years ago I danced\nSpending nights in clubs and bars\nIt is my secret"), 0644))
	os.Setenv("ODDMU_FILTER", "^testdata/archive/secret/")
	body := assert.HTTPBody(makeHandler(archiveHandler, true), "GET", "/archive/testdata/data.zip", nil)
	r, err := zip.NewReader(strings.NewReader(body), int64(len(body)))
	assert.NoError(t, err, "Unzip")
	names := []string{}
	for _, file := range r.File {
		names = append(names, file.Name)
	}
	assert.Contains(t, names, "testdata/archive/public/index.md")
	assert.NotContains(t, names, "testdata/archive/secret/index.md")
}
