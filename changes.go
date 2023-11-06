package main

import (
	"path"
	"regexp"
	"strings"
	"time"
)

// notify adds a link to the "changes" page, as well as to all the existing hashtag pages. If the "changes" page does
// not exist, it is created. If the hashtag page does not exist, it is not. Hashtag pages are considered optional.
func (p *Page) notify() error {
	p.handleTitle(false)
	if p.Title == "" {
		p.Title = p.Name
	}
	esc := nameEscape(p.Name)
	link := "* [" + p.Title + "](" + esc + ")\n"
	re := regexp.MustCompile(`(?m)^\* \[[^\]]+\]\(` + esc + `\)\n`)
	dir := path.Dir(p.Name)
	// Recent changes for all pages
	err := addLinkWithDate("changes", link, re)
	if err != nil {
		return err
	}
	// For blog pages only…
	if p.isBlog() {
		// Add to the index only if the blog post is for the current year
		if strings.HasPrefix(path.Base(p.Name), time.Now().Format("2006")) {
			err := addLink("index", link, re)
			if err != nil {
				return err
			}
		}
		// Update hashtag pages
		p.renderHtml() // to set hashtags
		for _, hashtag := range p.Hashtags {
			err := addLink(path.Join(dir, hashtag), link, re)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// addLinkWithDate adds the link to a page, with date header for today. If a match already exists, it is removed. If
// this leaves a date header without any links, it is removed as well. If a list is found, the link is added at the top
// of the list. Lists must use the asterisk, not the minus character.
func addLinkWithDate(name, link string, re *regexp.Regexp) error {
	date := time.Now().Format(time.DateOnly)
	org := ""
	c, err := loadPage(name)
	if err != nil {
		// create a new page
		c = &Page{Name: "changes", Body: []byte("# Changes\n\n## " + date + "\n" + link)}
	} else {
		org = string(c.Body)
		// remove the old match, if one exists
		loc := re.FindIndex(c.Body)
		if loc != nil {
			r := c.Body[:loc[0]]
			if loc[1] < len(c.Body) {
				r = append(r, c.Body[loc[1]:]...)
			}
			c.Body = r
			if loc[0] >= 14 && len(c.Body) >= loc[0]+15 {
				// remove the preceding date if there are now two dates following each other
				re := regexp.MustCompile(`(?m)^## (\d\d\d\d-\d\d-\d\d)\n\n## (\d\d\d\d-\d\d-\d\d)\n`)
				if re.Match(c.Body[loc[0]-14 : loc[0]+15]) {
					c.Body = append(c.Body[0 : loc[0]-14], c.Body[loc[0]+1 : ]...)
				}
			} else if len(c.Body) == loc[0] {
				// remove a trailing date
				re := regexp.MustCompile(`## (\d\d\d\d-\d\d-\d\d)\n`)
				if re.Match(c.Body[loc[0]-14 : loc[0]]) {
					c.Body = c.Body[0 : loc[0]-14]
				}
			}
		}
		// locate the beginning of the list to insert the line
		re := regexp.MustCompile(`(?m)^\* \[[^\]]+\]\([^\)]+\)\n`)
		loc = re.FindIndex(c.Body)
		if loc == nil {
			// if no list was found, use the end of the page
			loc = []int{len(c.Body)}
		}
		// start with new page content
		r := []byte("")
		// check if there is a date right before the insertion point
		addDate := true
		if loc[0] >= 14 {
			re := regexp.MustCompile(`(?m)^## (\d\d\d\d-\d\d-\d\d)\n`)
			m := re.Find(c.Body[loc[0]-14 : loc[0]])
			if m == nil {
				// not a date: insert date, don't move insertion point
			} else if string(c.Body[loc[0]-11 : loc[0]-1]) == date {
				// if the date is our date, don't add it, don't move insertion point
				addDate = false
			} else {
				// if the date is not out date, move the insertion point
				loc[0] -= 14
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
			r = append(r, []byte("## ")...)
			r = append(r, []byte(date)...)
			r = append(r, '\n')
		}
		// append link
		r = append(r, []byte(link)...)
		// if we just added a date, add an empty line after the single-element list
		if len(c.Body) > loc[0] && c.Body[loc[0]] != '*' {
			r = append(r, '\n')
		}
		// append the rest
		r = append(r, c.Body[loc[0]:]...)
		c.Body = r
	}
	// only save if something changed
	if string(c.Body) != org {
		return c.save()
	}
	return nil
}

// addLink adds a link to a named page, if the page exists and doesn't contain the link. If the link exists but with a
// different title, the title is fixed.
func addLink(name, link string, re *regexp.Regexp) error {
	c, err := loadPage(name)
	if err != nil {
		// Skip non-existing files: no error
		return nil
	}
	org := string(c.Body)
	// if a link exists, that's the place to insert the new link (in which case loc[0] and loc[1] differ)
	loc := re.FindIndex(c.Body)
	// if no link exists, find a good place to insert it
	if loc == nil {
		// locate the beginning of the list to insert the line
		re = regexp.MustCompile(`(?m)^\* \[[^\]]+\]\([^\)]+\)\n`)
		loc = re.FindIndex(c.Body)
		if loc == nil {
			// if no list was found, use the end of the page
			m := len(c.Body)
			loc = []int{m, m}
		} else {
			// if a list item was found, use just the beginning as insertion point
			loc[1] = loc[0]
		}
	}
	// start with new page content
	r := []byte("")
	// append up to the insertion point
	r = append(r, c.Body[:loc[0]]...)
	// append link
	r = append(r, []byte(link)...)
	// append the rest
	r = append(r, c.Body[loc[1]:]...)
	c.Body = r
	// only save if something changed
	if string(c.Body) != org {
		return c.save()
	}
	return nil
}
