package main

import (
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

func TestPageTitle(t *testing.T) {
	p := &Page{Body: []byte(`# Ache
My back aches for you
I sit, stare and type for hours
But yearn for blue sky`)}
	p.handleTitle(false)
	assert.Equal(t, "Ache", p.Title)
	assert.Regexp(t, regexp.MustCompile("^# Ache"), string(p.Body))
	p.handleTitle(true)
	assert.Regexp(t, regexp.MustCompile("^My back"), string(p.Body))
}

func TestPageDir(t *testing.T) {
	cleanup(t, "testdata/dir")
	index.load()
	p := &Page{Name: "testdata/dir/moon", Body: []byte(`# Moon
From bed to bathroom
A slow shuffle in the dark
Moonlight floods the aisle`)}
	p.save()

	o, err := loadPage("testdata/dir/moon")
	assert.NoError(t, err, "load page")
	assert.Equal(t, p.Body, o.Body)
	assert.FileExists(t, "testdata/dir/moon.md")

	// Saving an empty page deletes it.
	p = &Page{Name: "testdata/dir/moon", Body: []byte("")}
	p.save()
	assert.NoFileExists(t, "testdata/dir/moon.md")

	// But the backup still exists.
	assert.FileExists(t, "testdata/dir/moon.md~")
}

func TestPageParents(t *testing.T) {
	cleanup(t, "testdata/parents")
	index.load()
	p := &Page{Name: "testdata/parents/index", Body: []byte(`# Solar
The air dances here
Water puddles flicker and
disappear anon`)}
	p.save()
	p = &Page{Name: "testdata/parents/children/index", Body: []byte(`# Lunar
Behind running clouds
Shines cold light from ages past
And untouchable`)}
	p.save()
	p = &Page{Name: "testdata/parents/children/something/other"}
	// "testdata/parents/children/something/index" is a sibling and doesn't count!
	parents := p.Parents()
	assert.Equal(t, "Welcome to Oddμ", parents[0].Title)
	assert.Equal(t, "../../../../index", parents[0].Url)
	assert.Equal(t, "…", parents[1].Title)
	assert.Equal(t, "../../../index", parents[1].Url)
	assert.Equal(t, "Solar", parents[2].Title)
	assert.Equal(t, "../../index", parents[2].Url)
	assert.Equal(t, "Lunar", parents[3].Title)
	assert.Equal(t, "../index", parents[3].Url)
	assert.Equal(t, 4, len(parents))
}
