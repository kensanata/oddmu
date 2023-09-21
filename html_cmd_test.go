package main

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
	"github.com/google/subcommands"
)

func TestHtmlCmd(t *testing.T) {
	b := new(bytes.Buffer)
	s := htmlCli(b, false, []string{"index"})
	assert.Equal(t, subcommands.ExitSuccess, s)
	r := `<h1>Welcome to Oddµ</h1>

<p>Hello! 🙃</p>

<p>Check out the <a href="README" rel="nofollow">README</a>.</p>

`
	assert.Equal(t, r, b.String())
}
