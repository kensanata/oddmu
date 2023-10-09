package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDiff(t *testing.T) {
	cleanup(t, "testdata/diff")
	index.load()
	s := `#Bread

The oven breathes
Fills us with the thought of bread
Oh so fresh, so warm.`
	r := `#Bread

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
