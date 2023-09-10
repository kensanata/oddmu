package main

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
	"github.com/google/subcommands"
)

func TestSearchCmd(t *testing.T) {
	b := new(bytes.Buffer)
	s := searchCli(b, 1, false, []string{"oddµ"})
	assert.Equal(t, subcommands.ExitSuccess, s)
	r := `Search for oddµ, page 1: 2 results
* [Oddµ: A minimal wiki](README) (5)
* [Welcome to Oddµ](index) (5)
`
	assert.Equal(t, r, b.String())
}
