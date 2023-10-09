package main

import (
	"github.com/stretchr/testify/assert"
	"net/url"
	"os"
	"regexp"
	"testing"
)

func TestAddAppend(t *testing.T) {
	cleanup(t, "testdata/add")
	index.load()

	p := &Page{Name: "testdata/add/fire", Body: []byte(`# Fire
Orange sky above
Reflects a distant fire
It's not `)}
	p.save()

	data := url.Values{}
	data.Set("body", "barbecue")

	assert.Regexp(t, regexp.MustCompile("a distant fire"),
		assert.HTTPBody(makeHandler(viewHandler, true),
			"GET", "/view/testdata/add/fire", nil))
	assert.NotRegexp(t, regexp.MustCompile("a distant fire"),
		assert.HTTPBody(makeHandler(addHandler, true),
			"GET", "/add/testdata/add/fire", nil))
	HTTPRedirectTo(t, makeHandler(appendHandler, true),
		"POST", "/append/testdata/add/fire", data, "/view/testdata/add/fire")
	assert.Regexp(t, regexp.MustCompile("It’s not barbecue"),
		assert.HTTPBody(makeHandler(viewHandler, true),
			"GET", "/view/testdata/add/fire", nil))
}

func TestAddAppendNotificationNoChanges(t *testing.T) {
	cleanup(t, "testdata/notification2", "changes.md")
	os.Remove("changes.md")
	p := &Page{Name: "testdata/notification2/water", Body: []byte(`# Water
Sunlight dancing fast
Blue and green and pebbles gray
`)}
	p.save()

	data := url.Values{}
	data.Set("body", "Stand in cold water")
	data.Add("notify", "on")
	HTTPRedirectTo(t, makeHandler(appendHandler, true),
		"POST", "/append/testdata/notification2/water", data, "/view/testdata/notification2/water")
	// The changes.md file was created
	s, err := os.ReadFile("changes.md")
	assert.NoError(t, err)
	assert.Equal(t, "# Changes\n\n* [Water](testdata/notification2/water)\n", string(s))
}
