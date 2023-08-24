package main

import (
	"strings"
	"testing"
	"os"
)

func TestPageTitle (t *testing.T) {
	p := &Page{Body: []byte(`# Ache
My back aches for you
I sit, stare and type for hours
But yearn for blue sky`)}
	p.handleTitle(false)
	if p.Title != "Ache" {
		t.Logf("The page title was not extracted correctly: %s", p.Title)
		t.Fail()
	}
	if !strings.HasPrefix(string(p.Body), "# Ache") {
		t.Logf("The page title was removed: %s", p.Body)
		t.Fail()
	}
	p.handleTitle(true)
	if !strings.HasPrefix(string(p.Body), "My back") {
		t.Logf("The page title was not removed: %s", p.Body)
		t.Fail()
	}
}

func TestPagePlainText (t *testing.T) {
	p := &Page{Body: []byte(`# Water
The air will not come
To inhale is an effort
The summer heat kills`)}
	s := p.plainText()
	r := "Water The air will not come To inhale is an effort The summer heat kills"
	if s != r {
		t.Logf("The plain text version is wrong: %s", s)
		t.Fail()
	}
}

func TestPageHtml (t *testing.T) {
	p := &Page{Body: []byte(`# Sun
Silver leaves shine bright
They droop, boneless, weak and sad
A cruel sun stares down`)}
	p.renderHtml()
	s := string(p.Html)
	r := `<h1>Sun</h1>

<p>Silver leaves shine bright
They droop, boneless, weak and sad
A cruel sun stares down</p>
`
	if s != r {
		t.Logf("The HTML is wrong: %s", s)
		t.Fail()
	}
}

func TestPageDir (t *testing.T) {
	_ = os.RemoveAll("testdata")
	loadIndex()
	p := &Page{Name: "testdata/moon", Body: []byte(`# Moon
From bed to bathroom
A slow shuffle in the dark
Moonlight floods the aisle`)}
	p.save()
	o, err := loadPage("testdata/moon")
	if err != nil || string(o.Body) != string(p.Body) {
		t.Logf("File in subdirectory not loaded: %s", p.Name)
		t.Fail()
	}
}
