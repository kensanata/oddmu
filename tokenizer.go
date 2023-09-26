package main

import (
	"bytes"
	"strings"
	"unicode"
	"unicode/utf8"
)

// tokenize returns a slice of alphanumeric tokens for the given text.
// Use this to begin tokenizing the page body.
func tokenize(text string) []string {
	return strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
}

// lowercaseFilter returns a slice of lower case tokens.
func lowercaseFilter(tokens []string) []string {
	r := make([]string, len(tokens))
	for i, token := range tokens {
		r[i] = strings.ToLower(token)
	}
	return r
}

// tokens returns a slice of alphanumeric tokens.
func tokens(text string) []string {
	tokens := tokenize(text)
	tokens = lowercaseFilter(tokens)
	return tokens
}

// tokenizeWithFilters returns a slice of alphanumeric tokens for the
// given text, including colons. Use this to begin tokenizing the
// query string.
func tokenizeWithFilters(text string) []string {
	return strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != ':'
	})
}

// predicateFilter returns two slices of tokens: the first with
// predicates, the other without predicates. Use this for query
// string tokens.
func predicateFilter(tokens []string) ([]string, []string) {
	with := make([]string, 0)
	without := make([]string, 0)
	for _, token := range tokens {
		if strings.Contains(token, ":") {
			with = append(with, token)
		} else {
			without = append(without, token)
		}
	}
	return with, without
}

// predicatesAndTokens returns two slices of tokens: the first with
// predicates, the other without predicates, all of them lower case.
// Use this for query strings.
func predicatesAndTokens(q string) ([]string, []string) {
	tokens := tokenizeWithFilters(q)
	tokens = lowercaseFilter(tokens)
	return predicateFilter(tokens)
}

// hashtags returns a slice of hashtags. Use this to extract hashtags
// from a page body.
func hashtags(s []byte) []string {
	hashtags := make([]string, 0)
	for {
		i := bytes.IndexRune(s, '#')
		if i == -1 {
			return hashtags
		}
		from := i
		i++
		for {
			r, n := utf8.DecodeRune(s[i:])
			if n > 0 && (unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_') {
				i += n
			} else {
				break
			}
		}
		if i > from+1 { // not just "#"
			hashtags = append(hashtags, string(bytes.ToLower(s[from:i])))
		}
		s = s[i:]
	}
}
