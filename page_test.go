package main

import (
	"github.com/stretchr/testify/assert"
	"os"
	"regexp"
	"testing"
)

func TestPageTitle(t *testing.T) {
	p := &Page{Body: []byte(`# Ache
My back aches for you
I sit, stare and type for hours
But yearn for blue sky`)}
	p.handleTitle(false)
	assert.Equal(t, "Ache", p.Title)
	assert.Regexp(t, regexp.MustCompile("^# Ache"), string(p.Body))
	p.handleTitle(true)
	assert.Regexp(t, regexp.MustCompile("^My back"), string(p.Body))
}

// wipes testdata
func TestPageDir(t *testing.T) {
	_ = os.RemoveAll("testdata")
	index.load()
	p := &Page{Name: "testdata/moon", Body: []byte(`# Moon
From bed to bathroom
A slow shuffle in the dark
Moonlight floods the aisle`)}
	p.save()

	o, err := loadPage("testdata/moon")
	assert.NoError(t, err, "load page")
	assert.Equal(t, p.Body, o.Body)
	assert.FileExists(t, "testdata/moon.md")

	// Saving an empty page deletes it.
	p = &Page{Name: "testdata/moon", Body: []byte("")}
	p.save()
	assert.NoFileExists(t, "testdata/moon.md")

	// But the backup still exists.
	assert.FileExists(t, "testdata/moon.md~")

	t.Cleanup(func() {
		_ = os.RemoveAll("testdata")
	})
}
