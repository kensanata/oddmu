package main

import (
	"bytes"
	"github.com/google/subcommands"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSearchCmd(t *testing.T) {
	b := new(bytes.Buffer)
	s := searchCli(b, &searchCmd{quiet: true}, []string{"oddμ"})
	assert.Equal(t, subcommands.ExitSuccess, s)
	r := `* [Oddμ: A minimal wiki](README)
* [Themes](themes/index)
* [Welcome to Oddμ](index)
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
	s := searchCli(b, &searchCmd{dir: "testdata/search", quiet: true}, []string{"speak"})
	assert.Equal(t, subcommands.ExitSuccess, s)
	r := `* [Wait](wait)
`
	assert.Equal(t, r, b.String())
}
