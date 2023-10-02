package main

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"image"
	"image/jpeg"
	"image/png"
	"mime/multipart"
	"os"
	"testing"
	"net/url"
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
		writer.FormDataContentType(), form, "/upload/testdata/files?last=ok.txt")
	assert.Contains(t,
		assert.HTTPBody(makeHandler(viewHandler, true), "GET", "/view/testdata/files/ok.txt", nil),
		"Hello!")
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
		writer.FormDataContentType(), form, "/upload/testdata/png?last=ok.png")
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
		writer.FormDataContentType(), form, "/upload/testdata/jpg?last=ok.jpg")
}

func TestUploadMultiple(t *testing.T) {
	cleanup(t, "testdata/multi")
	// for uploads, the directory is not created automatically
	os.MkdirAll("testdata/multi", 0755)
	form := new(bytes.Buffer)
	writer := multipart.NewWriter(form)
	field, _ := writer.CreateFormField("name")
	field.Write([]byte("2023-10-02-hike-1.jpg"))
	field, _ = writer.CreateFormField("maxwidth")
	field.Write([]byte("15"))
	field, _ = writer.CreateFormField("quality")
	field.Write([]byte("50"))
	file, _ := writer.CreateFormFile("file", "ok.jpg")
	img := image.NewRGBA(image.Rect(0, 0, 20, 20))
	jpeg.Encode(file, img, &jpeg.Options{Quality: 90})
	writer.Close()
	location := HTTPUploadLocation(t, makeHandler(dropHandler, false), "/drop/testdata/multi/",
		writer.FormDataContentType(), form)
	url, _ := url.Parse(location)
	assert.Equal(t, url.Path, "/upload/testdata/multi", "Redirect to upload location")
	values := url.Query()
	assert.Equal(t, "2023-10-02-hike-1.jpg", values.Get("last"))
	assert.Equal(t, "15", values.Get("maxwidth"))
	assert.Equal(t, "50", values.Get("quality"))
	assert.Equal(t, "1", values.Get("image"))
	body := assert.HTTPBody(makeHandler(uploadHandler, true), "GET", url.Path, values)
	assert.Contains(t, body, `value="2023-10-02-hike-2.jpg"`)
	assert.Contains(t, body, `value="15"`)
	assert.Contains(t, body, `value="50"`)
	assert.Contains(t, body, `src="/view/testdata/multi/2023-10-02-hike-1.jpg"`)
}
