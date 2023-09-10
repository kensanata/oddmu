package main

import (
	"strings"
	"unicode"
)

// tokenize returns a slice of tokens for the given text.
func tokenize(text string) []string {
	return strings.FieldsFunc(text, func(r rune) bool {
		// Split on any character that is not a letter or a
		// number, not the hash sign (for hash tags)
		return !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != '#'
	})
}

// shortWordFilter removes all the words three characters or less
// except for all caps words like USA, EUR, CHF and the like.
func shortWordFilter(tokens []string) []string {
	r := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if len(token) > 3 ||
			len(token) == 3 && token == strings.ToUpper(token) {
			r = append(r, token)
		}
        }
	return r
}

// lowercaseFilter returns a slice of lower case tokens.
func lowercaseFilter(tokens []string) []string {
	r := make([]string, len(tokens))
	for i, token := range tokens {
		r[i] = strings.ToLower(token)
	}
	return r
}

// tokens returns a slice of tokens.
func tokens(text string) []string {
	tokens := tokenize(text)
	tokens = shortWordFilter(tokens)
	tokens = lowercaseFilter(tokens)
	return tokens
}
