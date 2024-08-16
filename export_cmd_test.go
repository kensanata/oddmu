package main

import (
	"bytes"
	"github.com/google/subcommands"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExportCmd(t *testing.T) {
	b := new(bytes.Buffer)
	s := exportCli(b)
	assert.Equal(t, subcommands.ExitSuccess, s)
	assert.Contains(t, b.String(), "<title>Oddµ: A minimal wiki</title>")
	assert.Contains(t, b.String(), "<title>Welcome to Oddµ</title>")
}
