package main

import (
	"bytes"
	"github.com/google/subcommands"
	"github.com/stretchr/testify/assert"
	"testing"
)

// ## is promoted to level 1 because there is just one instance of level 1
func TestTocCmd(t *testing.T) {
	b := new(bytes.Buffer)
	s := tocCli(b, []string{"README"})
	assert.Equal(t, subcommands.ExitSuccess, s)
	x := b.String()
	assert.Contains(t, x, "\n* [Bugs](#bugs)\n")
}

// ## is promoted to level 1 because there is no instance of level 1
func TestTocNoH1(t *testing.T) {
	p := &Page{
		Body: []byte(`## Venti
Es drückt der Sommer
Weit weg hör' ich ein Flugzeug
Ventilator hilf!`)}
	b := new(bytes.Buffer)
	p.toc().print(b)
	assert.Equal(t, "* [Venti](#venti)\n", b.String())
}

// # is dropped because it's just one level 1 heading
func TestTocDropH1(t *testing.T) {
	p := &Page{Body: []byte("# One\n## Two\n### Three\n")}
	b := new(bytes.Buffer)
	p.toc().print(b)
	assert.Equal(t, "* [Two](#two)\n  * [Three](#three)\n", b.String())
}

// # is kept because there is more than one level 1 heading
func TestTocMultipleH1(t *testing.T) {
	p := &Page{Body: []byte("# One\n# Two\n## Three\n")}
	b := new(bytes.Buffer)
	p.toc().print(b)
	assert.Equal(t, "* [One](#one)\n* [Two](#two)\n  * [Three](#three)\n", b.String())
}
