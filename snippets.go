package main

import (
	"regexp"
	"strings"
)

func snippets(q string, s string) (string, int) {
	// Look for Snippets
	snippetlen := 100
	maxsnippets := 4
	// Compile the query as a regular expression
	re, err := regexp.Compile("(?i)(" + strings.Join(strings.Split(q, " "), "|") + ")")
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
	res := t + " …"
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
			from := j - snippetlen/2
			if from < 0 {
				from = 0
			}
			start := strings.Index(s[from:], " ")
			if start == -1 {
				start = 0
			} else {
				start += from
			}
			to := j + snippetlen/2
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
			t = s[start:end]
			res = res + t + " …"
			// truncate text to avoid rematching the same string.
			s = s[end:]
		}
	}
	return highlight(q, res)
}
