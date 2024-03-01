package main

import (
	"bytes"
	"github.com/google/subcommands"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSearchCmd(t *testing.T) {
	b := new(bytes.Buffer)
	s := searchCli(b, "", 1, false, false, true, []string{"oddµ"})
	assert.Equal(t, subcommands.ExitSuccess, s)
	r := `* [Oddµ: A minimal wiki](README)
* [Themes](themes/index)
* [Welcome to Oddµ](index)
`
	assert.Equal(t, r, b.String())
}

func TestSearchSubdirCmd(t *testing.T) {
	cleanup(t, "testdata/search")
	p := &Page{Name: "testdata/search/wait", Body: []byte(`# Wait
We should make it so
that before we type and speak
we hear that moment`)}
	p.save()
	b := new(bytes.Buffer)
	s := searchCli(b, "testdata/search", 1, false, false, true, []string{"speak"})
	assert.Equal(t, subcommands.ExitSuccess, s)
	r := `* [Wait](wait)
`
	assert.Equal(t, r, b.String())
}
