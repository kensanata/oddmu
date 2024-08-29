package main

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

// relies on index.md in the current directory!
func TestListHandler(t *testing.T) {
	assert.Contains(t,
		assert.HTTPBody(makeHandler(listHandler, false), "GET", "/list/", nil),
		"index.md")
}

func TestDeleteHandler(t *testing.T) {
	cleanup(t, "testdata/delete")
	assert.NoError(t, os.Mkdir("testdata/delete", 0755))
	p := &Page{Name: "testdata/delete/haiku", Body: []byte(`# Sunset

Walk the fields outside
See the forest loom above
And an orange sky
`)}
	p.save()
	list := assert.HTTPBody(makeHandler(listHandler, false), "GET", "/list/testdata/delete/", nil)
	assert.Contains(t, list, `<a href="/view/testdata/delete/haiku.md">haiku.md</a>`)
	assert.Contains(t, list, `<td>Sunset</td>`)
	assert.Contains(t, list, `<button formaction="/delete/testdata/delete/haiku.md">Delete</button>`)
}
