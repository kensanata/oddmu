package main

import (
	"github.com/stretchr/testify/assert"
	"os"
	"regexp"
	"testing"
	"time"
)

// Note TestEditSaveChanges and TestAddAppendChanges.

func TestAddLinkToPageWithNoList(t *testing.T) {
	// no newlines
	title := "# Test"
	p := &Page{Body: []byte(title)}
	re := regexp.MustCompile(`(?m)^\* \[[^\]]+\]\(2025-08-08\)\n`)
	link := "* [2025-08-08](2025-08-08)\n"
	addLinkToPage(p, link, re)
	assert.Equal(t, title+"\n\n"+link, string(p.Body))
}

func TestAddLinkToPageWithOlderLink(t *testing.T) {
	// one newline
	title := "# Test\n"
	old := "* [2025-08-08](2025-08-08)\n"
	p := &Page{Body: []byte(title + old)}
	re := regexp.MustCompile(`(?m)^\* \[[^\]]+\]\(2025-08-10\)\n`)
	link := "* [2025-08-10](2025-08-10)\n"
	addLinkToPage(p, link, re)
	assert.Equal(t, title+"\n"+link+old, string(p.Body))
}

func TestAddLinkToPageBetweenToExistingLinks(t *testing.T) {
	title := "# Test\n\n"
	new := "* [2025-08-10](2025-08-10)\n"
	old := "* [2025-08-08](2025-08-08)\n"
	p := &Page{Body: []byte(title + new + old)}
	re := regexp.MustCompile(`(?m)^\* \[[^\]]+\]\(2025-08-09\)\n`)
	link := "* [2025-08-09](2025-08-09)\n"
	addLinkToPage(p, link, re)
	assert.Equal(t, title+new+link+old, string(p.Body))
}

func TestAddLinkToPageBetweenToExistingLinks2(t *testing.T) {
	title := "# Test\n\n"
	new := "* [2025-08-10](2025-08-10)\n* [2025-08-09](2025-08-09)\n"
	old := "* [2025-08-07](2025-08-07)\n"
	p := &Page{Body: []byte(title + new + old)}
	re := regexp.MustCompile(`(?m)^\* \[[^\]]+\]\(2025-08-08\)\n`)
	link := "* [2025-08-08](2025-08-08)\n"
	addLinkToPage(p, link, re)
	assert.Equal(t, title+new+link+old, string(p.Body))
}

func TestAddLinkToPageAtTheEnd(t *testing.T) {
	title := "# Test\n\n"
	new := "* [2025-08-10](2025-08-10)\n"
	old := "* [2025-08-08](2025-08-08)\n"
	p := &Page{Body: []byte(title + new + old)}
	re := regexp.MustCompile(`(?m)^\* \[[^\]]+\]\(2025-08-07\)\n`)
	link := "* [2025-08-07](2025-08-07)\n"
	addLinkToPage(p, link, re)
	assert.Equal(t, title+new+old+link, string(p.Body))
}

func TestChanges(t *testing.T) {
	cleanup(t, "testdata/washing")
	today := time.Now().Format(time.DateOnly)
	p := &Page{Name: "testdata/washing/" + today + "-machine",
		Body: []byte(`# Washing machine
Churning growling thing
Water spraying in a box 
Out of sight and dark`)}
	p.notify()
	// Link added to changes.md file
	s, err := os.ReadFile("testdata/washing/changes.md")
	assert.NoError(t, err)
	assert.Contains(t, string(s), "[Washing machine]("+today+"-machine)")
	// Link added to index.md file
	s, err = os.ReadFile("testdata/washing/index.md")
	assert.NoError(t, err)
	// New index contains just the link
	assert.Equal(t, string(s), "* [Washing machine]("+today+"-machine)\n")
}

func TestChangesWithHashtag(t *testing.T) {
	cleanup(t, "testdata/changes")
	intro := "# Haiku\n"
	line := "* [Hotel room](2023-10-27-hotel)\n"
	h := &Page{Name: "testdata/changes/Haiku", Body: []byte(intro)}
	h.save()
	p := &Page{Name: "testdata/changes/2023-10-27-hotel",
		Body: []byte(`# Hotel room
White linen and white light
Wooden floor and painted walls
Home away from home

#Haiku #Poetry`)}
	p.notify()
	s, err := os.ReadFile("testdata/changes/changes.md")
	assert.NoError(t, err)
	assert.Contains(t, string(s), line)
	s, err = os.ReadFile("testdata/changes/Haiku.md")
	assert.NoError(t, err)
	// ensure an empty line when adding at the end of the page
	assert.Equal(t, intro+"\n"+line, string(s))
	assert.NoFileExists(t, "testdata/changes/Poetry.md")
}

