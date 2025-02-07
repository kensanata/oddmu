package main

import (
	"bytes"
	"github.com/google/subcommands"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHashtagsCmd(t *testing.T) {
	cleanup(t, "testdata/hashtag")
	p := &Page{Name: "testdata/hashtag/hash", Body: []byte(`# Hash

I hope for a time
not like today, relentless,
just crocus blooming

#Crocus`)}
	p.save()
	b := new(bytes.Buffer)
	s := hashtagsCli(b)
	assert.Equal(t, subcommands.ExitSuccess, s)
	x := b.String()
	assert.Contains(t, x, "crocus\t")
}
