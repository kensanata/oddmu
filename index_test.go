package main

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

// TestIndex relies on README.md being indexed
func TestIndex(t *testing.T) {
	index.load()
	q := "Oddµ"
	pages, _ := search(q, 1)
	assert.NotZero(t, len(pages))
	for _, p := range pages {
		assert.NotContains(t, p.Title, "<b>")
		assert.True(t, strings.Contains(string(p.Body), q) || strings.Contains(string(p.Title), q))
		assert.NotZero(t, p.Score, "Score %d for %s", p.Score, p.Name)
	}
}

func TestSearchHashtag(t *testing.T) {
	index.load()
	q := "#like_this"
	pages, _ := search(q, 1)
	assert.NotZero(t, len(pages))
}

func TestIndexUpdates(t *testing.T) {
	cleanup(t, "testdata/update")
	name := "testdata/update/test"
	index.load()
	p := &Page{Name: name, Body: []byte("This is a test.")}
	p.save()

	// Find the phrase
	pages, _ := search("This is a test", 1)
	found := false
	for _, p := range pages {
		if p.Name == name {
			found = true
			break
		}
	}
	assert.True(t, found)

	// Find the phrase, case insensitive
	pages, _ = search("this is a test", 1)
	found = false
	for _, p := range pages {
		if p.Name == name {
			found = true
			break
		}
	}
	assert.True(t, found)

	// Find some words
	pages, _ = search("this test", 1)
	found = false
	for _, p := range pages {
		if p.Name == name {
			found = true
			break
		}
	}
	assert.True(t, found)

	// Update the page and no longer find it with the old phrase
	p = &Page{Name: name, Body: []byte("Guvf vf n grfg.")}
	p.save()
	pages, _ = search("This is a test", 1)
	found = false
	for _, p := range pages {
		if p.Name == name {
			found = true
			break
		}
	}
	assert.False(t, found)

	// Find page using a new word
	pages, _ = search("Guvf", 1)
	found = false
	for _, p := range pages {
		if p.Name == name {
			found = true
			break
		}
	}
	assert.True(t, found)
}
