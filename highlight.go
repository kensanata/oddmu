package main

import (
	"regexp"
	"strings"
)

// highlight splits the query string q into terms and highlights them
// using the bold tag. Return the highlighted string and a score.
func highlight(q string, s string) (string, int) {
	c := 0
	re, err := regexp.Compile("(?i)" + q)
	if err == nil {
		m := re.FindAllString(s, -1)
		if m != nil {
			// Score increases for each full match of q.
			c += len(m)
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
		r := make(map[string]string)
		for _, m := range re.FindAllStringSubmatch(s, -1) {
			// Term matched increases the score.
			c++
			// Terms matching at the beginning and
			// end of words and matching entire
			// words increase the score further.
			if len(m[1]) == 0 {
				c++
			}
			if len(m[3]) == 0 {
				c++
			}
			if len(m[1]) == 0 && len(m[3]) == 0 {
				c++
			}
			r[m[2]] = "<b>" + m[2] + "</b>"
		}
		for old, new := range r {
			s = strings.ReplaceAll(s, old, new)
		}
	}
	return s, c
}
