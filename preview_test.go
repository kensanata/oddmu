package main

import (
	"github.com/stretchr/testify/assert"
	"net/url"
	"net/http"
	"testing"
)

func TestPreview(t *testing.T) {
	cleanup(t, "testdata/preview")

	data := url.Values{}
	data.Set("body", "**Hallo**!")

	r := assert.HTTPBody(makeHandler(previewHandler, false, http.MethodGet), "POST", "/view/testdata/preview/alex", data)
	assert.Contains(t, r, "<strong>Hallo</strong>!")
}
