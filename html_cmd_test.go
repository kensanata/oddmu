package main

import (
	"bytes"
	"github.com/google/subcommands"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHtmlCmd(t *testing.T) {
	b := new(bytes.Buffer)
	s := htmlCli(b, "", []string{"index.md"})
	assert.Equal(t, subcommands.ExitSuccess, s)
	r := `<h1 id="welcome-to-oddμ">Welcome to Oddμ</h1>

<p>Hello! 🙃</p>

<p>Check out the <a href="README">README</a> and <a href="themes">themes</a>.</p>

<p>Or <a href="test">create a new page</a>.</p>

`
	assert.Equal(t, b.String(), r)
}
