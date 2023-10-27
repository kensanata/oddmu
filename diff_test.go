package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDiff(t *testing.T) {
	cleanup(t, "testdata/diff")
	index.load()
	s := `# Bread

The oven breathes
Fills us with the thought of bread
Oh so fresh, so warm.`
	r := `# Bread

The oven whispers
Fills us with the thought of bread
Oh so fresh, so warm.`
	p := &Page{Name: "testdata/diff/bread", Body: []byte(s)}
	p.save()
	p.Body = []byte(r)
	p.save()
	body := assert.HTTPBody(makeHandler(diffHandler, true),
		"GET", "/diff/testdata/diff/bread", nil)
	assert.Contains(t, body, `<del>breathe</del>`)
	assert.Contains(t, body, `<ins>whisper</ins>`)
}

func TestDiffPercentEncoded(t *testing.T) {
	cleanup(t, "testdata/diff")
	index.load()
	s := `# Coup de Gras

Playing D&D
We talk about a killing
Mispronouncing words`
	r := `# Coup de Grace

Playing D&D
We talk about a killing
Mispronouncing words`
	p := &Page{Name: "testdata/diff/coup de grace", Body: []byte(s)}
	p.save()
	p.Body = []byte(r)
	p.save()
	body := assert.HTTPBody(makeHandler(diffHandler, true),
		"GET", "/diff/testdata/diff/coup%20de%20grace", nil)
	assert.Contains(t, body, `<del>s</del>`)
	assert.Contains(t, body, `<ins>ce</ins>`)
}
