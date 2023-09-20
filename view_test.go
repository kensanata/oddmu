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

// wipes testdata
func TestPageTitleWithAmp(t *testing.T) {
	_ = os.RemoveAll("testdata")

	p := &Page{Name: "testdata/Rock & Roll", Body: []byte("Dancing")}
	p.save()

	assert.Regexp(t, regexp.MustCompile("Rock &amp; Roll"),
		assert.HTTPBody(makeHandler(viewHandler, true), "GET", "/view/testdata/Rock%20%26%20Roll", nil))

	p = &Page{Name: "testdata/Rock & Roll", Body: []byte("# Sex & Drugs & Rock'n'Roll\nOh no!")}
	p.save()

	assert.Regexp(t, regexp.MustCompile("Sex &amp; Drugs"),
		assert.HTTPBody(makeHandler(viewHandler, true), "GET", "/view/testdata/Rock%20%26%20Roll", nil))

	t.Cleanup(func() {
		_ = os.RemoveAll("testdata")
	})
}

// wipes testdata
func TestPageTitleWithQuestionMark(t *testing.T) {
	_ = os.RemoveAll("testdata")

	p := &Page{Name: "testdata/How about no?", Body: []byte("No means no")}
	p.save()

	body := assert.HTTPBody(makeHandler(viewHandler, true), "GET", "/view/testdata/How%20about%20no%3F", nil)
	assert.Contains(t, body, "No means no")
	assert.Contains(t, body, "<a href=\"/edit/testdata/How%20about%20no%3F\" accesskey=\"e\">Edit</a>")

	t.Cleanup(func() {
		_ = os.RemoveAll("testdata")
	})
}

// wipes testdata
func TestFileLastModified(t *testing.T) {
	_ = os.RemoveAll("testdata")
	assert.NoError(t, os.Mkdir("testdata", 0755))
	assert.NoError(t, os.WriteFile("testdata/now.txt", []byte(`
A spider sitting
Unmoving and still
In the autumn chill
`), 0644))
	fi, err := os.Stat("testdata/now.txt")
	assert.NoError(t, err)
	h := makeHandler(viewHandler, true)
	assert.Equal(t, []string{fi.ModTime().UTC().Format(http.TimeFormat)},
		HTTPHeaders(h, "GET", "/view/testdata/now.txt", nil, "Last-Modified"))
	HTTPStatusCodeIfModifiedSince(t, h, "/view/testdata/now.txt", fi.ModTime())
	t.Cleanup(func() {
		_ = os.RemoveAll("testdata")
	})
}

// wipes testdata
func TestPageLastModified(t *testing.T) {
	_ = os.RemoveAll("testdata")
	p := &Page{Name: "testdata/now", Body: []byte(`
The sky glows softly
Sadly, the birds are quiet
I like spring better
`)}
	p.save()
	fi, err := os.Stat("testdata/now.md")
	assert.NoError(t, err)
	h := makeHandler(viewHandler, true)
	assert.Equal(t, []string{fi.ModTime().UTC().Format(http.TimeFormat)},
	HTTPHeaders(h, "GET", "/view/testdata/now", nil, "Last-Modified"))
	HTTPStatusCodeIfModifiedSince(t, h, "/view/testdata/now", fi.ModTime())
	t.Cleanup(func() {
		_ = os.RemoveAll("testdata")
	})
}


// wipes testdata
func TestPageHead(t *testing.T) {
	_ = os.RemoveAll("testdata")
	p := &Page{Name: "testdata/peace", Body: []byte(`
No urgent typing
No todos, no list, no queue.
Just me and the birds.
`)}
	p.save()
	fi, err := os.Stat("testdata/peace.md")
	assert.NoError(t, err)
	h := makeHandler(viewHandler, true)
	assert.Equal(t, []string(nil),
		HTTPHeaders(h, "HEAD", "/view/testdata/war", nil, "Last-Modified"))
	assert.Equal(t, []string(nil),
		HTTPHeaders(h, "GET", "/view/testdata/war", nil, "Last-Modified"))
	assert.Equal(t, []string{fi.ModTime().UTC().Format(http.TimeFormat)},
		HTTPHeaders(h, "HEAD", "/view/testdata/peace", nil, "Last-Modified"))
	assert.Equal(t, "",
		assert.HTTPBody(h, "HEAD", "/view/testdata/peace", nil))
	t.Cleanup(func() {
		_ = os.RemoveAll("testdata")
	})
}
