package main

import (
	"github.com/google/subcommands"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestStatusCmd(t *testing.T) {
	cleanup(t, "testdata/static")
	s := staticCli("testdata/static", true)
	assert.Equal(t, subcommands.ExitSuccess, s)
	// pages
	assert.FileExists(t, "testdata/static/index.html")
	assert.FileExists(t, "testdata/static/README.html")
	// regular files
	assert.FileExists(t, "testdata/static/static_cmd.go")
	assert.FileExists(t, "testdata/static/static_cmd_test.go")
}

func TestFeedStatusCmd(t *testing.T) {
	cleanup(t, "testdata/static-feed")
	cleanup(t, "testdata/static-feed-out")
	p := &Page{Name: "testdata/static-feed/Haiku",
		Body: []byte(`# Haiku
This is a list.

#Haiku
`)}
	p.save()
	h := &Page{Name: "testdata/static-feed/poem",
		Body: []byte(`# Rain
I cannot hear you
The birds outside are singing
And the cars so loud

#Haiku
`)}
	h.save()
	h.notify()
	wd, err := os.Getwd()
	assert.NoError(t, err)
	assert.NoError(t, os.Chdir("testdata/static-feed"))
	s := staticCli("../static-feed-out/", true)
	assert.Equal(t, subcommands.ExitSuccess, s)
	assert.NoError(t, os.Chdir(wd))
	assert.FileExists(t, "testdata/static-feed-out/poem.html")
	assert.FileExists(t, "testdata/static-feed-out/Haiku.html")
	assert.FileExists(t, "testdata/static-feed-out/Haiku.rss")
}
