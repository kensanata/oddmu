package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTokenizer(t *testing.T) {
	assert.EqualValues(t, []string{}, tokens(""), "empty string")
	assert.EqualValues(t, []string{"franc"}, tokens("Franc"), "lower case")
	assert.EqualValues(t, []string{"i", "don", "t", "know", "what", "to", "do"}, tokens("I don't know what to do."))
}

func TestHashtags(t *testing.T) {
	assert.EqualValues(t, []string{"#truth"}, hashtags([]byte("This is boring. #Truth")), "hashtags")
}

func TestTokensAndPredicates(t *testing.T) {
	predicates, terms := predicatesAndTokens("foo title:bar")
	assert.EqualValues(t, []string{"foo"}, terms)
	assert.EqualValues(t, []string{"title:bar"}, predicates)
}
