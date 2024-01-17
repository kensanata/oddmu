package main

import (
	"github.com/stretchr/testify/assert"
	"net/url"
	"os"
	"testing"
	"time"
)

func TestEditSave(t *testing.T) {
	cleanup(t, "testdata/save")

	data := url.Values{}
	data.Set("body", "Hallo!")

	// View of the non-existing page redirects to the edit page
	HTTPRedirectTo(t, makeHandler(viewHandler, true),
		"GET", "/view/testdata/save/alex", nil, "/edit/testdata/save/alex")
	// Edit page can be fetched
	assert.HTTPStatusCode(t, makeHandler(editHandler, true),
		"GET", "/edit/testdata/save/alex", nil, 200)
	// Posting to the save URL saves a page
	HTTPRedirectTo(t, makeHandler(saveHandler, true),
		"POST", "/save/testdata/save/alex", data, "/view/testdata/save/alex")
	// Page now contains the text
	assert.Contains(t, assert.HTTPBody(makeHandler(viewHandler, true),
		"GET", "/view/testdata/save/alex", nil),
		"Hallo!")
	// Delete the page and you're sent to the empty page
	data.Set("body", "")
	HTTPRedirectTo(t, makeHandler(saveHandler, true),
		"POST", "/save/testdata/save/alex", data, "/view/testdata/save/alex")
	// Viewing the non-existing page redirects to the edit page (like in the beginning)
	HTTPRedirectTo(t, makeHandler(viewHandler, true),
		"GET", "/view/testdata/save/alex", nil, "/edit/testdata/save/alex")
}

func TestEditSaveChanges(t *testing.T) {
	cleanup(t, "testdata/notification")
	data := url.Values{}
	data.Set("body", "Hallo!")
	data.Add("notify", "on")
	today := time.Now().Format("2006-01-02")
	// Posting to the save URL saves a page
	HTTPRedirectTo(t, makeHandler(saveHandler, true),
		"POST", "/save/testdata/notification/" + today,
		data, "/view/testdata/notification/" + today)
	// The changes.md file was created
	s, err := os.ReadFile("testdata/notification/changes.md")
	assert.NoError(t, err)
	d := time.Now().Format(time.DateOnly)
	assert.Equal(t, "# Changes\n\n## "+d+
		"\n* [testdata/notification/"+today+"]("+today+")\n",
		string(s))
	// Link added to index.md file
	s, err = os.ReadFile("testdata/notification/index.md")
	assert.NoError(t, err)
	// New index contains just the link
	assert.Equal(t, string(s), "* [testdata/notification/"+today+"]("+today+")\n")
}
