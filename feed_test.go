package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"net/url"
)

func TestFeed(t *testing.T) {
	assert.Contains(t,
		assert.HTTPBody(makeHandler(viewHandler, false, http.MethodGet), "GET", "/view/index.rss", nil),
		"Welcome to Oddμ")
}

func TestNoFeed(t *testing.T) {
	assert.HTTPStatusCode(t,
		makeHandler(viewHandler, false, http.MethodGet), "GET", "/view/no-feed.rss", nil, http.StatusNotFound)
}

func TestFeedItems(t *testing.T) {
	cleanup(t, "testdata/feed")

	p1 := &Page{Name: "testdata/feed/cactus", Body: []byte(`# Cactus
Green head and white hair
A bench in the evening sun
Unmoved by the news

#Succulent`)}
	p1.save()

	p2 := &Page{Name: "testdata/feed/dragon", Body: []byte(`# Dragon
My palm tree grows straight
Up and up to touch the sky
Ignoring the roof

#Palmtree`)}
	p2.save()

	p3 := &Page{Name: "testdata/feed/plants", Body: []byte(`# Plants
Writing poems about plants.

* [My Cactus](cactus)
* [My Dragon Tree](dragon)`)}
	p3.save()

	body := assert.HTTPBody(makeHandler(viewHandler, false, http.MethodGet), "GET", "/view/testdata/feed/plants.rss", nil)
	assert.Contains(t, body, "<title>Plants</title>")
	assert.Contains(t, body, "<title>Cactus</title>")
	assert.Contains(t, body, "<title>Dragon</title>")
	assert.Contains(t, body, "&lt;h1 id=&#34;cactus&#34;&gt;Cactus&lt;/h1&gt;")
	assert.Contains(t, body, "&lt;h1 id=&#34;dragon&#34;&gt;Dragon&lt;/h1&gt;")
	assert.Contains(t, body, "<category>Succulent</category>")
	assert.Contains(t, body, "<category>Palmtree</category>")
}


func TestFeedPagination(t *testing.T) {
	cleanup(t, "testdata/pagination")

	p := &Page{Name: "testdata/pagination/one", Body: []byte("# One\n")}; p.save()
	p = &Page{Name: "testdata/pagination/two", Body: []byte("# Two\n")}; p.save()
	p = &Page{Name: "testdata/pagination/three", Body: []byte("# Three\n")}; p.save()
	p = &Page{Name: "testdata/pagination/four", Body: []byte("# Four\n")}; p.save()
	p = &Page{Name: "testdata/pagination/five", Body: []byte("# Five\n")}; p.save()
	p = &Page{Name: "testdata/pagination/six", Body: []byte("# Six\n")}; p.save()
	p = &Page{Name: "testdata/pagination/seven", Body: []byte("# Seven\n")}; p.save()
	p = &Page{Name: "testdata/pagination/eight", Body: []byte("# Eight\n")}; p.save()
	p = &Page{Name: "testdata/pagination/nine", Body: []byte("# Nine\n")}; p.save()
	p = &Page{Name: "testdata/pagination/ten", Body: []byte("# Ten\n")}; p.save()

	p = &Page{Name: "testdata/pagination/index", Body: []byte(`# Index
* [one](one)
* [two](two)
* [three](three)
* [four](four)
* [five](five)
* [six](six)
* [seven](seven)
* [eight](eight)
* [nine](nine)
* [ten](ten)
`)}
	p.save()

	body := assert.HTTPBody(makeHandler(viewHandler, false, http.MethodGet), "GET", "/view/testdata/pagination/index.rss", nil)
	assert.Contains(t, body, "<title>One</title>")
	assert.Contains(t, body, "<title>Ten</title>")
	assert.NotContains(t, body, `<atom:link href="https://example.org/view/testdata/pagination/index.rss?from=10&n=10" rel="next" type="application/rss+xml"/>`)

	p = &Page{Name: "testdata/pagination/eleven", Body: []byte("# Eleven\n")}; p.save()
	p = &Page{Name: "testdata/pagination/index", Body: []byte(`# Index
* [one](one)
* [two](two)
* [three](three)
* [four](four)
* [five](five)
* [six](six)
* [seven](seven)
* [eight](eight)
* [nine](nine)
* [ten](ten)
* [eleven](eleven)
`)}
	p.save()

	body = assert.HTTPBody(makeHandler(viewHandler, false, http.MethodGet), "GET", "/view/testdata/pagination/index.rss", nil)
	assert.NotContains(t, body, "<title>Eleven</title>")
	assert.Contains(t, body, `<atom:link href="https://example.org/view/testdata/pagination/index.rss?from=10&amp;n=10" rel="next" type="application/rss+xml"/>`)

	params := url.Values{}
	params.Set("n", "0")
	body = assert.HTTPBody(makeHandler(viewHandler, false, http.MethodGet), "GET", "/view/testdata/pagination/index.rss", params)
	assert.Contains(t, body, "<title>Eleven</title>")
	assert.Contains(t, body, `<fh:complete/>`)

	params = url.Values{}
	params.Set("n", "3")
	body = assert.HTTPBody(makeHandler(viewHandler, false, http.MethodGet), "GET", "/view/testdata/pagination/index.rss", params)
	assert.Contains(t, body, "<title>One</title>")
	assert.Contains(t, body, "<title>Three</title>")
	assert.NotContains(t, body, "<title>Four</title>")
	assert.Contains(t, body, `<atom:link href="https://example.org/view/testdata/pagination/index.rss?from=3&amp;n=3" rel="next" type="application/rss+xml"/>`)

	params = url.Values{}
	params.Set("from", "3")
	params.Set("n", "3")
	body = assert.HTTPBody(makeHandler(viewHandler, false, http.MethodGet), "GET", "/view/testdata/pagination/index.rss", params)
	assert.NotContains(t, body, "<title>Three</title>")
	assert.Contains(t, body, "<title>Four</title>")
	assert.Contains(t, body, "<title>Six</title>")
	assert.NotContains(t, body, "<title>Seven</title>")
	assert.Contains(t, body, `<atom:link href="https://example.org/view/testdata/pagination/index.rss?from=0&amp;n=3" rel="previous" type="application/rss+xml"/>`)
	assert.Contains(t, body, `<atom:link href="https://example.org/view/testdata/pagination/index.rss?from=6&amp;n=3" rel="next" type="application/rss+xml"/>`)

	params = url.Values{}
	params.Set("from", "2")
	params.Set("n", "3")
	body = assert.HTTPBody(makeHandler(viewHandler, false, http.MethodGet), "GET", "/view/testdata/pagination/index.rss", params)
	assert.NotContains(t, body, "<title>Two</title>")
	assert.Contains(t, body, "<title>Three</title>")
	assert.Contains(t, body, "<title>Five</title>")
	assert.NotContains(t, body, "<title>Six</title>")
	assert.Contains(t, body, `<atom:link href="https://example.org/view/testdata/pagination/index.rss?from=0&amp;n=3" rel="previous" type="application/rss+xml"/>`)
	assert.Contains(t, body, `<atom:link href="https://example.org/view/testdata/pagination/index.rss?from=5&amp;n=3" rel="next" type="application/rss+xml"/>`)
}
