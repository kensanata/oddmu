package main

import (
	"bytes"
	"encoding/base64"
	"github.com/stretchr/testify/assert"
	"image"
	"image/jpeg"
	"image/png"
	"mime/multipart"
	"net/url"
	"os"
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
	_, err = file.Write([]byte("Hello!"))
	assert.NoError(t, err)
	err = writer.Close()
	assert.NoError(t, err)
	HTTPUploadAndRedirectTo(t, makeHandler(dropHandler, false), "/drop/testdata/files/",
		writer.FormDataContentType(), form, "/upload/testdata/files/?last=ok.txt")
	assert.Contains(t,
		assert.HTTPBody(makeHandler(viewHandler, false), "GET", "/view/testdata/files/ok.txt", nil),
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
		writer.FormDataContentType(), form, "/upload/testdata/png/?last=ok.png")
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
		writer.FormDataContentType(), form, "/upload/testdata/jpg/?last=ok.jpg")
}

func TestUploadHeic(t *testing.T) {
	cleanup(t, "testdata/heic")
	// for uploads, the directory is not created automatically
	os.MkdirAll("testdata/heic", 0755)
	form := new(bytes.Buffer)
	writer := multipart.NewWriter(form)
	field, _ := writer.CreateFormField("name")
	field.Write([]byte("ok.jpg"))                       // target
	file, _ := writer.CreateFormFile("file", "ok.heic") // source
	// convert -size 1x1 canvas: heic:- | base64
	imgBase64 := `
AAAAGGZ0eXBoZWljAAAAAG1pZjFoZWljAAABqm1ldGEAAAAAAAAAIWhkbHIAAAAAAAAAAHBpY3QA
AAAAAAAAAAAAAAAAAAAADnBpdG0AAAAAAAIAAAAQaWRhdAAAAAAAAQABAAAAOGlsb2MBAAAAREAA
AgABAAAAAAAAAcoAAQAAAAAAAAAtAAIAAQAAAAAAAAABAAAAAAAAAAgAAAA4aWluZgAAAAAAAgAA
ABVpbmZlAgAAAQABAABodmMxAAAAABVpbmZlAgAAAAACAABncmlkAAAAANVpcHJwAAAAs2lwY28A
AABzaHZjQwEDcAAAAAAAAAAAAB7wAPz9+PgAAA8DIAABABhAAQwB//8DcAAAAwCQAAADAAADAB66
AkAhAAEAJ0IBAQNwAAADAJAAAAMAAAMAHqAggQWW6q6a5sCAAAADAIAAAAMAhCIAAQAGRAHBc8GJ
AAAAFGlzcGUAAAAAAAAAQAAAAEAAAAAUaXNwZQAAAAAAAAABAAAAAQAAABBwaXhpAAAAAAMICAgA
AAAaaXBtYQAAAAAAAAACAAECgQIAAgIDhAAAABppcmVmAAAAAAAAAA5kaW1nAAIAAQABAAAANW1k
YXQAAAApKAGvEyE1mvXho5qH3STtzcWnOxedwNIXAKNDaJNqz3uONoCHeUhi/HA=`
	img, err := base64.StdEncoding.DecodeString(imgBase64)
	assert.NoError(t, err)
	file.Write(img)
	writer.Close()
	HTTPUploadAndRedirectTo(t, makeHandler(dropHandler, false), "/drop/testdata/heic/",
		writer.FormDataContentType(), form, "/upload/testdata/heic/?last=ok.jpg")
}

func TestDeleteFile(t *testing.T) {
	cleanup(t, "testdata/delete")
	os.MkdirAll("testdata/delete", 0755)
	assert.NoError(t, os.WriteFile("testdata/delete/nothing.txt", []byte(`# Nothing

I pause and look up
Look at the mountains you say
What happened just now?`), 0644))
	// check that it worked
	assert.FileExists(t, "testdata/delete/nothing.txt")
	// delete it by upload a zero byte file
	form := new(bytes.Buffer)
	writer := multipart.NewWriter(form)
	field, _ := writer.CreateFormField("name")
	field.Write([]byte("nothing.txt"))
	file, _ := writer.CreateFormFile("file", "test.txt")
	file.Write([]byte(""))
	writer.Close()
	HTTPUploadAndRedirectTo(t, makeHandler(dropHandler, false), "/drop/testdata/delete/",
		writer.FormDataContentType(), form, "/upload/testdata/delete/?last=nothing.txt")
	// check that it worked
	assert.NoFileExists(t, "testdata/delete/nothing.txt")
}

func TestUploadMultiple(t *testing.T) {
	cleanup(t, "testdata/multi")
	p := &Page{Name: "testdata/multi/culture", Body: []byte(`# Culture

The road has walls
Iron gates and tree tops
But here: jasmin dreams`)}
	p.save()

	// check location for upload
	body := assert.HTTPBody(makeHandler(viewHandler, false), "GET", "/view/testdata/multi/culture", nil)
	assert.Contains(t, body, `href="/upload/testdata/multi/?filename=culture-1.jpg"`)

	// check location for drop
	body = assert.HTTPBody(makeHandler(uploadHandler, false), "GET", "/upload/testdata/multi/", nil)
	assert.Contains(t, body, `action="/drop/testdata/multi/"`)

	// actually do the upload
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
	assert.Equal(t, "/upload/testdata/multi/", url.Path, "Redirect to upload location")
	values := url.Query()
	assert.Equal(t, "2023-10-02-hike-1.jpg", values.Get("last"))
	assert.Equal(t, "15", values.Get("maxwidth"))
	assert.Equal(t, "50", values.Get("quality"))

	// check the result page
	body = assert.HTTPBody(makeHandler(uploadHandler, false), "GET", url.Path, values)
	assert.Contains(t, body, `value="2023-10-02-hike-2.jpg"`)
	assert.Contains(t, body, `value="15"`)
	assert.Contains(t, body, `value="50"`)
	assert.Contains(t, body, `src="/view/testdata/multi/2023-10-02-hike-1.jpg"`)
}

