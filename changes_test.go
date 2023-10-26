package main

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

// Note TestEditSaveChanges and TestAddAppendChanges.

func TestChanges(t *testing.T) {
	cleanup(t, "changes.md", "changes.md~")
	os.Remove("changes.md")
	p := &Page{Name: "testdata/machine",
		Body: []byte(`# Washing machine
Churning growling thing
Water spraying in a box 
Out of sight and dark`)}
	p.notify()
	assert.FileExists(t, "changes.md")
	s, err := os.ReadFile("changes.md")
	assert.NoError(t, err)
	assert.Contains(t, string(s), "[Washing machine](testdata/machine)")
}

func TestChangesWithHashtag(t *testing.T) {
	cleanup(t, "changes.md", "changes.md~")
	os.Remove("changes.md")
	intro := "# Haiku\n"
	line := "* [Hotel room](testdata/changes/hotel)\n"
	h := &Page{Name: "testdata/changes/Haiku", Body: []byte(intro)}
	h.save()
	p := &Page{Name: "testdata/changes/hotel",
		Body: []byte(`# Hotel room
White linen and white light
Wooden floor and painted walls
Home away from home

#Haiku #Poetry`)}
	p.notify()
	assert.FileExists(t, "changes.md")
	s, err := os.ReadFile("changes.md")
	assert.NoError(t, err)
	assert.Contains(t, string(s), line)
	s, err = os.ReadFile("testdata/changes/Haiku.md")
	assert.NoError(t, err)
	assert.Equal(t, intro + line, string(s))
	assert.NoFileExists(t, "testdata/changes/Poetry.md")
}

func TestChangesWithList(t *testing.T) {
	cleanup(t, "changes.md", "changes.md~")
	intro := "# Changes\n\nThis is a paragraph.\n\n"
	line := "* [a change](change)\n"
	d := "## " + time.Now().Format(time.DateOnly) + "\n"
	os.WriteFile("changes.md", []byte(intro+d+line), 0644)
	p := &Page{Name: "testdata/changes/alex", Body: []byte(`Hallo!`)}
	p.notify()
	s, err := os.ReadFile("changes.md")
	assert.NoError(t, err)
	new_line := "* [testdata/changes/alex](testdata/changes/alex)\n"
	// new line was added at the beginning of the list
	assert.Equal(t, intro+d+new_line+line, string(s))
}

func TestChangesWithOldList(t *testing.T) {
	cleanup(t, "changes.md", "changes.md~")
	intro := "# Changes\n\nThis is a paragraph.\n\n"
	line := "* [a change](change)\n"
	y := "## " + time.Now().Add(-24*time.Hour).Format(time.DateOnly) + "\n"
	d := "## " + time.Now().Format(time.DateOnly) + "\n"
	os.WriteFile("changes.md", []byte(intro+y+line), 0644)
	p := &Page{Name: "testdata/changes/alex", Body: []byte(`Hallo!`)}
	p.notify()
	s, err := os.ReadFile("changes.md")
	assert.NoError(t, err)
	new_line := "* [testdata/changes/alex](testdata/changes/alex)\n"
	// new line was added at the beginning of the list
	assert.Equal(t, intro+d+new_line+"\n"+y+line, string(s))
}

func TestChangesWithOldDisappearingListAtTheEnd(t *testing.T) {
	cleanup(t, "changes.md", "changes.md~")
	intro := "# Changes\n\nThis is a paragraph.\n\n"
	line := "* [a change](testdata/changes/alex)\n"
	y := "## " + time.Now().Add(-24*time.Hour).Format(time.DateOnly) + "\n"
	d := "## " + time.Now().Format(time.DateOnly) + "\n"
	os.WriteFile("changes.md", []byte(intro+y+line), 0644)
	p := &Page{Name: "testdata/changes/alex", Body: []byte(`Hallo!`)}
	p.notify()
	s, err := os.ReadFile("changes.md")
	assert.NoError(t, err)
	new_line := "* [testdata/changes/alex](testdata/changes/alex)\n"
	// new line was added at the beginning of the list, with the new date, and the old date disappeared
	assert.Equal(t, intro+d+new_line, string(s))
}

