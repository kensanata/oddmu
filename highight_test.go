package main

import (
	"testing"
)

func TestHighlight(t *testing.T) {
	s := "The windows open\nA wave of car noise hits me\nNo birds to be heard."
	q := "window"
	h, c := highlight(q, s)
	if h != "The <b>window</b>s open\nA wave of car noise hits me\nNo birds to be heard." {
		t.Logf("The highlighting is wrong in ｢%s｣", h)
		t.Fail()
	}
	// Score:
	// - q itself
	// - the single token
	// - the beginning of a word
	if c != 3 {
		t.Logf("%s score is %d", q, c)
		t.Fail()
	}
	q = "windows"
	_, c = highlight(q, s)
	// Score:
	// - q itself
	// - the single token
	// - the beginning of a word
	// - the end of a word
	// - the whole word
	if c != 5 {
		t.Logf("%s score is %d", q, c)
		t.Fail()
	}
	q = "car noise"
	_, c = highlight(q, s)
	// Score:
	// - q itself
	// - the car token
	// - the noise token
	// - each with beginning, end and whole token (3 each)
	if c != 9 {
		t.Logf("%s score is %d", q, c)
		t.Fail()
	}
	q = "noise car"
	_, c = highlight(q, s)
	// Score:
	// - the car token
	// - the noise token
	// - each with beginning, end and whole token (3 each)
	if c != 8 {
		t.Logf("%s score is %d", q, c)
		t.Fail()
	}
}
