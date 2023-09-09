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

// score splits the query string q into terms and scores the text
// based on those terms. This assumes that q already has all its meta
// characters quoted.
func score(q string, s string) int {
	score := 0
	re, err := regexp.Compile("(?i)" + q)
	if err == nil {
		m := re.FindAllString(s, -1)
		if m != nil {
			// Score increases for each full match of q.
			score += len(m)
		}
	}
	for _, v := range strings.Split(q, " ") {
		if len(v) == 0 {
			continue
		}
		re, err := regexp.Compile(`(?is)(\pL?)(` + v + `)(\pL?)`)
		if err != nil {
			continue
		}
		for _, m := range re.FindAllStringSubmatch(s, -1) {
			// Term matched increases the score.
			score++
			// Terms matching at the beginning and
			// end of words and matching entire
			// words increase the score further.
			if len(m[1]) == 0 {
				score++
			}
			if len(m[3]) == 0 {
				score++
			}
			if len(m[1]) == 0 && len(m[3]) == 0 {
				score++
			}
		}
	}
	return score
}
