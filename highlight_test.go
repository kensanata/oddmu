package main

import (
	"testing"
)

func TestHighlight(t *testing.T) {

	s := `The windows opens
A wave of car noise hits me
No birds to be heard.`

	h := `The <b>window</b>s opens
A wave of car noise hits me
No birds to be heard.`

	q := "window"
	re, _ := re(q)
	r := highlight(q, re, s)
	if r != h {
		t.Logf("The highlighting is wrong in ｢%s｣", r)
		t.Fail()
	}
}

func TestOverlap(t *testing.T) {

	s := `Sit with me my love
Kids shout and so do parents
I hear the fountain`

	h := `Sit with me my love
Kids <b>shout</b> and so do parents
I hear the fountain`

	q := "shout out"
	re, _ := re(q)
	r := highlight(q, re, s)
	if r != h {
		t.Logf("The highlighting is wrong in ｢%s｣", r)
		t.Fail()
	}
}