func TestChangesWithList(t *testing.T) {
	cleanup(t, "testdata/changes")
	intro := "# Changes\n\nThis is a paragraph.\n\n"
	d := "## " + time.Now().Format(time.DateOnly) + "\n"
	line := "* [a change](change)\n"
	assert.NoError(t, os.MkdirAll("testdata/changes", 0755))
	assert.NoError(t, os.WriteFile("testdata/changes/changes.md", []byte(intro+d+line), 0644))
	p := &Page{Name: "testdata/changes/alex", Body: []byte(`Hallo!`)}
	p.notify()
	s, err := os.ReadFile("testdata/changes/changes.md")
	assert.NoError(t, err)
	new_line := "* [testdata/changes/alex](alex)\n"
	// new line was added at the beginning of the list
	assert.Equal(t, intro+d+new_line+line, string(s))
}

func TestChangesWithOldList(t *testing.T) {
	cleanup(t, "testdata/changes")
	intro := "# Changes\n\nThis is a paragraph.\n\n"
	line := "* [a change](change)\n"
	y := "## " + time.Now().Add(-24*time.Hour).Format(time.DateOnly) + "\n"
	d := "## " + time.Now().Format(time.DateOnly) + "\n"
	assert.NoError(t, os.MkdirAll("testdata/changes", 0755))
	assert.NoError(t, os.WriteFile("testdata/changes/changes.md", []byte(intro+y+line), 0644))
	p := &Page{Name: "testdata/changes/alex", Body: []byte(`Hallo!`)}
	p.notify()
	s, err := os.ReadFile("testdata/changes/changes.md")
	assert.NoError(t, err)
	new_line := "* [testdata/changes/alex](alex)\n"
	// new line was added at the beginning of the list
	assert.Equal(t, intro+d+new_line+"\n"+y+line, string(s))
}

func TestChangesWithOldDisappearingListAtTheEnd(t *testing.T) {
	cleanup(t, "testdata/changes")
	intro := "# Changes\n\nThis is a paragraph.\n\n"
	line := "* [a change](alex)\n"
	y := "## " + time.Now().Add(-24*time.Hour).Format(time.DateOnly) + "\n"
	d := "## " + time.Now().Format(time.DateOnly) + "\n"
	assert.NoError(t, os.MkdirAll("testdata/changes", 0755))
	assert.NoError(t, os.WriteFile("testdata/changes/changes.md", []byte(intro+y+line), 0644))
	p := &Page{Name: "testdata/changes/alex", Body: []byte(`Hallo!`)}
	p.notify()
	s, err := os.ReadFile("testdata/changes/changes.md")
	assert.NoError(t, err)
	new_line := "* [testdata/changes/alex](alex)\n"
	// new line was added at the beginning of the list, with the new date, and the old date disappeared
	assert.Equal(t, intro+d+new_line, string(s))
}

func TestChangesWithOldDisappearingListInTheMiddle(t *testing.T) {
	cleanup(t, "testdata/changes")
	intro := "# Changes\n\nThis is a paragraph.\n\n"
	line := "* [a change](alex)\n"
	other := "* [other change](whatever)\n"
	yy := "## " + time.Now().Add(-48*time.Hour).Format(time.DateOnly) + "\n"
	y := "## " + time.Now().Add(-24*time.Hour).Format(time.DateOnly) + "\n"
	d := "## " + time.Now().Format(time.DateOnly) + "\n"
	assert.NoError(t, os.MkdirAll("testdata/changes", 0755))
	assert.NoError(t, os.WriteFile("testdata/changes/changes.md", []byte(intro+y+line+"\n"+yy+other), 0644))
	p := &Page{Name: "testdata/changes/alex", Body: []byte(`Hallo!`)}
	p.notify()
	s, err := os.ReadFile("testdata/changes/changes.md")
	assert.NoError(t, err)
	new_line := "* [testdata/changes/alex](alex)\n"
	// new line was added at the beginning of the list, with the new date, and the old date disappeared
	assert.Equal(t, intro+d+new_line+"\n"+yy+other, string(s))
}

