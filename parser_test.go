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
	r := `<h1 id="sun">Sun</h1>

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
	r := `<h1 id="comet">Comet</h1>

<p>Stars flicker above
Too faint to focus, so far
I am cold, alone</p>

<p><a class="tag" href="/search/?q=%23Haiku">#Haiku</a> <a class="tag" href="/search/?q=%23Cold_Poets">#Cold Poets</a></p>
`
	assert.Equal(t, r, string(p.Html))
}

func TestPageHtmlHashtagCornerCases(t *testing.T) {
	p := &Page{Body: []byte(`#

ok # #o #ok
[oh #ok \#nok](ok)`)}
	p.renderHtml()
	r := `<p>#</p>

<p>ok # <a class="tag" href="/search/?q=%23o">#o</a> <a class="tag" href="/search/?q=%23ok">#ok</a>
<a href="ok">oh #ok #nok</a></p>
`
	assert.Equal(t, r, string(p.Html))
}

func TestPageHtmlWikiLink(t *testing.T) {
	p := &Page{Body: []byte(`# Photos and Books
Blue and green and black
Sky and grass and [ragged cliffs](cliffs)
Our [[time together]]`)}
	p.renderHtml()
	r := `<h1 id="photos-and-books">Photos and Books</h1>

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
	r := `<h1 id="no-dollar-can-buy-this">No $dollar$ can buy this</h1>

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

// The fractions available in Latin 1 (?) are rendered.
func TestFractions(t *testing.T) {
	p := &Page{Body: []byte(`1/4`)}
	p.renderHtml()
	assert.Contains(t, string(p.Html), "&frac14;")
}

// Other fractions are not rendered.
func TestNoFractions(t *testing.T) {
	p := &Page{Body: []byte(`1/6`)}
	p.renderHtml()
	assert.Contains(t, string(p.Html), "1/6")
}

// webfinger
func TestAt(t *testing.T) {
	// enable webfinger
	useWebfinger = true
	// prevent lookups
	accounts.Lock()
	accounts.uris = make(map[string]string)
	accounts.uris["alex@alexschroeder.ch"] = "https://social.alexschroeder.ch/@alex";
	accounts.Unlock()
	// test account
	p := &Page{Body: []byte(`My fedi handle is @alex@alexschroeder.ch.`)}
	p.renderHtml()
	assert.Contains(t,string(p.Html),
		`My fedi handle is <a class="account" href="https://social.alexschroeder.ch/@alex" title="@alex@alexschroeder.ch">@alex</a>.`)
	// test escaped account
	p = &Page{Body: []byte(`My fedi handle is \@alex@alexschroeder.ch. \`)}
	p.renderHtml()
	assert.Contains(t,string(p.Html),
		`My fedi handle is @alex@alexschroeder.ch.`)
	// disable webfinger
	useWebfinger = false
}
