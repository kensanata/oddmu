package main

import (
	"github.com/stretchr/testify/assert"
	"net/url"
	"regexp"
	"testing"
)

func TestEditSave(t *testing.T) {
	cleanup(t, "testdata/save")

	data := url.Values{}
	data.Set("body", "Hallo!")

	HTTPRedirectTo(t, makeHandler(viewHandler, true),
		"GET", "/view/testdata/save/alex", nil, "/edit/testdata/save/alex")
	assert.HTTPStatusCode(t, makeHandler(editHandler, true),
		"GET", "/edit/testdata/save/alex", nil, 200)
	HTTPRedirectTo(t, makeHandler(saveHandler, true),
		"POST", "/save/testdata/save/alex", data, "/view/testdata/save/alex")
	assert.Regexp(t, regexp.MustCompile("Hallo!"),
		assert.HTTPBody(makeHandler(viewHandler, true),
			"GET", "/view/testdata/save/alex", nil))
}
