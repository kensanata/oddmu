package main

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestIndexAdd(t *testing.T) {
	idx := &indexStore{}
	idx.reset()
	idx.Lock()
	defer idx.Unlock()
	tag := "hello"
	id := idx.addDocument([]byte("oh hi #" + tag))
	assert.Contains(t, idx.token, tag)
	idx.deleteDocument(id)
	assert.NotContains(t, idx.token, tag)
}

// TestIndex relies on README.md being indexed
func TestIndex(t *testing.T) {
	index.load()
	q := "Oddμ"
	pages, _ := search(q, "", "", 1, false)
	assert.NotZero(t, len(pages))
	for _, p := range pages {
		assert.NotContains(t, p.Title, "<b>")
		assert.True(t, strings.Contains(string(p.Body), q) || strings.Contains(string(p.Title), q))
		assert.NotZero(t, p.Score, "Score %d for %s", p.Score, p.Name)
	}
}

// Lower case hashtag!
func TestSearchHashtag(t *testing.T) {
	cleanup(t, "testdata/search-hashtag")
	p := &Page{Name: "testdata/search-hashtag/search", Body: []byte(`# Search

I'm back in this room
Shelf, table, chair, and shelf again
Where are my glasses?

#Searching`)}
	p.save()
	index.load()
	pages, _ := search("#searching", "", "", 1, false)
	assert.NotZero(t, len(pages))
}

func TestIndexUpdates(t *testing.T) {
	cleanup(t, "testdata/update")
	name := "testdata/update/test"
	index.load()
	p := &Page{Name: name, Body: []byte("#Old Name\nThis is a test.")}
	p.save()

	// Find the phrase
	pages, _ := search("This is a test", "", "", 1, false)
	found := false
	for _, p := range pages {
		if p.Name == name {
			found = true
			break
		}
	}
	assert.True(t, found)

	// Find the phrase, case insensitive
	pages, _ = search("this is a test", "", "", 1, false)
	found = false
	for _, p := range pages {
		if p.Name == name {
			found = true
			break
		}
	}
	assert.True(t, found)

	// Find some words
	pages, _ = search("this test", "", "", 1, false)
	found = false
	for _, p := range pages {
		if p.Name == name {
			found = true
			break
		}
	}
	assert.True(t, found)

	// Update the page and no longer find it with the old phrase
	p = &Page{Name: name, Body: []byte("# New page\nGuvf vf n grfg.")}
	p.save()
	pages, _ = search("This is a test", "", "", 1, false)
	found = false
	for _, p := range pages {
		if p.Name == name {
			found = true
			break
		}
	}
	assert.False(t, found)

	// Find page using a new word
	pages, _ = search("Guvf", "", "", 1, false)
	found = false
	for _, p := range pages {
		if p.Name == name {
			found = true
			break
		}
	}
	assert.True(t, found)

	// Make sure the title was updated
	index.RLock()
	defer index.RUnlock()
	assert.Equal(t, "New page", index.titles[name])
}
