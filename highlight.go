package main

import (
	"strings"
)

// highlight splits the query string q into terms and highlights them
// using the bold tag. Return the highlighted string and a score.
func highlight (q string, s string) (string, int) {
	c := strings.Count(s, q)
	for _, v := range strings.Split(q, " ") {
		if len(v) == 0 {
			continue
		}
		n := strings.Count(s, v)
		if n > 0 {
			c += n
			// Various ways to get higher scores: matches
			// at the beginning and end of words and
			// entire words.
			c += strings.Count(s, v + " ")
			c += strings.Count(s, " " + v)
			c += strings.Count(s, " " + v + " ")
			// Finally, add the bold tags after the final
			// score is tallied up.
			s = strings.ReplaceAll(s, v, "<b>" + v + "</b>")
		}
	}
	return s, c
}
