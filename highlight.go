package main

import (
	"regexp"
	"strings"
)

// highlight splits the query string q into terms and highlights them
// using the bold tag. Return the highlighted string.
// This assumes that q already has all its meta characters quoted.
func highlight(q string, s string) string {
	for _, v := range strings.Split(q, " ") {
		if len(v) == 0 {
			continue
		}
		re, err := regexp.Compile(`(?is)(` + v + `)`)
		if err != nil {
			continue
		}
		r := make(map[string]string)
		for _, m := range re.FindAllStringSubmatch(s, -1) {
			r[m[1]] = "<b>" + m[1] + "</b>"
		}
		// TODO: check for overlap?
		for old, new := range r {
			s = strings.ReplaceAll(s, old, new)
		}
	}
	return s
}
