package main

import (
	"github.com/stretchr/testify/assert"
	"net/url"
	"os"
	"regexp"
	"testing"
)

// wipes testdata
func TestAddAppend(t *testing.T) {
	_ = os.RemoveAll("testdata")
	index.load()

	p := &Page{Name: "testdata/fire", Body: []byte(`# Fire
Orange sky above
Reflects a distant fire
It's not `)}
	p.save()

	data := url.Values{}
	data.Set("body", "barbecue")

	assert.Regexp(t, regexp.MustCompile("a distant fire"),
		assert.HTTPBody(makeHandler(viewHandler, true), "GET", "/view/testdata/fire", nil))
	assert.NotRegexp(t, regexp.MustCompile("a distant fire"),
		assert.HTTPBody(makeHandler(addHandler, true), "GET", "/add/testdata/fire", nil))
	HTTPRedirectTo(t, makeHandler(appendHandler, true), "POST", "/append/testdata/fire", data, "/view/testdata/fire")
	assert.Regexp(t, regexp.MustCompile("It’s not barbecue"),
		assert.HTTPBody(makeHandler(viewHandler, true), "GET", "/view/testdata/fire", nil))
	t.Cleanup(func() {
		_ = os.RemoveAll("testdata")
	})
}