func TestUploadDir(t *testing.T) {
	cleanup(t, "testdata/dir")
	p := &Page{Name: "testdata/dir/test", Body: []byte(`# Test

Eyes are an abyss
We stare into each other
There is no answer`)}
	p.save()

	// check location for upload
	body := assert.HTTPBody(makeHandler(viewHandler, false), "GET", "/view/testdata/dir/test", nil)
	assert.Contains(t, body, `href="/upload/testdata/dir/?filename=test-1.jpg"`)

	// check location for drop
	body = assert.HTTPBody(makeHandler(uploadHandler, false), "GET", "/upload/testdata/dir/", nil)
	assert.Contains(t, body, `action="/drop/testdata/dir/"`)

	// actually do the upload
	form := new(bytes.Buffer)
	writer := multipart.NewWriter(form)
	field, _ := writer.CreateFormField("name")
	field.Write([]byte("test.jpg"))
	file, _ := writer.CreateFormFile("file", "ok.jpg")
	img := image.NewRGBA(image.Rect(0, 0, 20, 20))
	jpeg.Encode(file, img, &jpeg.Options{Quality: 90})
	writer.Close()
	location := HTTPUploadLocation(t, makeHandler(dropHandler, false), "/drop/testdata/dir/",
		writer.FormDataContentType(), form)
	url, _ := url.Parse(location)
	assert.Equal(t, "/upload/testdata/dir/", url.Path, "Redirect to upload location")
	values := url.Query()
	assert.Equal(t, "test.jpg", values.Get("last"))

	// check the result page
	body = assert.HTTPBody(makeHandler(uploadHandler, false), "GET", url.Path, values)
	assert.Contains(t, body, `src="/view/testdata/dir/test.jpg"`)
}

func TestUploadTwoInOne(t *testing.T) {
	cleanup(t, "testdata/two")
	os.MkdirAll("testdata/two", 0755)
	form := new(bytes.Buffer)
	writer := multipart.NewWriter(form)
	field, _ := writer.CreateFormField("name")
	field.Write([]byte("2024-02-19-hike-1.jpg"))
	file1, _ := writer.CreateFormFile("file", "one.jpg")
	img1 := image.NewRGBA(image.Rect(0, 0, 10, 10))
	jpeg.Encode(file1, img1, &jpeg.Options{Quality: 90})
	file2, _ := writer.CreateFormFile("file", "two.jpg")
	img2 := image.NewRGBA(image.Rect(0, 0, 20, 20))
	jpeg.Encode(file2, img2, &jpeg.Options{Quality: 90})
	writer.Close()
	location := HTTPUploadLocation(t, makeHandler(dropHandler, false), "/drop/testdata/two/",
		writer.FormDataContentType(), form)
	url, _ := url.Parse(location)
	assert.Equal(t, "/upload/testdata/two/", url.Path, "Redirect to upload location")
	values := url.Query()
	assert.Equal(t, "2024-02-19-hike-2.jpg", values.Get("last"))
	// check the files
	assert.FileExists(t, "testdata/two/2024-02-19-hike-1.jpg")
	assert.FileExists(t, "testdata/two/2024-02-19-hike-2.jpg")
}

func TestUploadTwoInOneAgain(t *testing.T) {
	cleanup(t, "testdata/zwei")
	os.MkdirAll("testdata/zwei", 0755)
	form := new(bytes.Buffer)
	writer := multipart.NewWriter(form)
	field, _ := writer.CreateFormField("name")
	field.Write([]byte("image.jpg")) // cannot be incremented!
	file1, _ := writer.CreateFormFile("file", "one.jpg")
	img1 := image.NewRGBA(image.Rect(0, 0, 10, 10))
	jpeg.Encode(file1, img1, &jpeg.Options{Quality: 90})
	file2, _ := writer.CreateFormFile("file", "two.jpg")
	img2 := image.NewRGBA(image.Rect(0, 0, 20, 20))
	jpeg.Encode(file2, img2, &jpeg.Options{Quality: 90})
	writer.Close()
	location := HTTPUploadLocation(t, makeHandler(dropHandler, false), "/drop/testdata/zwei/",
		writer.FormDataContentType(), form)
	url, _ := url.Parse(location)
	assert.Equal(t, "/upload/testdata/zwei/", url.Path, "Redirect to upload location")
	values := url.Query()
	assert.Equal(t, "image-1.jpg", values.Get("last"))
	// check the files
	assert.FileExists(t, "testdata/zwei/image.jpg")
	assert.FileExists(t, "testdata/zwei/image-1.jpg")
}
