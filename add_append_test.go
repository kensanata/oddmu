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
	cleanup(t, "testdata/notification2", "changes.md", "changes.md~")
	restore(t, "index.md")
	os.Remove("changes.md")
	today := time.Now().Format(time.DateOnly)

	p := &Page{Name: "testdata/notification2/" + today + "-water", Body: []byte(`# Water
Sunlight dancing fast
Blue and green and pebbles gray
`)}
	p.save()

	data := url.Values{}
	data.Set("body", "Stand in cold water")
	data.Add("notify", "on")
	HTTPRedirectTo(t, makeHandler(appendHandler, true),
		"POST", "/append/testdata/notification2/"+today+"-water",
		data, "/view/testdata/notification2/"+today+"-water")
	// The changes.md file was created
	s, err := os.ReadFile("changes.md")
	assert.NoError(t, err)
	d := time.Now().Format(time.DateOnly)
	assert.Equal(t, "# Changes\n\n## "+d+"\n* [Water](testdata/notification2/"+today+"-water)\n", string(s))
	// Link added to index.md file
	s, err = os.ReadFile("index.md")
	assert.NoError(t, err)
	assert.Contains(t, string(s), "\n* [Water](testdata/notification2/"+today+"-water)\n")
}
