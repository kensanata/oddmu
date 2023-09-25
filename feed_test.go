package main

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestFeed(t *testing.T) {
	assert.Contains(t,
		assert.HTTPBody(makeHandler(viewHandler, true), "GET", "/view/index.rss", nil),
		"Welcome to Oddµ")
}

// wipes testdata
func TestFeedItems(t *testing.T) {
	_ = os.RemoveAll("testdata")
	index.load()

	p1 := &Page{Name: "testdata/cactus", Body: []byte(`# Cactus
Green head and white hair
A bench in the evening sun
Unmoved by the news

#Succulent`)}
	p1.save()

	p2 := &Page{Name: "testdata/dragon", Body: []byte(`# Dragon
My palm tree grows straight
Up and up to touch the sky
Ignoring the roof

#Palmtree`)}
	p2.save()

	p3 := &Page{Name: "testdata/plants", Body: []byte(`# Plants
Writing poems about plants.

* [My Cactus](cactus)
* [My Dragon Tree](dragon)`)}
	p3.save()

	body := assert.HTTPBody(makeHandler(viewHandler, true), "GET", "/view/testdata/plants.rss", nil)
	assert.Contains(t, body, "<title>Plants</title>")
	assert.Contains(t, body, "<title>Cactus</title>")
	assert.Contains(t, body, "<title>Dragon</title>")
	assert.Contains(t, body, "&lt;h1&gt;Cactus&lt;/h1&gt;")
	assert.Contains(t, body, "&lt;h1&gt;Dragon&lt;/h1&gt;")
	assert.Contains(t, body, "<category>Succulent</category>")
	assert.Contains(t, body, "<category>Palmtree</category>")

	t.Cleanup(func() {
		_ = os.RemoveAll("testdata")
	})
}
