package main

import (
	"log"
	"regexp"
	"strings"
)

// re returns a regular expression matching any word in q.
func re(q string) (*regexp.Regexp, error) {
	fields := highlightTokens(q)
	quoted := make([]string, len(fields))
	for i, w := range fields {
		quoted[i] = regexp.QuoteMeta(w)
	}
	re, err := regexp.Compile(`(?i)(` + strings.Join(quoted, "|") + `)`)
	if err != nil {
		log.Printf("Cannot compile %s %v: %s", q, quoted, err)
		return nil, err
	}
	return re, nil
}

func snippets(q string, s string) string {
	// Look for Snippets
	snippetlen := 100
	maxsnippets := 4
	re, err := re(q)
	// If the compilation didn't work, truncate and return
	if err != nil {
		if len(s) > 400 {
			s = s[0:400] + " …"
		}
		return s
	}
	// Short cut for short pages
	if len(s) <= snippetlen {
		return highlight(re, s)
	}
	// show a snippet from the beginning of the document
	j := strings.LastIndex(s[:snippetlen], " ")
	if j == -1 {
		// OK, look for a longer word
		j = strings.Index(s, " ")
		if j == -1 {
			// Or just truncate the body.
			if len(s) > 400 {
				s = s[0:400] + " …"
			}
			return highlight(re, s)
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
		wl := len(m[1])
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
			to := j + wl + snippetlen/2
			if to > len(s) {
				to = len(s)
			}
			end := strings.LastIndex(s[:to], " ")
			if end == -1 || end <= j+wl {
				// OK, look for a longer word
				end = strings.Index(s[to:], " ")
				if end == -1 {
					end = len(s)
				} else {
					end += to
				}
			}
			t = s[start:end]
			res = res + t
			if len(s) > end {
				res = res + " …"
			}
			// truncate text to avoid rematching the same string.
			s = s[end:]
		}
	}
	return highlight(re, res)
}
