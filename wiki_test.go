package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"
)

// HTTPHeaders is a helper that returns HTTP headers of the response. It returns
// nil if building a new request fails.
func HTTPHeaders(handler http.HandlerFunc, method, url string, values url.Values, header string) []string {
	w := httptest.NewRecorder()
	req, err := http.NewRequest(method, url+"?"+values.Encode(), nil)
	if err != nil {
		return nil
	}
	handler(w, req)
	return w.Result().Header[header]
}

// HTTPRedirectTo checks that the request results in a redirect and it
// checks the destination of the redirect. It returns whether the
// request did in fact result in a redirect. Note: This method assumes
// that POST requests ignore the query part of the URL which is often
// true but not mandated by the standards.
func HTTPRedirectTo(t *testing.T, handler http.HandlerFunc, method, url string, values url.Values, destination string) bool {
	w := httptest.NewRecorder()
	var req *http.Request
	var err error
	if method == http.MethodPost {
		body := strings.NewReader(values.Encode())
		req, err = http.NewRequest(method, url, body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		req, err = http.NewRequest(method, url+"?"+values.Encode(), nil)
	}
	if err != nil {
		assert.Fail(t, fmt.Sprintf("Failed to build test request, got error: %s", err))
	}
	handler(w, req)
	code := w.Code
	isRedirectCode := code >= http.StatusMultipleChoices && code <= http.StatusTemporaryRedirect
	if !isRedirectCode {
		assert.Fail(t, fmt.Sprintf("Expected HTTP redirect status code for %q but received %d", url+"?"+values.Encode(), code))
	}
	headers := w.Result().Header["Location"]
	if len(headers) != 1 || headers[0] != destination {
		assert.Fail(t, fmt.Sprintf("Expected HTTP redirect location %s for %q but received %v", destination, url+"?"+values.Encode(), headers))
	}
	return isRedirectCode
}

func TestRootHandler(t *testing.T) {
	HTTPRedirectTo(t, rootHandler, "GET", "/", nil, "/view/index")
}

// relies on index.md in the current directory!
func TestViewHandler(t *testing.T) {
	assert.Regexp(t, regexp.MustCompile("Welcome to Oddµ"),
		assert.HTTPBody(makeHandler(viewHandler), "GET", "/view/index", nil))
}

// wipes testdata
func TestEditSave(t *testing.T) {
	_ = os.RemoveAll("testdata")

	data := url.Values{}
	data.Set("body", "Hallo!")

	HTTPRedirectTo(t, makeHandler(viewHandler), "GET", "/view/testdata/alex", nil, "/edit/testdata/alex")
	assert.HTTPStatusCode(t, makeHandler(editHandler), "GET", "/edit/testdata/alex", nil, 200)
	HTTPRedirectTo(t, makeHandler(saveHandler), "POST", "/save/testdata/alex", data, "/view/testdata/alex")
	assert.Regexp(t, regexp.MustCompile("Hallo!"),
		assert.HTTPBody(makeHandler(viewHandler), "GET", "/view/testdata/alex", nil))

	t.Cleanup(func() {
		_ = os.RemoveAll("testdata")
	})
}

// wipes testdata
func TestAddAppend(t *testing.T) {
	_ = os.RemoveAll("testdata")

	p := &Page{Name: "testdata/fire", Body: []byte(`# Fire
Orange sky above
Reflects a distant fire
It's not `)}
	p.save()

	data := url.Values{}
	data.Set("body", "barbecue")

	assert.Regexp(t, regexp.MustCompile("a distant fire"),
		assert.HTTPBody(makeHandler(viewHandler), "GET", "/view/testdata/fire", nil))
	assert.NotRegexp(t, regexp.MustCompile("a distant fire"),
		assert.HTTPBody(makeHandler(addHandler), "GET", "/add/testdata/fire", nil))
	HTTPRedirectTo(t, makeHandler(appendHandler), "POST", "/append/testdata/fire", data, "/view/testdata/fire")
	assert.Regexp(t, regexp.MustCompile("It’s not barbecue"),
		assert.HTTPBody(makeHandler(viewHandler), "GET", "/view/testdata/fire", nil))

	t.Cleanup(func() {
		_ = os.RemoveAll("testdata")
	})
}

// wipes testdata
func TestPageTitleWithAmp(t *testing.T) {
	_ = os.RemoveAll("testdata")

	p := &Page{Name: "testdata/Rock & Roll", Body: []byte("Dancing")}
	p.save()

	assert.Regexp(t, regexp.MustCompile("Rock &amp; Roll"),
		assert.HTTPBody(makeHandler(viewHandler), "GET", "/view/testdata/Rock%20%26%20Roll", nil))

	p = &Page{Name: "testdata/Rock & Roll", Body: []byte("# Sex & Drugs & Rock'n'Roll\nOh no!")}
	p.save()

	assert.Regexp(t, regexp.MustCompile("Sex &amp; Drugs"),
		assert.HTTPBody(makeHandler(viewHandler), "GET", "/view/testdata/Rock%20%26%20Roll", nil))

	t.Cleanup(func() {
		_ = os.RemoveAll("testdata")
	})
}

func TestPageTitleWithQuestionMark(t *testing.T) {
	_ = os.RemoveAll("testdata")

	p := &Page{Name: "testdata/How about no?", Body: []byte("No means no")}
	p.save()

	body := assert.HTTPBody(makeHandler(viewHandler), "GET", "/view/testdata/How%20about%20no%3F", nil)
	assert.Contains(t, body, "No means no")
	assert.Contains(t, body, "<a href=\"/edit/testdata/How%20about%20no%3F\">Edit</a>")

	t.Cleanup(func() {
		_ = os.RemoveAll("testdata")
	})
}
