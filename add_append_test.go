package main

import (
	"github.com/stretchr/testify/assert"
	"net/url"
	"os"
	"regexp"
	"testing"
	"time"
)

func TestEmptyLineAdd(t *testing.T) {
	p := &Page{Name: "testdata/add/fire", Body: []byte(`# Coal
Black rocks light as foam
Shaking, puring, shoveling`)}
	p.append([]byte("Into the oven"))
	assert.Equal(t, string(p.Body), `# Coal
Black rocks light as foam
Shaking, puring, shoveling

Into the oven`)
}

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
	assert.Regexp(t, regexp.MustCompile(`not</p>\s*<p>barbecue`),
		assert.HTTPBody(makeHandler(viewHandler, true),
			"GET", "/view/testdata/add/fire", nil))
}

func TestAddAppendChanges(t *testing.T) {
	cleanup(t, "testdata/append")
	today := time.Now().Format(time.DateOnly)
	p := &Page{Name: "testdata/append/" + today + "-water", Body: []byte(`# Water
Sunlight dancing fast
Blue and green and pebbles gray
`)}
	p.save()
	data := url.Values{}
	data.Set("body", "Stand in cold water")
	data.Add("notify", "on")
	HTTPRedirectTo(t, makeHandler(appendHandler, true),
		"POST", "/append/testdata/append/"+today+"-water",
		data, "/view/testdata/append/"+today+"-water")
	// The changes.md file was created
	s, err := os.ReadFile("testdata/append/changes.md")
	assert.NoError(t, err)
	assert.Equal(t, "# Changes\n\n## "+today+"\n* [Water]("+today+"-water)\n", string(s))
	// Link added to index.md file
	s, err = os.ReadFile("testdata/append/index.md")
	assert.NoError(t, err)
	// New index contains just the link
	assert.Equal(t, string(s), "* [Water]("+today+"-water)\n")
}
