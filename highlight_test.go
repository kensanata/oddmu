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
	r := highlight(q, s)
	if r != h {
		t.Logf("The highlighting is wrong in ｢%s｣", r)
		t.Fail()
	}
}
