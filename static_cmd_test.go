package main

import (
	"github.com/google/subcommands"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStatusCmd(t *testing.T) {
	cleanup(t, "testdata/static")
	s := staticCli("testdata/static")
	assert.Equal(t, subcommands.ExitSuccess, s)
	// pages
	assert.FileExists(t, "testdata/static/index.html")
	assert.FileExists(t, "testdata/static/README.html")
	// regular files
	assert.FileExists(t, "testdata/static/static_cmd.go")
	assert.FileExists(t, "testdata/static/static_cmd_test.go")
}
