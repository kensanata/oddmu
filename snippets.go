package main

import (
	"strings"
	"regexp"
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

func snippets (q string, s string) (string, int) {
	// Look for Snippets
	snippetlen := 100
	maxsnippets := 4
	// Compile the query as a regular expression
	re, err := regexp.Compile("((?i)" + q + ")")
	// If the compilation didn't work, truncate
	if err != nil || len(s) <= snippetlen {
		if len(s) > 400 {
			s = s[0:400]
		}
		return highlight(q, s)
	}
	// show a snippet from the beginning of the document
	j := strings.LastIndex(s[:snippetlen], " ")
	if j == -1 {
		// OK, look for a longer word
		j = strings.Index(s, " ")
		if j == -1 {
			// Or just truncate the body.
			if len(s) > 400 {
				s = s[0:400]
			}
			return highlight(q, s)
		}
	}
	t := s[0:j]
	res := t + " … "
	s = s[j:] // avoid rematching
	jsnippet := 0
	for jsnippet < maxsnippets {
		m := re.FindStringSubmatch(s)
		if m == nil {
			break
		}
		jsnippet++
		j = strings.Index(s, m[1])
		if j > -1 {
			// get the substring containing the start of
			// the match, ending on word boundaries
			from := j - snippetlen / 2
			if from < 0 {
				from = 0
			}
			start := strings.Index(s[from:], " ")
			if start == -1 {
				start = 0
			} else {
				start += from
			}
			to := j + snippetlen / 2
			if to > len(s) {
				to = len(s)
			}
			end := strings.LastIndex(s[:to], " ")
			if end == -1 {
				// OK, look for a longer word
				end = strings.Index(s[to:], " ")
				if end == -1 {
					end = len(s)
				} else {
					end += to
				}
			}
			t = s[start : end];
			res = res + t + " … ";
			// truncate text to avoid rematching the same string.
			s = s[end:]
		}
	}
	return highlight(q, res)
}
