package main

import (
	"regexp"
)

func (p *Page) notify() error {
	c, err := loadPage("changes")
	p.handleTitle(false)
	if p.Title == "" {
		p.Title = p.Name
	}
	esc := nameEscape(p.Name)
	if err != nil {
		// create a new page
		c = &Page{Name: "changes", Body: []byte("# Changes\n\n* [" + p.Title + "](" + esc + ")\n")}
	} else {
		// remove the old match, if one exists
		re := regexp.MustCompile(`(?m)^\* \[[^\]]+\]\(` + esc + `\)\n`)
		loc := re.FindIndex(c.Body)
		if loc != nil {
			r := c.Body[:loc[0]]
			if loc[1] < len(c.Body) {
				r = append(r, c.Body[loc[1]:]...)
			}
			c.Body = r
		}
		// locate the beginning of the list to insert the line
		re = regexp.MustCompile(`(?m)^\* \[[^\]]+\]\([^\)]+\)\n`)
		loc = re.FindIndex(c.Body)
		if loc == nil {
			// if no list was found, use the end of the page
			loc = []int{len(c.Body)}
		}
		r := []byte("")
		r = append(r, c.Body[:loc[0]]...)
		if len(r) > 0 && r[len(r)-1] != '\n' {
			r = append(r, '\n')
		}
		if len(r) > 1 && r[len(r)-2] != '\n' {
			r = append(r, '\n')
		}
		r = append(r, []byte("* ["+p.Title+"]("+esc+")\n")...)
		r = append(r, c.Body[loc[0]:]...)
		c.Body = r
	}
	return c.save()
}
