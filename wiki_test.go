package main

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"slices"
	"strings"
	"testing"
	"time"
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
// that POST requests ignore the query part of the URL.
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
	assert.NoError(t, err)
	handler(w, req)
	code := w.Code
	isRedirectCode := code >= http.StatusMultipleChoices && code <= http.StatusTemporaryRedirect
	assert.True(t, isRedirectCode, "Expected HTTP redirect status code for %q but received %d", url+"?"+values.Encode(), code)
	headers := w.Result().Header["Location"]
	assert.True(t, len(headers) == 1 && headers[0] == destination,
		"Expected HTTP redirect location %s for %q but received %v", destination, url+"?"+values.Encode(), headers)
	return isRedirectCode
}

// HTTPUploadLocation returns the location header after an upload.
func HTTPUploadLocation(t *testing.T, handler http.HandlerFunc, url, contentType string, body *bytes.Buffer) string {
	w := httptest.NewRecorder()
	req, err := http.NewRequest("POST", url, body)
	req.Header.Set("Content-Type", contentType)
	assert.NoError(t, err)
	handler(w, req)
	code := w.Code
	isRedirectCode := code >= http.StatusMultipleChoices && code <= http.StatusTemporaryRedirect
	assert.True(t, isRedirectCode, "Expected HTTP redirect status code for %q but received %d", url, code)
	headers := w.Result().Header["Location"]
	assert.True(t, len(headers) == 1, "Expected a single redirect header but got %d locations", len(headers))
	return headers[0]
}

// HTTPUploadAndRedirectTo checks that the request results in a redirect and it
// checks the destination of the redirect. It returns whether the
// request did in fact result in a redirect.
func HTTPUploadAndRedirectTo(t *testing.T, handler http.HandlerFunc, url, contentType string, body *bytes.Buffer, destination string) {
	location := HTTPUploadLocation(t, handler, url, contentType, body)
	assert.Equal(t, destination, location,
		"Expected HTTP redirect location %s for %q but received %s", destination, url, location)
}

// HTTPStatusCodeIfModifiedSince checks that the request results in a
// 304 response for the given time.
func HTTPStatusCodeIfModifiedSince(t *testing.T, handler http.HandlerFunc, url string, ti time.Time) {
	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", url, nil)
	assert.NoError(t, err)
	req.Header.Set("If-Modified-Since", ti.UTC().Format(http.TimeFormat))
	handler(w, req)
	assert.Equal(t, http.StatusNotModified, w.Code)
}

// restore remembers the file content before the test starts and restores the file at the end. Important for files such
// as "index.md".
func restore(t *testing.T, files ...string) {
	data := make(map[string][]byte)
	stat := make(map[string]os.FileInfo)
	for _, file := range files {
		s, err := os.Stat(file)
		if err != nil {
			t.Log("Could not stat ", file, ": ", err)
			continue;
		}
		c, err := os.ReadFile(file)
		if err != nil {
			t.Log("Could not read ", file, ": ", err)
			continue;
		}
		stat[file] = s
		data[file] = c

	}
	t.Cleanup(func() {
		for file, c := range data {
			m := stat[file].Mode()
			err := os.WriteFile(file, c, m)
			if err != nil {
				t.Log("Could not restore ", file, ": ", err)
			}
			t := stat[file].ModTime()
			os.Chtimes(file, t, t)
		}
	})
}

// cleanup deletes any directories mentioned and removes all pages in those directories from the index. Incidentally, if
// a filename such as "changes.md" or "changes.md~" is provided instead of a directory, then that page file is removed
// and any mention of it is removed from the index.
func cleanup(t *testing.T, dirs ...string) {
	t.Cleanup(func() {
		for _, dir := range dirs {
			_ = os.RemoveAll(dir)
		}
		index.Lock()
		defer index.Unlock()
		for name := range index.titles {
			for _, dir := range dirs {
				if strings.HasPrefix(name, dir) {
					delete(index.titles, name)
				}
			}
		}
		ids := []docid{}
		for id, name := range index.documents {
			for _, dir := range dirs {
				if strings.HasPrefix(name, dir) {
					delete(index.documents, id)
					ids = append(ids, id)
				}
			}
		}
		for hashtag, docs := range index.token {
			index.token[hashtag] = slices.DeleteFunc(ids, func(id docid) bool {
				return slices.Contains(docs, id)
			})
		}
	})
}
