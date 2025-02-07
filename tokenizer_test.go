package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHashtags(t *testing.T) {
	assert.EqualValues(t, []string{"Truth"}, hashtags([]byte("This is boring. #Truth")), "hashtags")
}

func TestEscapedHashtags(t *testing.T) {
	assert.EqualValues(t, []string{}, hashtags([]byte("This is not a hashtag: \\#False")), "escaped hashtags")
}

func TestBorkedHashtags(t *testing.T) {
	assert.EqualValues(t, []string{}, hashtags([]byte("This is borked: \\#")), "borked hashtag")
}

func TestTokensAndPredicates(t *testing.T) {
	predicates, terms := predicatesAndTokens("foo title:bar")
	assert.EqualValues(t, []string{"foo"}, terms)
	assert.EqualValues(t, []string{"title:bar"}, predicates)
}

func TestQuoteRunes(t *testing.T) {
	s := `'"‘’‘‚“”„«»«‹›‹｢｣「」『』`
	for _, rune := range s {
		assert.True(t, IsQuote(rune), fmt.Sprintf("%c is a quote", rune))
	}
}

func TestQuotes(t *testing.T) {
	s := `'foo' "foo" ‘foo’ ‚foo‘ ’foo’ “foo” „foo“ ”foo” «foo» »foo« ‹foo› ›foo‹ ｢foo｣
「ｆｏｏ」 『ｆｏｏ』`
	tokens := tokenizeWithQuotes(s)
	assert.EqualValues(t, []string{
		"foo", "foo", "foo", "foo", "foo", "foo", "foo", "foo", "foo", "foo", "foo", "foo", "foo",
		"ｆｏｏ", "ｆｏｏ"}, tokens)
}

func TestPhrases(t *testing.T) {
	s := `look for 'foo bar'`
	tokens := tokenizeWithQuotes(s)
	assert.EqualValues(t, []string{"look", "for", "foo bar"}, tokens)
}

func TestKlingon(t *testing.T) {
	s := `quSDaq ba’lu’’a’`
	tokens := tokenizeWithQuotes(s)
	assert.EqualValues(t, []string{"quSDaq", "ba’lu’’a’"}, tokens)
	// quotes at the beginning of a word are not handled correctly
	s = `nuqDaq ‘oH tach’e’`
	tokens = tokenizeWithQuotes(s)
	assert.EqualValues(t, []string{"nuqDaq", "oH tach", "e’"}, tokens) // this is wrong 🤷
}
