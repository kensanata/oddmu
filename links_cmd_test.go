package main

import (
	"bytes"
	"github.com/google/subcommands"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLinksCmd(t *testing.T) {
	b := new(bytes.Buffer)
	s := linksCli(b, []string{"README.md"})
	assert.Equal(t, subcommands.ExitSuccess, s)
	x := b.String()
	assert.Contains(t, x, "https://alexschroeder.ch/view/oddmu/oddmu.1\n")
}
