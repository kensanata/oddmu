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

func TestScoreSubstring(t *testing.T) {
	s := `The loneliness of space means that receiving messages means knowledge that other people are out there. Not satellites pinging forever. Not bots searching and probing. Instead, humans. People who care. Curious and cautious.`
	q := "search probe"
	c := score(q, s)
	// Score:
	// - search, beginning (2)
	// - probe (0)
	if c != 2 {
		t.Logf("%s score is %d", q, c)
		t.Fail()
	}
	q = "ear"
	c = score(q, s)
	// Score:
	// - ear, all (2)
	if c != 2 {
		t.Logf("%s score is %d", q, c)
		t.Fail()
	}
}

func TestScorePageAndMarkup(t *testing.T) {
	s := `The Transjovian Council accepts new members. If you think we'd be a good fit, apply for an account. Contact [Alex Schroeder](https://alexschroeder.ch/wiki/Contact). Mail is best. Encrypted mail is best. [Delta Chat](https://delta.chat/de/) is a messenger app that uses encrypted mail. It's the bestest best.`
	p := &Page{Title: "Test", Name: "Test", Body: []byte(s)}
	q := "wiki"
	p.summarize(q)
	// "wiki" is not visible in the plain text but the score is no affected:
	// - wiki, all, whole, beginning, end (5)
	if p.Score != 5 {
		t.Logf("%s score is %d", q, p.Score)
		t.Fail()
	}
}
