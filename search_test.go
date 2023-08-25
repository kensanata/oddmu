package main

import (
	"os"
	"strings"
	"testing"
)

var name string = "test"

// TestIndex relies on README.md being indexed
func TestIndex(t *testing.T) {
	_ = os.Remove(name + ".md")
	loadIndex()
	q := "Oddµ"
	pages := search(q)
	if len(pages) == 0 {
		t.Log("Search found no result")
		t.Fail()
	}
	for _, p := range pages {
		if !strings.Contains(string(p.Body), q) && !strings.Contains(string(p.Title), q) {
			t.Logf("Page %s does not contain %s", p.Name, q)
			t.Fail()
		}
		if p.Score == 0 {
			t.Logf("Page %s has no score", p.Name)
			t.Fail()
		}
	}
	p := &Page{Name: name, Body: []byte("This is a test.")}
	p.save()
	pages = search("This is a test")
	found := false
	for _, p := range pages {
		if p.Name == name {
			found = true
			break
		}
	}
	if !found {
		t.Logf("Page '%s' was not found", name)
		t.Fail()
	}
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
	if found {
		t.Logf("Page '%s' was still found using the old content: %s", name, p.Body)
		t.Fail()
	}
	pages = search("Guvf")
	found = false
	for _, p := range pages {
		if p.Name == name {
			found = true
			break
		}
	}
	if !found {
		t.Logf("Page '%s' not found using the new content: %s", name, p.Body)
		t.Fail()
	}
	t.Cleanup(func() {
		_ = os.Remove(name + ".md")
	})
}
