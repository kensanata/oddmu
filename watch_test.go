package main

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
 	"time"
)

func TestWatchedPageUpdate(t *testing.T) {
	cleanup(t, "testdata/watched-page")
	index.load()
	watches.install()
	assert.NoError(t, os.MkdirAll("testdata/watched-page", 0755))
	time.Sleep(time.Millisecond)
	assert.Contains(t, watches.watcher.WatchList(), "testdata/watched-page")

	haiku := []byte(`Pine cones

Soft steps on the trail
Up and up in single file
Who ate half a cone?`)
	assert.NoError(t, os.WriteFile("testdata/watched-page/haiku.md", haiku, 0644))

	time.Sleep(time.Millisecond)

	watches.RLock()
	assert.Contains(t, watches.files, "testdata/watched-page/haiku.md")
	watches.RUnlock()

	watches.Lock()
	watches.files["testdata/watched-page/haiku.md"] = watches.files["testdata/watched-page/haiku.md"].Add(-2 * time.Second)
	watches.Unlock()

	watches.watchTimer()

	index.RLock()
	assert.Contains(t, index.titles, "testdata/watched-page/haiku")
	index.RUnlock()
}

func TestWatchedTemplateUpdate(t *testing.T) {
	cleanup(t, "testdata/watched-template")
	index.load()
	watches.install()
	assert.NoError(t, os.MkdirAll("testdata/watched-template", 0755))

	time.Sleep(time.Millisecond)

	assert.Contains(t, watches.watcher.WatchList(), "testdata/watched-template")

	p := &Page{Name: "testdata/watched-template/raclette", Body: []byte(`# Raclette

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
		os.WriteFile("testdata/watched-template/view.html",
			[]byte("<body><h1>{{.Title}}</h1>{{.Html}}"),
			0644))

	time.Sleep(time.Millisecond)

	watches.RLock()
	assert.Contains(t, watches.files, "testdata/watched-template/view.html")
	watches.RUnlock()

	watches.Lock()
	watches.files["testdata/watched-template/view.html"] = watches.files["testdata/watched-template/view.html"].Add(-2 * time.Second)
	watches.Unlock()

	watches.watchTimer()
	
	body := assert.HTTPBody(makeHandler(viewHandler, true), "GET", "/view/testdata/watched-template/raclette", nil)
	assert.Contains(t, body, "<h1>Raclette</h1>") // page text is still there
	assert.NotContains(t, body, "Skip") // but the header is not
}
