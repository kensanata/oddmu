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

func TestListDot(t *testing.T) {
	cleanup(t, "testdata/list-dot")
	p := &Page{Name: "testdata/list-dot/haiku", Body: []byte(`# Pressure

fingers tap and dance
round and round they go at night
before we go to bed
`)}
	p.save()
	_, err := os.Create("testdata/list-dot/.secret")
	assert.NoError(t, err)
	body := assert.HTTPBody(makeHandler(listHandler, false), "GET", "/list/testdata/list-dot/", nil)
	assert.NotContains(t, body, "secret", "secret file was not found")
	assert.Contains(t, body, "haiku", "regular page was found")
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
	body := assert.HTTPBody(makeHandler(listHandler, false), "GET", "/list/testdata/delete/", nil)
	assert.Contains(t, body, `<a href="/view/testdata/delete/haiku.md">haiku.md</a>`)
	assert.Contains(t, body, `<td>Sunset</td>`)
	assert.Contains(t, body, `<button form="manage" formaction="/delete/testdata/delete/haiku.md" title="Delete haiku.md">`)
	// ensure that it exists
	assert.FileExists(t, "testdata/delete/haiku.md")
	// delete file
	HTTPRedirectTo(t, makeHandler(deleteHandler, false), "GET", "/delete/testdata/delete/haiku.md", nil, "/list/testdata/delete/")
	// verify that it is gone
	body = assert.HTTPBody(makeHandler(listHandler, false), "GET", "/list/testdata/delete/", nil)
	assert.NotContains(t, body, `<a href="/view/testdata/delete/haiku.md">haiku.md</a>`)
	assert.NoFileExists(t, "testdata/delete/haiku.md")
}

func TestListUmlautHandler(t *testing.T) {
	cleanup(t, "testdata/list-umlaut")
	p := &Page{Name: "testdata/list-umlaut/hägar", Body: []byte(`# Hägar

Hägar was a man
Loud and strong and quick to act
he did not like it
`)}
	p.save()
	body := assert.HTTPBody(makeHandler(listHandler, false), "GET", "/list/testdata/list-umlaut/", nil)
	assert.Contains(t, body, `<button form="manage" formaction="/delete/testdata/list-umlaut/h%c3%a4gar.md" title="Delete hägar.md">`)
}

func TestListHash(t *testing.T) {
	cleanup(t, "testdata/list-hash")
	os.Mkdir("testdata/list-hash", 0755)
	_, err := os.Create("testdata/list-hash/#secret")
	assert.NoError(t, err)
	body := assert.HTTPBody(makeHandler(listHandler, false), "GET", "/list/testdata/list-hash/", nil)
	assert.Contains(t, body, `<button form="manage" formaction="/delete/testdata/list-hash/%23secret" title="Delete #secret">`)
}
