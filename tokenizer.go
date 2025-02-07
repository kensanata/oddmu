package main

import (
	"strings"
	"unicode"
)

// lowercaseFilter returns a slice of lower case tokens.
func lowercaseFilter(tokens []string) []string {
	r := make([]string, len(tokens))
	for i, token := range tokens {
		r[i] = strings.ToLower(token)
	}
	return r
}

// IsQuote reports whether the rune has the Quotation Mark property.
func IsQuote(r rune) bool {
	// This property isn't the same as Z; special-case it.
	return unicode.Is(unicode.Quotation_Mark, r)
}

// tokenizeWithQuotes returns a slice of tokens for the given text, including punctuation. Use this to begin tokenizing
// the query string. Note that quotation marks need a matching rune to end: 'foo' "foo" ‘foo’ ‚foo‘ ’foo’ “foo” „foo“
// ”foo” «foo» »foo« ‹foo› ›foo‹ ｢foo｣ 「ｆｏｏ」 『ｆｏｏ』 – read and despair:
// https://en.wikipedia.org/wiki/Quotation_mark
//
// Also note that 〈ｆｏｏ〉 and 《ｆｏｏ》 are not considered to be quotation marks by Unicode.
func tokenizeWithQuotes(s string) []string {
	type span struct {
		start int
		end   int
	}

	waitFor := rune(0)
	matchingRunes := [][]rune{{'\'', '\''}, {'"', '"'}, {'‘', '’'}, {'‚', '‘'}, {'’', '’'}, {'“', '”'}, {'„', '“'}, {'”', '”'},
		{'«', '»'}, {'»', '«'}, {'‹', '›'}, {'›', '‹'}, {'｢', '｣'}, {'「', '」'}, {'『', '』'}}

	spans := make([]span, 0, 32)

	// The comments in FieldsFunc say that doing this in a separate pass is faster.
	start := -1 // valid span start if >= 0
RUNE:
	for end, rune := range s {
		if waitFor > 0 {
			if rune == waitFor {
				if start >= 0 {
					// skip "" and the like
					spans = append(spans, span{start, end})
				}
				// The comments in FieldsFunc say that doing this instead of using -1 is faster.
				start = ^start
				waitFor = 0
			} else if start < 0 {
				start = end
			}
		} else if unicode.IsSpace(rune) {
			if start >= 0 {
				spans = append(spans, span{start, end})
				start = ^start
			}
		} else {
			if start < 0 {
				// Only check for starting quote at the beginning of a token
				if IsQuote(rune) {
					waitFor = rune
					for _, match := range matchingRunes {
						if rune == match[0] {
							waitFor = match[1]
							continue RUNE
						}
					}
				}
				start = end
			}
		}
	}

	// Last field might end at EOF.
	if start >= 0 {
		spans = append(spans, span{start, len(s)})
	}

	// Create strings from recorded field indices.
	a := make([]string, len(spans))
	for i, span := range spans {
		a[i] = s[span.start:span.end]
	}

	return a
}

// predicateFilter returns two slices of tokens: the first with predicates, the other without predicates. Use this for
// query string tokens.
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

// predicatesAndTokens returns two slices of tokens: the first with predicates, the other without predicates, all of
// them lower case. Use this for query strings.
func predicatesAndTokens(q string) ([]string, []string) {
	tokens := tokenizeWithQuotes(q)
	tokens = lowercaseFilter(tokens)
	return predicateFilter(tokens)
}

// noPredicateFilter returns a slice of tokens: the predicates without the predicate, and all the others. That is:
// "foo:bar baz" is turned into ["bar", "baz"] and the predicate "foo:" is dropped.
func noPredicateFilter(tokens []string) []string {
	r := make([]string, 0)
	for _, token := range tokens {
		parts := strings.Split(token, ":")
		r = append(r, parts[len(parts)-1])
	}
	return r
}

// highlightTokens returns the tokens to highlight, including title
// predicates.
func highlightTokens(q string) []string {
	tokens := tokenizeWithQuotes(q)
	tokens = lowercaseFilter(tokens)
	return noPredicateFilter(tokens)
}
