package main

import (
	"bytes"
	"github.com/google/subcommands"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVersionCmd(t *testing.T) {
	b := new(bytes.Buffer)
	s := versionCli(b, false, nil)
	assert.Equal(t, subcommands.ExitSuccess, s)
	assert.Contains(t, "vcs.revision", b.String())
}
