package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// Use go test -race to see whether this is a race condition.
func TestLoadAndSearch(t *testing.T) {
	go loadIndex()
	q := "Oddµ"
	pages := search(q)
	assert.Zero(t, len(pages))
}
