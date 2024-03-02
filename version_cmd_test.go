package main

import (
	"bytes"
	"github.com/google/subcommands"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVersionCmd(t *testing.T) {
	b := new(bytes.Buffer)
	s := versionCli(b, false)
	assert.Equal(t, subcommands.ExitSuccess, s)
	assert.Contains(t, b.String(), "oddmu")
}
