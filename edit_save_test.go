package main

import (
	"github.com/stretchr/testify/assert"
	"net/url"
	"os"
	"regexp"
	"testing"
)

// wipes testdata
func TestEditSave(t *testing.T) {
	_ = os.RemoveAll("testdata")

	data := url.Values{}
	data.Set("body", "Hallo!")

	HTTPRedirectTo(t, makeHandler(viewHandler, true), "GET", "/view/testdata/alex", nil, "/edit/testdata/alex")
	assert.HTTPStatusCode(t, makeHandler(editHandler, true), "GET", "/edit/testdata/alex", nil, 200)
	HTTPRedirectTo(t, makeHandler(saveHandler, true), "POST", "/save/testdata/alex", data, "/view/testdata/alex")
	assert.Regexp(t, regexp.MustCompile("Hallo!"),
		assert.HTTPBody(makeHandler(viewHandler, true), "GET", "/view/testdata/alex", nil))

	t.Cleanup(func() {
		_ = os.RemoveAll("testdata")
	})
}
