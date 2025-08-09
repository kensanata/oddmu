package main

import (
	"log"
	"path"
	"regexp"
	"strings"
	"time"
)

// notify adds a link to the "changes" page, the "index" page, as well as to all the existing hashtag pages. The link to
// the "index" page is only added if the page being edited is a blog page for the current year. The link to existing
// hashtag pages is only added for blog pages. If the "changes" page does not exist, it is created. If the hashtag page
// does not exist, it is not. Hashtag pages are considered optional. If the page that's being edited is in a
// subdirectory, then the "changes", "index" and hashtag pages of that particular subdirectory are affected. Every
// subdirectory is treated like a potentially independent wiki. Errors are logged before being returned because the
// error messages are confusing from the point of view of the saveHandler.
func (p *Page) notify() error {
	p.handleTitle(false)
	if p.Title == "" {
		p.Title = p.Name
	}
	esc := nameEscape(p.Base())
	link := "* [" + p.Title + "](" + esc + ")\n"
	re := regexp.MustCompile(`(?m)^\* \[[^\]]+\]\(` + esc + `\)\n`)
	dir := p.Dir()
	err := addLinkWithDate(path.Join(dir, "changes"), link, re)
	if err != nil {
		log.Printf("Updating changes in %s failed: %s", dir, err)
		return err
	}
	if p.IsBlog() {
		// Add to the index only if the blog post is for the current year
		if strings.HasPrefix(p.Base(), time.Now().Format("2006")) {
			err := addLink(path.Join(dir, "index"), true, link, re)
			if err != nil {
				log.Printf("Updating index in %s failed: %s", dir, err)
				return err
			}
		}
		p.renderHtml() // to set hashtags
		for _, hashtag := range p.Hashtags {
			err := addLink(path.Join(dir, hashtag), false, link, re)
			if err != nil {
				log.Printf("Updating hashtag %s in %s failed: %s", hashtag, dir, err)
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
	p, err := loadPage(name)
	if err != nil {
		// create a new page
		p = &Page{Name: name, Body: []byte("# Changes\n\n## " + date + "\n" + link)}
	} else {
		org = string(p.Body)
		// remove the old match, if one exists
		loc := re.FindIndex(p.Body)
		if loc != nil {
			r := p.Body[:loc[0]]
			if loc[1] < len(p.Body) {
				r = append(r, p.Body[loc[1]:]...)
			}
			p.Body = r
			if loc[0] >= 14 && len(p.Body) >= loc[0]+15 {
				// remove the preceding date if there are now two dates following each other
				re := regexp.MustCompile(`(?m)^## (\d\d\d\d-\d\d-\d\d)\n\n## (\d\d\d\d-\d\d-\d\d)\n`)
				if re.Match(p.Body[loc[0]-14 : loc[0]+15]) {
					p.Body = append(p.Body[0:loc[0]-14], p.Body[loc[0]+1:]...)
				}
			} else if len(p.Body) == loc[0] {
				// remove a trailing date
				re := regexp.MustCompile(`## (\d\d\d\d-\d\d-\d\d)\n`)
				if re.Match(p.Body[loc[0]-14 : loc[0]]) {
					p.Body = p.Body[0 : loc[0]-14]
				}
			}
		}
		// locate the beginning of the list to insert the line
		re := regexp.MustCompile(`(?m)^\* \[[^\]]+\]\([^\)]+\)\n`)
		loc = re.FindIndex(p.Body)
		if loc == nil {
			// if no list was found, use the end of the page
			loc = []int{len(p.Body)}
		}
		// start with new page content
		r := []byte("")
		// check if there is a date right before the insertion point
		addDate := true
		if loc[0] >= 14 {
			re := regexp.MustCompile(`(?m)^## (\d\d\d\d-\d\d-\d\d)\n`)
			m := re.Find(p.Body[loc[0]-14 : loc[0]])
			if m == nil {
				// not a date: insert date, don't move insertion point
			} else if string(p.Body[loc[0]-11:loc[0]-1]) == date {
				// if the date is our date, don't add it, don't move insertion point
				addDate = false
			} else {
				// if the date is not out date, move the insertion point
				loc[0] -= 14
			}
		}
		// append up to the insertion point
		r = append(r, p.Body[:loc[0]]...)
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
		if len(p.Body) > loc[0] && p.Body[loc[0]] != '*' {
			r = append(r, '\n')
		}
		// append the rest
		r = append(r, p.Body[loc[0]:]...)
		p.Body = r
	}
	// only save if something changed
	if string(p.Body) != org {
		return p.save()
	}
	return nil
}

// addLink adds a link to a named page, if the page exists and doesn't contain the link. If the link exists but with a
// different title, the title is fixed.
func addLink(name string, mandatory bool, link string, re *regexp.Regexp) error {
	p, err := loadPage(name)
	if err != nil {
		if mandatory {
			p = &Page{Name: name, Body: []byte(link)}
			return p.save()
		} else {
			// Skip non-existing files: no error
			return nil
		}
	}
	org := string(p.Body)
	addLinkToPage(p, link, re)
	// only save if something changed
	if string(p.Body) != org {
		return p.save()
	}
	return nil
}

func addLinkToPage(p *Page, link string, re *regexp.Regexp) {
	// if a link exists, that's the place to insert the new link (in which case loc[0] and loc[1] differ)
	loc := re.FindIndex(p.Body)
	// if no link exists, find a good place to insert it
	if loc == nil {
		// locate the list items
		re = regexp.MustCompile(`(?m)^\* \[[^\]]+\]\([^\)]+\)\n?`)
		items := re.FindAllIndex(p.Body, -1)
		first := false
		pos := -1
		// skip newer items
		for i, it := range items {
			// break if the current line is older (earlier in sort order)
			stop := string(p.Body[it[0]:it[1]]) < link
			// before the first match is always a good insert point
			if i == 0 {
				pos = it[0]
				first = true
			}
			// if we're not stopping, then after the current item is a good insert point
			if !stop {
				pos = it[1]
				first = false
			} else {
				break
			}
		}
		// otherwise it's at the end of the list, after the last item
		if pos == -1 && len(items) > 0 {
			pos = items[len(items)-1][1]
			first = false
		}
		// if no list was found, use the end of the page
		if pos == -1 {
			pos = len(p.Body)
			first = true
		}
		if first {
			p.Body, pos = ensureTwoNewlines(p.Body, pos)
		}
		// mimic a zero-width match at the insert point
		loc = []int{pos, pos}
	}
	// start with new page content
	r := []byte("")
	// append up to the insertion point
	r = append(r, p.Body[:loc[0]]...)
	// append link
	r = append(r, []byte(link)...)
	// append the rest
	r = append(r, p.Body[loc[1]:]...)
	p.Body = r
}

// ensureTwoNewlines makes sure that the two bytes before pos in buf are newlines. If the are not, newlines are inserted
// and pos is increased. The new buf and pos is returned.
func ensureTwoNewlines(buf []byte, pos int) ([]byte, int) {
	var insert []byte
	if pos >= 1 && buf[pos-1] != '\n' {
		// add two newlines if buf doesn't end with a newline
		insert = []byte("\n\n")
	} else if pos >= 2 && buf[pos-2] != '\n' {
		// add one newline if Body ends with just one newline
		insert = []byte("\n")
	}
	if insert != nil {
		r := []byte("")
		r = append(r, buf[:pos]...)
		r = append(r, insert...)
		r = append(r, buf[pos:]...)
		buf = r
		pos += len(insert)

	}
	return buf, pos
}
