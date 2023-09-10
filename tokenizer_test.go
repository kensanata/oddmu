package main

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestTokenizer(t *testing.T) {
	assert.EqualValues(t, []string{}, tokens(""), "empty string")
	assert.EqualValues(t, []string{}, tokens("the a"), "no short words")
	assert.EqualValues(t, []string{"chf"}, tokens("CHF"), "three letter acronyms")
	assert.EqualValues(t, []string{}, tokens("CH"), "no two letter acronyms")
	assert.EqualValues(t, []string{"franc"}, tokens("Franc"), "lower case")
	assert.EqualValues(t, []string{"know", "what"}, tokens("I don't know what to do."))
}
