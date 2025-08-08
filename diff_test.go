package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestDiff(t *testing.T) {
	cleanup(t, "testdata/diff")
	index.load()
	s := `# Bread

The oven breathes
Fills us with the thought of bread
Oh so fresh, so warm.`
	r := `# Bread

The oven whispers
Fills us with the thought of bread
Oh so fresh, so warm.`
	p := &Page{Name: "testdata/diff/bread", Body: []byte(s)}
	p.save()
	p.Body = []byte(r)
	p.save()
	body := assert.HTTPBody(makeHandler(diffHandler, true, http.MethodGet),
		"GET", "/diff/testdata/diff/bread", nil)
	assert.Contains(t, body, `<del>breathe</del>`)
	assert.Contains(t, body, `<ins>whisper</ins>`)
}

func TestDiffPercentEncoded(t *testing.T) {
	cleanup(t, "testdata/diff")
	index.load()
	s := `# Coup de Gras

Playing D&D
We talk about a killing
Mispronouncing words`
	r := `# Coup de Grace

Playing D&D
We talk about a killing
Mispronouncing words`
	p := &Page{Name: "testdata/diff/coup de grace", Body: []byte(s)}
	p.save()
	p.Body = []byte(r)
	p.save()
	body := assert.HTTPBody(makeHandler(diffHandler, true, http.MethodGet),
		"GET", "/diff/testdata/diff/coup%20de%20grace", nil)
	assert.Contains(t, body, `<del>s</del>`)
	assert.Contains(t, body, `<ins>ce</ins>`)
}

func TestDiffBackup(t *testing.T) {
	cleanup(t, "testdata/backup")
	s := `# Cold Rooms

I shiver at home
the monitor glares and moans
fear or cold, who knows?`
	r := `# Cold Rooms

I shiver at home
the monitor glares and moans
I hate the machine!`
	u := `# Cold Rooms

I shiver at home
the monitor glares and moans
my grey heart grows cold`
	// create s and overwrite it with r
	p := &Page{Name: "testdata/backup/cold", Body: []byte(s)}
	p.save()
	p = &Page{Name: "testdata/backup/cold", Body: []byte(r)}
	p.save()
	body := string(p.Diff())
	// diff from s to r:
	assert.Contains(t, body, `<del>fear or cold, who knows?</del>`)
	assert.Contains(t, body, `<ins>I hate the machine!</ins>`)
	// save u
	p = &Page{Name: "testdata/backup/cold", Body: []byte(u)}
	p.save()
	body = string(p.Diff())
	// diff from s to u since r was not 60 min or older and so the backup is kept
	assert.Contains(t, body, `<del>fear or cold, who knows?</del>`)
	assert.Contains(t, body, `<ins>my grey heart grows cold</ins>`)
	// set timestamp 2h in the past
	ts := time.Now().Add(-2 * time.Hour)
	assert.NoError(t, os.Chtimes("testdata/backup/cold.md~", ts, ts))
	assert.NoError(t, os.Chtimes("testdata/backup/cold.md", ts, ts))
	// save r
	p = &Page{Name: "testdata/backup/cold", Body: []byte(r)}
	p.save()
	body = string(p.Diff())
	// diff from u to r since enough time has passed and the old backup is discarded
	assert.Contains(t, body, `<del>my grey heart grows cold</del>`)
	assert.Contains(t, body, `<ins>I hate the machine!</ins>`)
	// save s
	p = &Page{Name: "testdata/backup/cold", Body: []byte(s)}
	p.save()
	body = string(p.Diff())
	// diff from u to s since this is still "the same" editing window
	assert.Contains(t, body, `<del>my grey heart grows cold</del>`)
	assert.Contains(t, body, `<ins>fear or cold, who knows?</ins>`)
}