func TestChangesWithListAtTheTop(t *testing.T) {
	cleanup(t, "testdata/changes")
	line := "* [a change](change)\n"
	d := "## " + time.Now().Format(time.DateOnly) + "\n"
	assert.NoError(t, os.MkdirAll("testdata/changes", 0755))
	assert.NoError(t, os.WriteFile("testdata/changes/changes.md", []byte(line), 0644))
	p := &Page{Name: "testdata/changes/alex", Body: []byte(`Hallo!`)}
	p.notify()
	s, err := os.ReadFile("testdata/changes/changes.md")
	assert.NoError(t, err)
	new_line := "* [testdata/changes/alex](alex)\n"
	// new line was added at the top, no error due to missing introduction
	assert.Equal(t, d+new_line+line, string(s))
}

func TestChangesWithNoList(t *testing.T) {
	cleanup(t, "testdata/changes")
	intro := "# Changes\n\nThis is a paragraph."
	d := "## " + time.Now().Format(time.DateOnly) + "\n"
	assert.NoError(t, os.MkdirAll("testdata/changes", 0755))
	assert.NoError(t, os.WriteFile("testdata/changes/changes.md", []byte(intro), 0644))
	p := &Page{Name: "testdata/changes/alex", Body: []byte(`Hallo!`)}
	p.notify()
	s, err := os.ReadFile("testdata/changes/changes.md")
	assert.NoError(t, err)
	new_line := "* [testdata/changes/alex](alex)\n"
	// into is still there and a new list was started
	assert.Equal(t, intro+"\n\n"+d+new_line, string(s))
}

func TestChangesWithUpdate(t *testing.T) {
	cleanup(t, "testdata/changes")
	intro := "# Changes\n\nThis is a paragraph.\n\n"
	other := "* [other change](whatever)\n"
	d := "## " + time.Now().Format(time.DateOnly) + "\n"
	line := "* [a change](alex)\n"
	assert.NoError(t, os.MkdirAll("testdata/changes", 0755))
	assert.NoError(t, os.WriteFile("testdata/changes/changes.md", []byte(intro+d+other+line), 0644))
	p := &Page{Name: "testdata/changes/alex", Body: []byte(`Hallo!`)}
	p.notify()
	s, err := os.ReadFile("testdata/changes/changes.md")
	assert.NoError(t, err)
	new_line := "* [testdata/changes/alex](alex)\n"
	// the change was already listed, but now it moved up and has a new title
	assert.Equal(t, intro+d+new_line+other, string(s))
}

func TestChangesWithNoChangeToTheOrder(t *testing.T) {
	cleanup(t, "testdata/changes")
	intro := "# Changes\n\nThis is a paragraph.\n\n"
	d := "## " + time.Now().Format(time.DateOnly) + "\n"
	line := "* [a change](alex)\n"
	other := "* [other change](whatever)\n"
	assert.NoError(t, os.MkdirAll("testdata/changes", 0755))
	assert.NoError(t, os.WriteFile("testdata/changes/changes.md", []byte(intro+d+line+other), 0644))
	p := &Page{Name: "testdata/changes/alex", Body: []byte(`Hallo!`)}
	p.notify()
	s, err := os.ReadFile("testdata/changes/changes.md")
	assert.NoError(t, err)
	new_line := "* [testdata/changes/alex](alex)\n"
	// the change was already listed at the top, so just use the new title
	assert.Equal(t, intro+d+new_line+other, string(s))
	// since the file has changed, a backup was necessary
	assert.FileExists(t, "testdata/changes/changes.md~")
}

func TestChangesWithNoChanges(t *testing.T) {
	cleanup(t, "testdata/changes")
	intro := "# Changes\n\nThis is a paragraph.\n\n"
	d := "## " + time.Now().Format(time.DateOnly) + "\n"
	line := "* [a change](alex)\n"
	other := "* [other change](whatever)\n"
	assert.NoError(t, os.MkdirAll("testdata/changes", 0755))
	assert.NoError(t, os.WriteFile("testdata/changes/changes.md", []byte(intro+d+line+other), 0644))
	p := &Page{Name: "testdata/changes/alex", Body: []byte("# a change\nHallo!")}
	p.notify()
	s, err := os.ReadFile("testdata/changes/changes.md")
	assert.NoError(t, err)
	// the change was already listed at the top, so no change was necessary
	assert.Equal(t, intro+d+line+other, string(s))
	// since the file hasn't changed, no backup was necessary
	assert.NoFileExists(t, "testdata/changes/changes.md~")
}
