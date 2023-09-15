package main

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"mime/multipart"
	"os"
	"regexp"
	"testing"
)

// wipes testdata
func TestUpload(t *testing.T) {
	_ = os.RemoveAll("testdata")
	// for uploads, the directory is not created automatically
	os.MkdirAll("testdata", 0755)
	assert.HTTPStatusCode(t, makeHandler(uploadHandler, false), "GET", "/upload/", nil, 200)
	form := new(bytes.Buffer)
	writer := multipart.NewWriter(form)
	field, err := writer.CreateFormField("name")
	assert.NoError(t, err)
	_, err = field.Write([]byte("testdata/ok.txt"))
	assert.NoError(t, err)
	file, err := writer.CreateFormFile("file", "example.txt");
	assert.NoError(t, err)
	file.Write([]byte("Hello!"))
	err = writer.Close()
	assert.NoError(t, err)
	HTTPUploadAndRedirectTo(t, makeHandler(dropHandler, false), "/drop/",
		writer.FormDataContentType(), form, "/view/testdata/ok.txt")
	assert.Regexp(t, regexp.MustCompile("Hello!"),
		assert.HTTPBody(makeHandler(viewHandler, true), "GET", "/view/testdata/ok.txt", nil))
	t.Cleanup(func() {
		_ = os.RemoveAll("testdata")
	})
}
