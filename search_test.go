package main

import (
	"github.com/stretchr/testify/assert"
	"net/url"
	"os"
	"testing"
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

We look at the plants.
They need water. We need us.
The silence streches.`)}
	p.save()
	data := url.Values{}
	data.Set("q", "look")
	body := assert.HTTPBody(searchHandler, "GET", "/search", data)
	assert.Contains(t, body, "We <b>look</b>")
	assert.NotContains(t, body, "Odd?")
	assert.Contains(t, body, "Even?")
}
