package main

import (
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
	"os"
)

func TestSearch(t *testing.T) {
	data := url.Values{}
	data.Set("q", "oddµ")
	assert.Contains(t,
		assert.HTTPBody(searchHandler, "GET", "/search", data), "Welcome")
}

// wipes testdata
func TestSearchQuestionmark(t *testing.T) {
	_ = os.RemoveAll("testdata")
	p := &Page{Name: "testdata/Odd?", Body: []byte(`# Even?

yes or no?`)}
	p.save()
	data := url.Values{}
	data.Set("q", "yes")
	body := assert.HTTPBody(searchHandler, "GET", "/search", data)
	assert.Contains(t, body, "yes or no?")
	assert.NotContains(t, body, "Odd?")
	assert.Contains(t, body, "Even?")
}
