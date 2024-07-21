package main

import (
	"regexp"
)

// highlight matches for the regular expression using the bold tag.
func highlight(re *regexp.Regexp, s string) string {
	s = re.ReplaceAllString(s, "<b>$1</b>")
	return s
}
