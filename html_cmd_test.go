package main

import (
	"bytes"
	"github.com/google/subcommands"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHtmlCmd(t *testing.T) {
	b := new(bytes.Buffer)
	s := htmlCli(b, false, []string{"index"})
	assert.Equal(t, subcommands.ExitSuccess, s)
	r := `<h1>Welcome to Oddµ</h1>

<p>Hello! 🙃</p>

<p>Check out the <a href="README">README</a> and <a href="themes">themes</a>.</p>

<p>Or <a href="test">create a new page</a>.</p>

`
	assert.Equal(t, b.String(), r)
}
