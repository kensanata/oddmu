package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// Use go test -race to see whether this is a race condition.
func TestLoadAndSearch(t *testing.T) {
	index.reset()
	go index.load()
	q := "Oddµ"
	pages, _ := search(q, 1)
	assert.Zero(t, len(pages))
}
