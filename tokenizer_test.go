package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHashtags(t *testing.T) {
	assert.EqualValues(t, []string{"#truth"}, hashtags([]byte("This is boring. #Truth")), "hashtags")
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
