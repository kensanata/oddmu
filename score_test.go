package main

import (
	"testing"
)

func TestScore(t *testing.T) {

	s := `The windows opens
A wave of car noise hits me
No birds to be heard.`

	q := "window"
	// Score:
	// - q itself
	// - the single token
	// - the beginning of a word
	c := score(q, s)
	if c != 3 {
		t.Logf("%s score is %d", q, c)
		t.Fail()
	}
	q = "windows"
	c = score(q, s)
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
	c = score(q, s)
	// Score:
	// - car noise (+1)
	// - car, with beginning, end, whole word (+4)
	// - noise, with beginning, end, whole word (+4)
	if c != 9 {
		t.Logf("%s score is %d", q, c)
		t.Fail()
	}
	q = "noise car"
	c = score(q, s)
	// Score:
	// - the car token
	// - the noise token
	// - each with beginning, end and whole token (3 each)
	if c != 8 {
		t.Logf("%s score is %d", q, c)
		t.Fail()
	}
}

func TestScoreLong(t *testing.T) {
	s := `We are immersed in a sea of dead people. All the dead that have gone before us, silent now, just staring, gaping. As we move and talk and fret, never once stopping to ask ourselves – or them! – what it was all about. Instead we drown ourselves in noise. Incessantly we babble, surrounded by false friends claiming that all is well. And look at us! Yes, we are well. Patting our backs and expecting a pat – and we do! – we smugly do enjoy.`
	q := "all is well"
	c := score(q, s)
	// Score:
	// - all is well (1)
	// - all, beginning, end, whole word (+4 × 3 = 12)
	// - is, beginning, end, whole word (+4 × 1 = 4), and as a substring (1)
	// - well, beginning, end, whole word (+4 × 2 = 8)
	if c != 26 {
		t.Logf("%s score is %d", q, c)
		t.Fail()
	}
}
