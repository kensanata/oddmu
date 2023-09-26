package main

import (
	"github.com/stretchr/testify/assert"
	"net/url"
	"regexp"
	"testing"
)

func TestAddAppend(t *testing.T) {
	cleanup(t, "testdata/add")
	index.load()

	p := &Page{Name: "testdata/add/fire", Body: []byte(`# Fire
Orange sky above
Reflects a distant fire
It's not `)}
	p.save()

	data := url.Values{}
	data.Set("body", "barbecue")

	assert.Regexp(t, regexp.MustCompile("a distant fire"),
		assert.HTTPBody(makeHandler(viewHandler, true),
			"GET", "/view/testdata/add/fire", nil))
	assert.NotRegexp(t, regexp.MustCompile("a distant fire"),
		assert.HTTPBody(makeHandler(addHandler, true),
			"GET", "/add/testdata/add/fire", nil))
	HTTPRedirectTo(t, makeHandler(appendHandler, true),
		"POST", "/append/testdata/add/fire", data, "/view/testdata/add/fire")
	assert.Regexp(t, regexp.MustCompile("It’s not barbecue"),
		assert.HTTPBody(makeHandler(viewHandler, true),
			"GET", "/view/testdata/add/fire", nil))
}
