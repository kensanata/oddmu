package main

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"mime/multipart"
	"testing"
)

func TestTemplates(t *testing.T) {
	cleanup(t, "testdata/templates")
	// save a file to create the directory
	p := &Page{Name: "testdata/templates/snow", Body: []byte(`# Snow

A blob on the grass
Covered in needles and dust
Memories of cold
`)}
	p.save()
	assert.Contains(t,
		assert.HTTPBody(makeHandler(viewHandler, true), "GET", "/view/testdata/templates/snow", nil),
		"Skip navigation")
	// save a new view handler
	html := "<body><h1>{{.Title}}</h1>{{.Html}}"
	form := new(bytes.Buffer)
	writer := multipart.NewWriter(form)
	field, _ := writer.CreateFormField("name")
	field.Write([]byte("view.html"))
	file, _ := writer.CreateFormFile("file", "test.html")
	file.Write([]byte(html))
	writer.Close()
	HTTPUploadLocation(t, makeHandler(dropHandler, false), "/drop/testdata/templates/", writer.FormDataContentType(), form)
	assert.Contains(t,
		assert.HTTPBody(makeHandler(viewHandler, false), "GET", "/view/testdata/templates/view.html", nil),
		html)
	// verify that it works
	body := assert.HTTPBody(makeHandler(viewHandler, true), "GET", "/view/testdata/templates/snow", nil)
	assert.Contains(t, body, "<h1>Snow</h1>")
	assert.NotContains(t, body, "Skip")
	// verify that the top level still uses the old template
	assert.Contains(t,
		assert.HTTPBody(makeHandler(viewHandler, true), "GET", "/view/index", nil),
		"Skip navigation")
}
