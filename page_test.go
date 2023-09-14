package main

import (
	"github.com/stretchr/testify/assert"
	"os"
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

func TestPageHtmlHashtag(t *testing.T) {
	p := &Page{Body: []byte(`# Comet
Stars flicker above
Too faint to focus, so far
I am cold, alone

#Haiku #Cold_Poets`)}
	p.renderHtml()
	r := `<h1>Comet</h1>

<p>Stars flicker above
Too faint to focus, so far
I am cold, alone</p>

<p><a href="/search?q=%23Haiku" rel="nofollow">#Haiku</a> <a href="/search?q=%23Cold_Poets" rel="nofollow">#Cold Poets</a></p>
`
	assert.Equal(t, r, string(p.Html))
}

// wipes testdata
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

	// But the backup still exists.
	assert.FileExists(t, "testdata/moon.md~")
	
	t.Cleanup(func() {
		_ = os.RemoveAll("testdata")
	})
}
