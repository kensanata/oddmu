package main

import (
	"bytes"
	"github.com/google/subcommands"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSearchCmd(t *testing.T) {
	b := new(bytes.Buffer)
	s := searchCli(b, 1, false, []string{"oddµ"})
	assert.Equal(t, subcommands.ExitSuccess, s)
	r := `Search for oddµ, page 1: 2 results
* [Oddµ: A minimal wiki](README)
* [Welcome to Oddµ](index)
`
	assert.Equal(t, r, b.String())
}
