package main

import (
	"github.com/stretchr/testify/assert"
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

func TestPageDir(t *testing.T) {
	cleanup(t, "testdata/dir")
	index.load()
	p := &Page{Name: "testdata/dir/moon", Body: []byte(`# Moon
From bed to bathroom
A slow shuffle in the dark
Moonlight floods the aisle`)}
	p.save()

	o, err := loadPage("testdata/dir/moon")
	assert.NoError(t, err, "load page")
	assert.Equal(t, p.Body, o.Body)
	assert.FileExists(t, "testdata/dir/moon.md")

	// Saving an empty page deletes it.
	p = &Page{Name: "testdata/dir/moon", Body: []byte("")}
	p.save()
	assert.NoFileExists(t, "testdata/dir/moon.md")

	// But the backup still exists.
	assert.FileExists(t, "testdata/dir/moon.md~")
}