func TestChangesWithOldDisappearingListInTheMiddle(t *testing.T) {
	cleanup(t, "changes.md", "changes.md~")
	intro := "# Changes\n\nThis is a paragraph.\n\n"
	line := "* [a change](testdata/changes/alex)\n"
	other := "* [other change](testdata/changes/whatever)\n"
	yy := "## " + time.Now().Add(-48*time.Hour).Format(time.DateOnly) + "\n"
	y := "## " + time.Now().Add(-24*time.Hour).Format(time.DateOnly) + "\n"
	d := "## " + time.Now().Format(time.DateOnly) + "\n"
	os.WriteFile("changes.md", []byte(intro+y+line+yy+other), 0644)
	p := &Page{Name: "testdata/changes/alex", Body: []byte(`Hallo!`)}
	p.notify()
	s, err := os.ReadFile("changes.md")
	assert.NoError(t, err)
	new_line := "* [testdata/changes/alex](testdata/changes/alex)\n"
	// new line was added at the beginning of the list, with the new date, and the old date disappeared
	assert.Equal(t, intro+d+new_line+"\n"+yy+other, string(s))
}

func TestChangesWithListAtTheTop(t *testing.T) {
	cleanup(t, "testdata/changes", "changes.md", "changes.md~")
	line := "* [a change](change)\n"
	d := "## " + time.Now().Format(time.DateOnly) + "\n"
	os.WriteFile("changes.md", []byte(line), 0644)
	p := &Page{Name: "testdata/changes/alex", Body: []byte(`Hallo!`)}
	p.notify()
	s, err := os.ReadFile("changes.md")
	assert.NoError(t, err)
	new_line := "* [testdata/changes/alex](testdata/changes/alex)\n"
	// new line was added at the top, no error due to missing introduction
	assert.Equal(t, d+new_line+line, string(s))
}

func TestChangesWithNoList(t *testing.T) {
	cleanup(t, "testdata/changes", "changes.md", "changes.md~")
	intro := "# Changes\n\nThis is a paragraph."
	d := "## " + time.Now().Format(time.DateOnly) + "\n"
	os.WriteFile("changes.md", []byte(intro), 0644)
	p := &Page{Name: "testdata/changes/alex", Body: []byte(`Hallo!`)}
	p.notify()
	s, err := os.ReadFile("changes.md")
	assert.NoError(t, err)
	new_line := "* [testdata/changes/alex](testdata/changes/alex)\n"
	// into is still there and a new list was started
	assert.Equal(t, intro+"\n\n"+d+new_line, string(s))
}

func TestChangesWithUpdate(t *testing.T) {
	cleanup(t, "testdata/changes", "changes.md", "changes.md~")
	intro := "# Changes\n\nThis is a paragraph.\n\n"
	other := "* [other change](testdata/changes/whatever)\n"
	d := "## " + time.Now().Format(time.DateOnly) + "\n"
	line := "* [a change](testdata/changes/alex)\n"
	os.WriteFile("changes.md", []byte(intro+d+other+line), 0644)
	p := &Page{Name: "testdata/changes/alex", Body: []byte(`Hallo!`)}
	p.notify()
	s, err := os.ReadFile("changes.md")
	assert.NoError(t, err)
	new_line := "* [testdata/changes/alex](testdata/changes/alex)\n"
	// the change was already listed, but now it moved up and has a new title
	assert.Equal(t, intro+d+new_line+other, string(s))
}

func TestChangesWithNoChangeToTheOrder(t *testing.T) {
	cleanup(t, "testdata/changes", "changes.md", "changes.md~")
	intro := "# Changes\n\nThis is a paragraph.\n\n"
	other := "* [other change](testdata/changes/whatever)\n"
	line := "* [a change](testdata/changes/alex)\n"
	d := "## " + time.Now().Format(time.DateOnly) + "\n"
	os.WriteFile("changes.md", []byte(intro+d+line+other), 0644)
	p := &Page{Name: "testdata/changes/alex", Body: []byte(`Hallo!`)}
	p.notify()
	s, err := os.ReadFile("changes.md")
	assert.NoError(t, err)
	new_line := "* [testdata/changes/alex](testdata/changes/alex)\n"
	// the change was already listed at the top, so just use the new title
	assert.Equal(t, intro+d+new_line+other, string(s))
}

func TestChangesWithNoChanges(t *testing.T) {
	cleanup(t, "testdata/changes", "changes.md", "changes.md~")
	intro := "# Changes\n\nThis is a paragraph.\n\n"
	other := "* [other change](testdata/changes/whatever)\n"
	line := "* [a change](testdata/changes/alex)\n"
	d := "## " + time.Now().Format(time.DateOnly) + "\n"
	os.Remove("changes.md~")
	os.WriteFile("changes.md", []byte(intro+d+line+other), 0644)
	p := &Page{Name: "testdata/changes/alex", Body: []byte("# a change\nHallo!")}
	p.notify()
	s, err := os.ReadFile("changes.md")
	assert.NoError(t, err)
	// the change was already listed at the top, so no change was necessary
	assert.Equal(t, intro+d+line+other, string(s))
	// since the file hasn't changed, no backup was necessary
	assert.NoFileExists(t, "changes.md~")
}
