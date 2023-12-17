package main

import (
	"bytes"
	"github.com/google/subcommands"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestListCmd(t *testing.T) {
	b := new(bytes.Buffer)
	s := listCli(b, "", nil)
	assert.Equal(t, subcommands.ExitSuccess, s)
	x := b.String()
	assert.Contains(t, x, "README\tOddµ: A minimal wiki\n")
	assert.Contains(t, x, "index\tWelcome to Oddµ\n")
}

func TestListSubdirCmd(t *testing.T) {
	cleanup(t, "testdata/list")
	p := &Page{Name: "testdata/list/red", Body: []byte(`# Red
Shifting darkness waits
I open my eyes in fear
And see the red dot`)}
	p.save()
	b := new(bytes.Buffer)
	s := listCli(b, "testdata/list", nil)
	assert.Equal(t, subcommands.ExitSuccess, s)
	x := b.String()
	assert.Contains(t, x, "red\tRed\n")
}
