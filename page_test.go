package main

import (
	"os"
	"regexp"
	"testing"
	"github.com/stretchr/testify/assert"
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

func TestPagePlainText(t *testing.T) {
	p := &Page{Body: []byte(`# Water
The air will not come
To inhale is an effort
The summer heat kills`)}
	r := "Water The air will not come To inhale is an effort The summer heat kills"
	assert.Equal(t, r, p.plainText())
}

func TestPageHtml(t *testing.T) {
	p := &Page{Body: []byte(`# Sun
Silver leaves shine bright
They droop, boneless, weak and sad
A cruel sun stares down`)}
	p.renderHtml()
	r := `<h1>Sun</h1>

<p>Silver leaves shine bright
They droop, boneless, weak and sad
A cruel sun stares down</p>
`
	assert.Equal(t, r, string(p.Html))
}

func TestPageDir(t *testing.T) {
	_ = os.RemoveAll("testdata")
	loadIndex()
	p := &Page{Name: "testdata/moon", Body: []byte(`# Moon
From bed to bathroom
A slow shuffle in the dark
Moonlight floods the aisle`)}
	p.save()
	o, err := loadPage("testdata/moon")
	assert.NoError(t, err, "load page")
	assert.Equal(t, p.Body, o.Body)
	assert.FileExists(t, "testdata/moon.md")

	// Saving an empty page deletes it.
	p = &Page{Name: "testdata/moon", Body: []byte("")}
	p.save()
	assert.NoFileExists(t, "testdata/moon.md")
	t.Cleanup(func() {
		_ = os.RemoveAll("testdata")
	})
}

func TestLanguage(t *testing.T) {
	l := language(`
My back hurts at night
My shoulders won't budge today
Winter bones I say`)
	assert.Equal(t, "en", l)
}
