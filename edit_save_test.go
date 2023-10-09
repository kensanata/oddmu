package main

import (
	"github.com/stretchr/testify/assert"
	"net/url"
	"os"
	"testing"
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

func TestEditSaveNotificationNoChanges(t *testing.T) {
	cleanup(t, "testdata/notification", "changes.md")
	os.Remove("changes.md")
	data := url.Values{}
	data.Set("body", "Hallo!")
	data.Add("notify", "on")
	// Posting to the save URL saves a page
	HTTPRedirectTo(t, makeHandler(saveHandler, true),
		"POST", "/save/testdata/notification/alex", data, "/view/testdata/notification/alex")
	// The changes.md file was created
	s, err := os.ReadFile("changes.md")
	assert.NoError(t, err)
	assert.Equal(t, "# Changes\n\n* [testdata/notification/alex](testdata/notification/alex)\n", string(s))
}

func TestEditSaveNotificationWithChanges(t *testing.T) {
	cleanup(t, "testdata/notification", "changes.md")
	intro := "# Changes\n\nThis is a paragraph.\n\n"
	line := "* [a change](change)\n"
	os.WriteFile("changes.md", []byte(intro+line), 0644)
	data := url.Values{}
	data.Set("body", "Hallo!")
	data.Add("notify", "on")
	HTTPRedirectTo(t, makeHandler(saveHandler, true),
		"POST", "/save/testdata/notification/alex", data, "/view/testdata/notification/alex")
	s, err := os.ReadFile("changes.md")
	assert.NoError(t, err)
	new_line := "* [testdata/notification/alex](testdata/notification/alex)\n"
	// new line was added at the beginning of the list
	assert.Equal(t, intro+new_line+line, string(s))
}

func TestEditSaveNotificationWithChangesAtTheTop(t *testing.T) {
	cleanup(t, "testdata/notification", "changes.md")
	line := "* [a change](change)\n"
	os.WriteFile("changes.md", []byte(line), 0644)
	data := url.Values{}
	data.Set("body", "Hallo!")
	data.Add("notify", "on")
	HTTPRedirectTo(t, makeHandler(saveHandler, true),
		"POST", "/save/testdata/notification/alex", data, "/view/testdata/notification/alex")
	s, err := os.ReadFile("changes.md")
	assert.NoError(t, err)
	new_line := "* [testdata/notification/alex](testdata/notification/alex)\n"
	// new line was added at the top, no error due to missing introduction
	assert.Equal(t, new_line+line, string(s))
}

func TestEditSaveNotificationWithNoChangesList(t *testing.T) {
	cleanup(t, "testdata/notification", "changes.md")
	intro := "# Changes\n\nThis is a paragraph."
	os.WriteFile("changes.md", []byte(intro), 0644)
	data := url.Values{}
	data.Set("body", "Hallo!")
	data.Add("notify", "on")
	HTTPRedirectTo(t, makeHandler(saveHandler, true),
		"POST", "/save/testdata/notification/alex", data, "/view/testdata/notification/alex")
	s, err := os.ReadFile("changes.md")
	assert.NoError(t, err)
	new_line := "* [testdata/notification/alex](testdata/notification/alex)\n"
	// into is still there and a new list was started
	assert.Equal(t, intro+"\n\n"+new_line, string(s))
}

func TestEditSaveNotificationWithUpdateToChangesList(t *testing.T) {
	cleanup(t, "testdata/notification", "changes.md")
	intro := "# Changes\n\nThis is a paragraph.\n\n"
	other := "* [other change](testdata/notification/whatever)\n"
	line := "* [a change](testdata/notification/alex)\n"
	os.WriteFile("changes.md", []byte(intro+other+line), 0644)
	data := url.Values{}
	data.Set("body", "Hallo!")
	data.Add("notify", "on")
	HTTPRedirectTo(t, makeHandler(saveHandler, true),
		"POST", "/save/testdata/notification/alex", data, "/view/testdata/notification/alex")
	s, err := os.ReadFile("changes.md")
	assert.NoError(t, err)
	new_line := "* [testdata/notification/alex](testdata/notification/alex)\n"
	// the change was already listed, but now it moved up and has a new title
	assert.Equal(t, intro+new_line+other, string(s))
}

func TestEditSaveNotificationWithUpdateToChangesListTop(t *testing.T) {
	cleanup(t, "testdata/notification", "changes.md")
	intro := "# Changes\n\nThis is a paragraph.\n\n"
	other := "* [other change](testdata/notification/whatever)\n"
	line := "* [a change](testdata/notification/alex)\n"
	os.WriteFile("changes.md", []byte(intro+line+other), 0644)
	data := url.Values{}
	data.Set("body", "Hallo!")
	data.Add("notify", "on")
	HTTPRedirectTo(t, makeHandler(saveHandler, true),
		"POST", "/save/testdata/notification/alex", data, "/view/testdata/notification/alex")
	s, err := os.ReadFile("changes.md")
	assert.NoError(t, err)
	new_line := "* [testdata/notification/alex](testdata/notification/alex)\n"
	// the change was already listed at the top, so just use the new title
	assert.Equal(t, intro+new_line+other, string(s))
}
