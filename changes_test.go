package main

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

// Note TestEditSaveChanges and TestAddAppendChanges.

func TestChanges(t *testing.T) {
	cleanup(t, "changes.md")
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

func TestChangesWithList(t *testing.T) {
	cleanup(t, "changes.md")
	intro := "# Changes\n\nThis is a paragraph.\n\n"
	line := "* [a change](change)\n"
	os.WriteFile("changes.md", []byte(intro+line), 0644)
	p := &Page{Name: "testdata/changes/alex", Body: []byte(`Hallo!`)}
	p.notify()
	s, err := os.ReadFile("changes.md")
	assert.NoError(t, err)
	new_line := "* [testdata/changes/alex](testdata/changes/alex)\n"
	// new line was added at the beginning of the list
	assert.Equal(t, intro+new_line+line, string(s))
}

func TestChangesWithListAtTheTop(t *testing.T) {
	cleanup(t, "testdata/changes", "changes.md")
	line := "* [a change](change)\n"
	os.WriteFile("changes.md", []byte(line), 0644)
	p := &Page{Name: "testdata/changes/alex", Body: []byte(`Hallo!`)}
	p.notify()
	s, err := os.ReadFile("changes.md")
	assert.NoError(t, err)
	new_line := "* [testdata/changes/alex](testdata/changes/alex)\n"
	// new line was added at the top, no error due to missing introduction
	assert.Equal(t, new_line+line, string(s))
}

func TestChangesWithNoList(t *testing.T) {
	cleanup(t, "testdata/changes", "changes.md")
	intro := "# Changes\n\nThis is a paragraph."
	os.WriteFile("changes.md", []byte(intro), 0644)
	p := &Page{Name: "testdata/changes/alex", Body: []byte(`Hallo!`)}
	p.notify()
	s, err := os.ReadFile("changes.md")
	assert.NoError(t, err)
	new_line := "* [testdata/changes/alex](testdata/changes/alex)\n"
	// into is still there and a new list was started
	assert.Equal(t, intro+"\n\n"+new_line, string(s))
}

func TestChangesWithUpdate(t *testing.T) {
	cleanup(t, "testdata/changes", "changes.md")
	intro := "# Changes\n\nThis is a paragraph.\n\n"
	other := "* [other change](testdata/changes/whatever)\n"
	line := "* [a change](testdata/changes/alex)\n"
	os.WriteFile("changes.md", []byte(intro+other+line), 0644)
	p := &Page{Name: "testdata/changes/alex", Body: []byte(`Hallo!`)}
	p.notify()
	s, err := os.ReadFile("changes.md")
	assert.NoError(t, err)
	new_line := "* [testdata/changes/alex](testdata/changes/alex)\n"
	// the change was already listed, but now it moved up and has a new title
	assert.Equal(t, intro+new_line+other, string(s))
}

func TestChangesWithNoChangeToTheOrder(t *testing.T) {
	cleanup(t, "testdata/changes", "changes.md")
	intro := "# Changes\n\nThis is a paragraph.\n\n"
	other := "* [other change](testdata/changes/whatever)\n"
	line := "* [a change](testdata/changes/alex)\n"
	os.WriteFile("changes.md", []byte(intro+line+other), 0644)
	p := &Page{Name: "testdata/changes/alex", Body: []byte(`Hallo!`)}
	p.notify()
	s, err := os.ReadFile("changes.md")
	assert.NoError(t, err)
	new_line := "* [testdata/changes/alex](testdata/changes/alex)\n"
	// the change was already listed at the top, so just use the new title
	assert.Equal(t, intro+new_line+other, string(s))
}
