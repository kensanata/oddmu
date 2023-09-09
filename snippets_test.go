package main

import (
	"testing"
)

func TestSnippets(t *testing.T) {
	s := `We are immersed in a sea of dead people. All the dead that have gone before us, silent now, just staring, gaping. As we move and talk and fret, never once stopping to ask ourselves – or them! – what it was all about. Instead we drown ourselves in noise. Incessantly we babble, surrounded by false friends claiming that all is well. And look at us! Yes, we are well. Patting our backs and expecting a pat – and we do! – we smugly do enjoy.`

	h := `We are immersed in a sea of dead people. <b>All</b> the dead that have gone before us, silent now, just … to ask ourselves – or them! – what it was <b>all</b> about. Instead we drown ourselves in no<b>is</b>e. … surrounded by false friends claiming that <b>all</b> <b>is</b> <b>well</b>. And look at us! Yes, we are <b>well</b>. …`

	q := "all is well"
	r := snippets(q, s)
	if r != h {
		t.Logf("The snippets are wrong in ｢%s｣", r)
		t.Fail()
	}
	// Score:
	// - all is well (1)
	// - all, beginning, end, whole word (+4 × 3 = 12)
	// - is, beginning, end, whole word (+4 × 1 = 4), and as a substring (1)
	// - well, beginning, end, whole word (+4 × 2 = 8)
	c := score(q, s)
	if c != 26 {
		t.Logf("%s score is %d", q, c)
		t.Fail()
	}
}
