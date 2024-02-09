package main

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
 	"time"
)

func TestWatchedPageUpdate(t *testing.T) {
	dir := "testdata/watched-page"
	path := dir + "/haiku.md"
	cleanup(t, dir)
	index.load()
	watches.install()
	assert.NoError(t, os.MkdirAll(dir, 0755))
	time.Sleep(time.Millisecond)
	assert.Contains(t, watches.watcher.WatchList(), dir)

	haiku := []byte(`# Pine cones

Soft steps on the trail
Up and up in single file
Who ate half a cone?`)
	assert.NoError(t, os.WriteFile(path, haiku, 0644))

	time.Sleep(time.Millisecond)

	watches.RLock()
	assert.Contains(t, watches.files, path)
	watches.RUnlock()

	watches.Lock()
	watches.files[path] = watches.files[path].Add(-2 * time.Second)
	watches.Unlock()

	watches.watchTimer(path)

	index.RLock()
	assert.Contains(t, index.titles, path[:len(path)-3])
	index.RUnlock()
}

func TestWatchedTemplateUpdate(t *testing.T) {
	dir := "testdata/watched-template"
	name := dir + "/raclette"
	path := dir + "/view.html"
	cleanup(t, dir)
	index.load()
	watches.install()
	assert.NoError(t, os.MkdirAll(dir, 0755))

	time.Sleep(time.Millisecond)

	assert.Contains(t, watches.watcher.WatchList(), dir)

	p := &Page{Name: name, Body: []byte(`# Raclette

The heat element
glows red and the cheese bubbles
the smell is everywhere
`)}
	p.save()
	assert.Contains(t,
		assert.HTTPBody(makeHandler(viewHandler, true), "GET", "/view/testdata/watched-template/raclette", nil),
		"Skip navigation")
	
	// save a new view handler directly
	assert.NoError(t,
		os.WriteFile(path,
			[]byte("<body><h1>{{.Title}}</h1>{{.Html}}"),
			0644))

	time.Sleep(time.Millisecond)

	watches.RLock()
	assert.Contains(t, watches.files, path)
	watches.RUnlock()

	watches.Lock()
	watches.files[path] = watches.files[path].Add(-2 * time.Second)
	watches.Unlock()

	watches.watchTimer(path)
	
	body := assert.HTTPBody(makeHandler(viewHandler, true), "GET", "/view/" + name, nil)
	assert.Contains(t, body, "<h1>Raclette</h1>") // page text is still there
	assert.NotContains(t, body, "Skip") // but the header is not
}
