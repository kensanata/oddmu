package main

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"image"
	"image/jpeg"
	"image/png"
	"mime/multipart"
	"os"
	"regexp"
	"testing"
)

func TestUpload(t *testing.T) {
	cleanup(t, "testdata/files")
	// for uploads, the directory is not created automatically
	os.MkdirAll("testdata/files", 0755)
	assert.HTTPStatusCode(t, makeHandler(uploadHandler, false), "GET", "/upload/testdata/files/", nil, 200)
	form := new(bytes.Buffer)
	writer := multipart.NewWriter(form)
	field, err := writer.CreateFormField("name")
	assert.NoError(t, err)
	_, err = field.Write([]byte("ok.txt"))
	assert.NoError(t, err)
	file, err := writer.CreateFormFile("file", "example.txt")
	assert.NoError(t, err)
	file.Write([]byte("Hello!"))
	err = writer.Close()
	assert.NoError(t, err)
	HTTPUploadAndRedirectTo(t, makeHandler(dropHandler, false), "/drop/testdata/files/",
		writer.FormDataContentType(), form, "/view/testdata/files/ok.txt")
	assert.Regexp(t, regexp.MustCompile("Hello!"),
		assert.HTTPBody(makeHandler(viewHandler, true), "GET", "/view/testdata/files/ok.txt", nil))
}

func TestUploadPng(t *testing.T) {
	cleanup(t, "testdata/png")
	// for uploads, the directory is not created automatically
	os.MkdirAll("testdata/png", 0755)
	form := new(bytes.Buffer)
	writer := multipart.NewWriter(form)
	field, _ := writer.CreateFormField("name")
	field.Write([]byte("ok.png"))
	file, _ := writer.CreateFormFile("file", "ok.png")
	img := image.NewRGBA(image.Rect(0, 0, 20, 20))
	png.Encode(file, img)
	writer.Close()
	HTTPUploadAndRedirectTo(t, makeHandler(dropHandler, false), "/drop/testdata/png/",
		writer.FormDataContentType(), form, "/view/testdata/png/ok.png")
}

func TestUploadJpg(t *testing.T) {
	cleanup(t, "testdata/jpg")
	// for uploads, the directory is not created automatically
	os.MkdirAll("testdata/jpg", 0755)
	form := new(bytes.Buffer)
	writer := multipart.NewWriter(form)
	field, _ := writer.CreateFormField("name")
	field.Write([]byte("ok.jpg"))
	file, _ := writer.CreateFormFile("file", "ok.jpg")
	img := image.NewRGBA(image.Rect(0, 0, 20, 20))
	jpeg.Encode(file, img, &jpeg.Options{Quality: 90})
	writer.Close()
	HTTPUploadAndRedirectTo(t, makeHandler(dropHandler, false), "/drop/testdata/jpg/",
		writer.FormDataContentType(), form, "/view/testdata/jpg/ok.jpg")
}
