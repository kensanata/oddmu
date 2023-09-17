package main

import (
	"github.com/stretchr/testify/assert"
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
