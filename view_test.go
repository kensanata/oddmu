package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"regexp"
	"testing"
)

func TestRootHandler(t *testing.T) {
	HTTPRedirectTo(t, rootHandler, "GET", "/", nil, "/view/index")
}

// relies on index.md in the current directory!
func TestViewHandler(t *testing.T) {
	assert.Regexp(t, regexp.MustCompile("Welcome to Oddµ"),
		assert.HTTPBody(makeHandler(viewHandler, true), "GET", "/view/index", nil))
}

func TestPageTitleWithAmp(t *testing.T) {
	cleanup(t, "testdata/amp")

	p := &Page{Name: "testdata/amp/Rock & Roll", Body: []byte("Dancing")}
	p.save()

	assert.Regexp(t, regexp.MustCompile("Rock &amp; Roll"),
		assert.HTTPBody(makeHandler(viewHandler, true), "GET", "/view/testdata/amp/Rock%20%26%20Roll", nil))

	p = &Page{Name: "testdata/amp/Rock & Roll", Body: []byte("# Sex & Drugs & Rock'n'Roll\nOh no!")}
	p.save()

	assert.Regexp(t, regexp.MustCompile("Sex &amp; Drugs"),
		assert.HTTPBody(makeHandler(viewHandler, true), "GET", "/view/testdata/amp/Rock%20%26%20Roll", nil))
}

func TestPageTitleWithQuestionMark(t *testing.T) {
	cleanup(t, "testdata/q")

	p := &Page{Name: "testdata/q/How about no?", Body: []byte("No means no")}
	p.save()

	body := assert.HTTPBody(makeHandler(viewHandler, true), "GET", "/view/testdata/q/How%20about%20no%3F", nil)
	assert.Contains(t, body, "No means no")
	assert.Contains(t, body, "<a href=\"/edit/testdata/q/How%20about%20no%3F\" accesskey=\"e\">Edit</a>")
}

func TestFileLastModified(t *testing.T) {
	cleanup(t, "testdata/file-mod")
	assert.NoError(t, os.Mkdir("testdata/file-mod", 0755))
	assert.NoError(t, os.WriteFile("testdata/file-mod/now.txt", []byte(`
A spider sitting
Unmoving and still
In the autumn chill
`), 0644))
	fi, err := os.Stat("testdata/file-mod/now.txt")
	assert.NoError(t, err)
	h := makeHandler(viewHandler, true)
	assert.Equal(t, []string{fi.ModTime().UTC().Format(http.TimeFormat)},
		HTTPHeaders(h, "GET", "/view/testdata/file-mod/now.txt", nil, "Last-Modified"))
	HTTPStatusCodeIfModifiedSince(t, h, "/view/testdata/file-mod/now.txt", fi.ModTime())
}

// wipes testdata
func TestPageLastModified(t *testing.T) {
	cleanup(t, "testdata/page-mod")
	p := &Page{Name: "testdata/page-mod/now", Body: []byte(`
The sky glows softly
Sadly, the birds are quiet
I like spring better
`)}
	p.save()
	fi, err := os.Stat("testdata/page-mod/now.md")
	assert.NoError(t, err)
	h := makeHandler(viewHandler, true)
	assert.Equal(t, []string{fi.ModTime().UTC().Format(http.TimeFormat)},
		HTTPHeaders(h, "GET", "/view/testdata/page-mod/now", nil, "Last-Modified"))
	HTTPStatusCodeIfModifiedSince(t, h, "/view/testdata/page-mod/now", fi.ModTime())
}

func TestPageHead(t *testing.T) {
	cleanup(t, "testdata/head")
	p := &Page{Name: "testdata/head/peace", Body: []byte(`
No urgent typing
No todos, no list, no queue.
Just me and the birds.
`)}
	p.save()
	fi, err := os.Stat("testdata/head/peace.md")
	assert.NoError(t, err)
	h := makeHandler(viewHandler, true)
	assert.Equal(t, []string(nil),
		HTTPHeaders(h, "HEAD", "/view/testdata/head/war", nil, "Last-Modified"))
	assert.Equal(t, []string(nil),
		HTTPHeaders(h, "GET", "/view/testdata/head/war", nil, "Last-Modified"))
	assert.Equal(t, []string{fi.ModTime().UTC().Format(http.TimeFormat)},
		HTTPHeaders(h, "HEAD", "/view/testdata/head/peace", nil, "Last-Modified"))
	assert.Equal(t, "",
		assert.HTTPBody(h, "HEAD", "/view/testdata/head/peace", nil))
}
