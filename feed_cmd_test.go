package main

import (
	"bytes"
	"github.com/google/subcommands"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFeedCmd(t *testing.T) {
	cleanup(t, "testdata/complete")
	p := &Page{Name: "testdata/complete/one", Body: []byte("# One\n")}; p.save()
	p = &Page{Name: "testdata/complete/index", Body: []byte(`# Index
* [one](one)
`)}
	p.save()

	b := new(bytes.Buffer)
	s := feedCli(b, []string{"testdata/complete/index.md"})
	assert.Equal(t, subcommands.ExitSuccess, s)
	assert.Contains(t, b.String(), "<fh:complete/>")
}
