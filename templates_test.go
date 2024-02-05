package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTemplates(t *testing.T) {
	assert.Contains(t,
		assert.HTTPBody(makeHandler(viewHandler, true), "GET", "/view/index", nil),
		"Skip navigation")
}
