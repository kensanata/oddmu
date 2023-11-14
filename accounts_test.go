package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// This causes network access!
// func TestPageAccount(t *testing.T) {
// 	initAccounts()
// 	p := &Page{Body: []byte(`@alex, @alex@alexschroeder.ch said`)}
// 	p.renderHtml()
// 	r := `<p>@alex, <a href="https://alexschroeder.ch/users/alex">@alex</a> said</p>
// `
// 	assert.Equal(t, r, string(p.Html))
// }

func TestWebfingerParsing(t *testing.T) {
	body := []byte(`{
  "subject": "acct:Gargron@mastodon.social",
  "aliases": [
    "https://mastodon.social/@Gargron",
    "https://mastodon.social/users/Gargron"
  ],
  "links": [
    {
      "rel": "http://webfinger.net/rel/profile-page",
      "type": "text/html",
      "href": "https://mastodon.social/@Gargron"
    },
    {
      "rel": "self",
      "type": "application/activity+json",
      "href": "https://mastodon.social/users/Gargron"
    },
    {
      "rel": "http://ostatus.org/schema/1.0/subscribe",
      "template": "https://mastodon.social/authorize_interaction?uri={uri}"
    }
  ]
}`)
	uri, err := parseWebFinger(body)
	assert.NoError(t, err)
	assert.Equal(t, "https://mastodon.social/@Gargron", uri)
}
