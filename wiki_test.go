package main

import (
	"os"
	"github.com/stretchr/testify/assert"
	"net/url"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

func TestRootHandler(t *testing.T) {
	wr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rootHandler(wr, req)
	assert.Equal(t, http.StatusFound, wr.Code)
	assert.Equal(t, []string{"/view/index"}, wr.Result().Header["Location"])
}

// relies on index.md in the current directory!
func TestViewHandler(t *testing.T) {
	wr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/view/index", nil)
	fn := makeHandler(viewHandler)
	fn(wr, req)
	assert.Equal(t, http.StatusOK, wr.Code)
	assert.Regexp(t, regexp.MustCompile("Welcome to Oddµ"), wr.Body.String())
}

func TestEditSave(t *testing.T) {
	_ = os.RemoveAll("testdata")
	wr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/view/testdata/alex", nil)
	fn := makeHandler(viewHandler)
	fn(wr, req)
	assert.Equal(t, http.StatusFound, wr.Code)
	assert.Equal(t, []string{"/edit/testdata/alex"}, wr.Result().Header["Location"])

	wr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/edit/testdata/alex", nil)
	fn = makeHandler(editHandler)
	fn(wr, req)
	assert.Equal(t, http.StatusOK, wr.Code)

	wr = httptest.NewRecorder()
	data := url.Values{}
	data.Set("body", "Hallo!")
	req = httptest.NewRequest(http.MethodPost, "/save/testdata/alex", strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	fn = makeHandler(saveHandler)
	fn(wr, req)
	assert.Equal(t, http.StatusFound, wr.Code, wr.Body)
	assert.Equal(t, []string{"/view/testdata/alex"}, wr.Result().Header["Location"])
	
	wr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/view/testdata/alex", nil)
	fn = makeHandler(viewHandler)
	fn(wr, req)
	assert.Equal(t, http.StatusOK, wr.Code)
	assert.Regexp(t, regexp.MustCompile("Hallo!"), wr.Body.String())

	t.Cleanup(func() {
		// _ = os.RemoveAll("testdata")
	})
}

func TestAddAppend(t *testing.T) {
	_ = os.RemoveAll("testdata")
	p := &Page{Name: "testdata/fire", Body: []byte(`# Fire
Orange sky above
Reflects a distant fire
It's not `)}
	p.save()

	wr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/view/testdata/fire", nil)
	fn := makeHandler(viewHandler)
	fn(wr, req)
	assert.Equal(t, http.StatusOK, wr.Code)
	assert.Regexp(t, regexp.MustCompile("a distant fire"), wr.Body.String())

	wr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/add/testdata/fire", nil)
	fn = makeHandler(addHandler)
	fn(wr, req)
	assert.Equal(t, http.StatusOK, wr.Code)
	assert.NotRegexp(t, regexp.MustCompile("a distant fire"), wr.Body.String())

	wr = httptest.NewRecorder()
	data := url.Values{}
	data.Set("body", "barbecue")
	req = httptest.NewRequest(http.MethodPost, "/append/testdata/fire", strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	fn = makeHandler(appendHandler)
	fn(wr, req)
	assert.Equal(t, http.StatusFound, wr.Code, wr.Body)
	assert.Equal(t, []string{"/view/testdata/fire"}, wr.Result().Header["Location"])
	
	wr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/view/testdata/fire", nil)
	fn = makeHandler(viewHandler)
	fn(wr, req)
	assert.Equal(t, http.StatusOK, wr.Code)
	assert.Regexp(t, regexp.MustCompile("It’s not barbecue"), wr.Body.String())

	t.Cleanup(func() {
		_ = os.RemoveAll("testdata")
	})
}

func TestPageTitleWithAmp(t *testing.T) {
	_ = os.RemoveAll("testdata")
	p := &Page{Name: "testdata/Rock & Roll", Body: []byte("Dancing")}
	p.save()

	wr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/view/testdata/Rock%20%26%20Roll", nil)
	fn := makeHandler(viewHandler)
	fn(wr, req)
	assert.Equal(t, http.StatusOK, wr.Code)
	assert.Regexp(t, regexp.MustCompile("Rock &amp; Roll"), wr.Body.String())

	p = &Page{Name: "testdata/Rock & Roll", Body: []byte("# Sex & Drugs & Rock'n'Roll\nOh no!")}
	p.save()

	wr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/view/testdata/Rock%20%26%20Roll", nil)
	fn = makeHandler(viewHandler)
	fn(wr, req)
	assert.Equal(t, http.StatusOK, wr.Code)
	assert.Regexp(t, regexp.MustCompile("Sex &amp; Drugs"), wr.Body.String())

	t.Cleanup(func() {
		// _ = os.RemoveAll("testdata")
	})
}
