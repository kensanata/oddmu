package main

import (
	"regexp"
)

// score splits the query string q into terms and scores the text
// based on those terms. This assumes that q already has all its meta
// characters quoted.
func score(q string, s string) int {
	score := 0
	re, err := regexp.Compile("(?i)" + regexp.QuoteMeta(q))
	if err == nil {
		m := re.FindAllString(s, -1)
		if m != nil {
			// Score increases for each full match of q.
			score += len(m)
		}
	}
	for _, token := range highlightTokens(q) {
		re, err := regexp.Compile(`(?is)(\pL?)(` + regexp.QuoteMeta(token) + `)(\pL?)`)
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
