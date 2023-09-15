package main

import (
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
)

func TestSearch(t *testing.T) {
	data := url.Values{}
	data.Set("q", "oddµ")
	assert.Contains(t,
		assert.HTTPBody(searchHandler, "GET", "/search", data), "Welcome")
}
