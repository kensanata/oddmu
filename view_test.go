package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/url"
	"os"
	"testing"
)

func TestRootHandler(t *testing.T) {
	HTTPRedirectTo(t, rootHandler, "GET", "/", nil, "/view/index")
}

// relies on index.md in the current directory!
func TestViewHandler(t *testing.T) {
	assert.Contains(t,
		assert.HTTPBody(makeHandler(viewHandler, false), "GET", "/view/index", nil),
		"Welcome to Oddμ")
}

func TestViewHandlerDir(t *testing.T) {
	cleanup(t, "testdata/dir")
	HTTPRedirectTo(t, makeHandler(viewHandler, false), "GET", "/view/", nil, "/view/index")
	HTTPRedirectTo(t, makeHandler(viewHandler, false), "GET", "/view/testdata", nil, "/view/testdata/index")
	HTTPRedirectTo(t, makeHandler(viewHandler, false), "GET", "/view/testdata/", nil, "/view/testdata/index")
	assert.NoError(t, os.Mkdir("testdata/dir", 0755))
	HTTPRedirectTo(t, makeHandler(viewHandler, false), "GET", "/view/testdata/dir", nil, "/view/testdata/dir/index")
	HTTPRedirectTo(t, makeHandler(viewHandler, false), "GET", "/view/testdata/dir/", nil, "/view/testdata/dir/index")
	assert.NoError(t, os.Mkdir("testdata/dir/dir", 0755))
	HTTPRedirectTo(t, makeHandler(viewHandler, false), "GET", "/view/testdata/dir", nil, "/view/testdata/dir/index")
	HTTPRedirectTo(t, makeHandler(viewHandler, false), "GET", "/view/testdata/dir/", nil, "/view/testdata/dir/index")
	HTTPRedirectTo(t, makeHandler(viewHandler, false), "GET", "/view/testdata/dir/dir", nil, "/view/testdata/dir/dir/index")
	HTTPRedirectTo(t, makeHandler(viewHandler, false), "GET", "/view/testdata/dir/dir/", nil, "/view/testdata/dir/dir/index")
	assert.NoError(t, os.WriteFile("testdata/dir/dir.md", []byte(`# Blackbird

The oven hums and
the music plays, coffee smells
blackbirds sing outside
`), 0644))
	assert.Contains(t, assert.HTTPBody(makeHandler(viewHandler, false), "GET", "/view/testdata/dir/dir", nil), "<h1>Blackbird</h1>")
	assert.Contains(t, assert.HTTPBody(makeHandler(viewHandler, false), "GET", "/view/testdata/dir/dir.md", nil), "# Blackbird")
	HTTPRedirectTo(t, makeHandler(viewHandler, false), "GET", "/view/testdata/dir/dir/", nil, "/view/testdata/dir/dir/index")
}

// relies on index.md in the current directory!
func TestViewHandlerWithId(t *testing.T) {
	data := make(url.Values)
	data.Set("id", "index")
	assert.Contains(t,
		assert.HTTPBody(makeHandler(viewHandler, false), "GET", "/view/", data),
		"Welcome to Oddμ")
}

func TestPageTitleWithAmp(t *testing.T) {
	cleanup(t, "testdata/amp")

	p := &Page{Name: "testdata/amp/Rock & Roll", Body: []byte("Dancing")}
	p.save()

	assert.Contains(t,
		assert.HTTPBody(makeHandler(viewHandler, false), "GET", "/view/testdata/amp/Rock%20%26%20Roll", nil),
		"Rock &amp; Roll")

	p = &Page{Name: "testdata/amp/Rock & Roll", Body: []byte("# Sex & Drugs & Rock'n'Roll\nOh no!")}
	p.save()

	assert.Contains(t,
		assert.HTTPBody(makeHandler(viewHandler, false), "GET", "/view/testdata/amp/Rock%20%26%20Roll", nil),
		"Sex &amp; Drugs")
}

func TestPageTitleWithQuestionMark(t *testing.T) {
	cleanup(t, "testdata/q")

	p := &Page{Name: "testdata/q/How about no?", Body: []byte("No means no")}
	p.save()

	body := assert.HTTPBody(makeHandler(viewHandler, false), "GET", "/view/testdata/q/How%20about%20no%3F", nil)
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
	h := makeHandler(viewHandler, false)
	assert.Equal(t, []string{fi.ModTime().UTC().Format(http.TimeFormat)},
		HTTPHeaders(h, "GET", "/view/testdata/file-mod/now.txt", nil, "Last-Modified"))
	HTTPStatusCodeIfModifiedSince(t, h, "/view/testdata/file-mod/now.txt", fi.ModTime())
}

func TestForbidden(t *testing.T) {
	assert.HTTPStatusCode(t, makeHandler(viewHandler, false), "GET", "/view/", nil, http.StatusFound)
	assert.HTTPStatusCode(t, makeHandler(viewHandler, false), "GET", "/view/.", nil, http.StatusForbidden)
	assert.HTTPStatusCode(t, makeHandler(viewHandler, false), "GET", "/view/.htaccess", nil, http.StatusForbidden)
	assert.HTTPStatusCode(t, makeHandler(viewHandler, false), "GET", "/view/.git/description", nil, http.StatusForbidden)
	assert.HTTPStatusCode(t, makeHandler(viewHandler, false), "GET", "/view/../oddmu", nil, http.StatusForbidden)
	data := make(url.Values)
	data.Set("id", "..")
	assert.HTTPStatusCode(t, makeHandler(viewHandler, false), "GET", "/view/", data, http.StatusForbidden)
	data.Set("id", "foo/bar")
	assert.HTTPStatusCode(t, makeHandler(viewHandler, false), "GET", "/view/", data, http.StatusBadRequest)
}

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
	h := makeHandler(viewHandler, false)
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
	h := makeHandler(viewHandler, false)
	assert.Equal(t, []string(nil),
		HTTPHeaders(h, "HEAD", "/view/testdata/head/war", nil, "Last-Modified"))
	assert.Equal(t, []string(nil),
		HTTPHeaders(h, "GET", "/view/testdata/head/war", nil, "Last-Modified"))
	assert.Equal(t, []string{fi.ModTime().UTC().Format(http.TimeFormat)},
		HTTPHeaders(h, "HEAD", "/view/testdata/head/peace", nil, "Last-Modified"))
	assert.Equal(t, "",
		assert.HTTPBody(h, "HEAD", "/view/testdata/head/peace", nil))
}

func TestViewUmlaut(t *testing.T) {
	assert.Contains(t,
		assert.HTTPBody(makeHandler(viewHandler, false), "GET", "/view/%C3%A4rger", nil),
		`<a href="/edit/%C3%A4rger">`)
}

func TestMimeType(t *testing.T) {
	assert.Equal(t, []string{"text/markdown; charset=utf-8"},
		HTTPHeaders(makeHandler(viewHandler, false), "GET", "/view/index.md", nil, "Content-Type"))
	assert.Equal(t, []string{"text/html; charset=utf-8"},
		HTTPHeaders(makeHandler(viewHandler, false), "GET", "/view/view.html", nil, "Content-Type"))
}
