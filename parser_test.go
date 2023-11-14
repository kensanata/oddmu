package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

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

<p><a class="tag" href="/search/?q=%23Haiku">#Haiku</a> <a class="tag" href="/search/?q=%23Cold_Poets">#Cold Poets</a></p>
`
	assert.Equal(t, r, string(p.Html))
}

func TestPageHtmlWikiLink(t *testing.T) {
	p := &Page{Body: []byte(`# Photos and Books
Blue and green and black
Sky and grass and [ragged cliffs](cliffs)
Our [[time together]]`)}
	p.renderHtml()
	r := `<h1>Photos and Books</h1>

<p>Blue and green and black
Sky and grass and <a href="cliffs">ragged cliffs</a>
Our <a href="time%20together">time together</a></p>
`
	assert.Equal(t, r, string(p.Html))
}

func TestPageHtmlDollar(t *testing.T) {
	p := &Page{Body: []byte(`# No $dollar$ can buy this
Dragonfly hovers
darts chases turns lands and rests
A mighty jewel`)}
	p.renderHtml()
	r := `<h1>No $dollar$ can buy this</h1>

<p>Dragonfly hovers
darts chases turns lands and rests
A mighty jewel</p>
`
	assert.Equal(t, r, string(p.Html))
}

func TestLazyLoadImages(t *testing.T) {
	p := &Page{Body: []byte(`![](test.jpg)`)}
	p.renderHtml()
	assert.Contains(t, string(p.Html), "lazy")
}
