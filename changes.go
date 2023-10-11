package main

import (
	"regexp"
	"time"
)

func (p *Page) notify() error {
	c, err := loadPage("changes")
	p.handleTitle(false)
	if p.Title == "" {
		p.Title = p.Name
	}
	esc := nameEscape(p.Name)
	d := time.Now().Format(time.DateOnly)
	if err != nil {
		// create a new page
		c = &Page{Name: "changes", Body: []byte("# Changes\n\n# " + d + "\n* [" + p.Title + "](" + esc + ")\n")}
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
			if loc[0] >= 13 && len(c.Body) >= loc[0]+13 {
				// remove the preceding date if there are now two dates following each other
				re := regexp.MustCompile(`(?m)^# (\d\d\d\d-\d\d-\d\d)\n# (\d\d\d\d-\d\d-\d\d)\n`)
				if re.Match(c.Body[loc[0]-13 : loc[0]+13]) {
					c.Body = append(c.Body[0 : loc[0]-13], c.Body[loc[0] : ]...)
				}
			} else if len(c.Body) == loc[0] {
				// remove a trailing date
				re := regexp.MustCompile(`# (\d\d\d\d-\d\d-\d\d)\n`)
				if re.Match(c.Body[loc[0]-13 : loc[0]]) {
					c.Body = c.Body[0 : loc[0]-13]
				}
			}
		}
		// locate the beginning of the list to insert the line
		re = regexp.MustCompile(`(?m)^\* \[[^\]]+\]\([^\)]+\)\n`)
		loc = re.FindIndex(c.Body)
		if loc == nil {
			// if no list was found, use the end of the page
			loc = []int{len(c.Body)}
		}
		// start with new page content
		r := []byte("")
		// check if there is a date right before the insertion point
		addDate := true
		if loc[0] >= 13 {
			re := regexp.MustCompile(`(?m)^# (\d\d\d\d-\d\d-\d\d)\n`) // 13 characters
			m := re.Find(c.Body[loc[0]-13 : loc[0]])
			if m == nil {
				// not a date: insert date, don't move insertion point
			} else if string(c.Body[loc[0]-11 : loc[0]-1]) == d {
				// if the date is our date, don't add it, don't move insertion point
				addDate = false
			} else {
				// if the date is not out date, move the insertion point
				loc[0] -= 13
			}
		}
		// append up to the insertion point
		r = append(r, c.Body[:loc[0]]...)
		// append date, if necessary
		if addDate {
			// ensure paragraph break
			if len(r) > 0 && r[len(r)-1] != '\n' {
				r = append(r, '\n')
			}
			if len(r) > 1 && r[len(r)-2] != '\n' {
				r = append(r, '\n')
			}
			r = append(r, []byte("# ")...)
			r = append(r, []byte(d)...)
			r = append(r, '\n')
		}
		// append link
		r = append(r, []byte("* ["+p.Title+"]("+esc+")\n")...)
		// if we just added a date, add an empty line after the single-element list
		if len(c.Body) > loc[0] && c.Body[loc[0]] != '*' {
			r = append(r, '\n')
		}
		// append the rest
		r = append(r, c.Body[loc[0]:]...)
		c.Body = r
	}
	return c.save()
}
