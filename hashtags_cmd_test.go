package main

import (
	"bytes"
	"github.com/google/subcommands"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHashtagsCmd(t *testing.T) {
	b := new(bytes.Buffer)
	s := hashtagsCli(b)
	assert.Equal(t, subcommands.ExitSuccess, s)
	x := b.String()
	assert.Contains(t, x, "#like_this\t")
}
