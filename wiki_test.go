package main

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
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

// HTTPUploadAndRedirectTo checks that the request results in a redirect and it
// checks the destination of the redirect. It returns whether the
// request did in fact result in a redirect.
func HTTPUploadAndRedirectTo(t *testing.T, handler http.HandlerFunc, url, contentType string, body *bytes.Buffer, destination string) bool {
	w := httptest.NewRecorder()
	req, err := http.NewRequest("POST", url, body)
	req.Header.Set("Content-Type", contentType)
	assert.NoError(t, err)
	handler(w, req)
	code := w.Code
	isRedirectCode := code >= http.StatusMultipleChoices && code <= http.StatusTemporaryRedirect
	assert.True(t, isRedirectCode, "Expected HTTP redirect status code for %q but received %d", url, code)
	headers := w.Result().Header["Location"]
	assert.True(t, len(headers) == 1 && headers[0] == destination,
		"Expected HTTP redirect location %s for %q but received %v", destination, url, headers)
	return isRedirectCode
}
