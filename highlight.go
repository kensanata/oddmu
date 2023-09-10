package main

import (
	"regexp"
)

// highlight splits the query string q into terms and highlights them
// using the bold tag. Return the highlighted string.
// This assumes that q already has all its meta characters quoted.
func highlight(q string, re *regexp.Regexp, s string) string {
	s = re.ReplaceAllString(s, "<b>$1</b>")
	return s
}
