package main

import (
	"bytes"
	"github.com/google/subcommands"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMissingCmd(t *testing.T) {
	b := new(bytes.Buffer)
	s := missingCli(b, minimalIndex(t))
	assert.Equal(t, subcommands.ExitSuccess, s)
	r := `Page	Missing
index	test
`
	assert.Equal(t, r, b.String())
}
