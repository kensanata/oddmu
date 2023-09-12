package main

import (
	"os"
	"strings"
	"testing"
	"github.com/stretchr/testify/assert"
)

// TestIndex relies on README.md being indexed
func TestIndex(t *testing.T) {
	loadIndex()
	q := "Oddµ"
	pages := search(q)
	assert.NotZero(t, len(pages))
	for _, p := range pages {
		assert.NotContains(t, p.Title, "<b>");
		assert.True(t, strings.Contains(string(p.Body), q) || strings.Contains(string(p.Title), q))
		assert.NotZero(t, p.Score)
	}
}

func TestSearchHashtag (t *testing.T) {
	loadIndex()
	q := "#Another_Tag"
	pages := search(q)
	assert.NotZero(t, len(pages))
}

func TestIndexUpdates(t *testing.T) {
	name := "test"
	_ = os.Remove(name + ".md")
	loadIndex()
	p := &Page{Name: name, Body: []byte("This is a test.")}
	p.save()

	// Find the phrase
	pages := search("This is a test")
	found := false
	for _, p := range pages {
		if p.Name == name {
			found = true
			break
		}
	}
	assert.True(t, found)

	// Find the phrase, case insensitive
	pages = search("this is a test")
	found = false
	for _, p := range pages {
		if p.Name == name {
			found = true
			break
		}
	}
	assert.True(t, found)

	// Find some words
	pages = search("this test")
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
	pages = search("This is a test")
	found = false
	for _, p := range pages {
		if p.Name == name {
			found = true
			break
		}
	}
	assert.False(t, found)

	// Find page using a new word
	pages = search("Guvf")
	found = false
	for _, p := range pages {
		if p.Name == name {
			found = true
			break
		}
	}
	assert.True(t, found)

	t.Cleanup(func() {
		_ = os.Remove(name + ".md")
	})
}
