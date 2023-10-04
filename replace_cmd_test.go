package main

import (
	"bytes"
	"github.com/google/subcommands"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReplaceCmd(t *testing.T) {
	cleanup(t, "testdata/replace")
	index.load()
	p := &Page{Name: "testdata/replace/pluto", Body: []byte(`# Pluto
Out there is a rock
And more rocks uncountable
You are no planet`)}
	p.save()

	r := `--- testdata/replace/pluto.md~
+++ testdata/replace/pluto.md
@@ -1,4 +1,4 @@
 # Pluto
 Out there is a rock
 And more rocks uncountable
-You are no planet
\ No newline at end of file
+You are planetoid
\ No newline at end of file

1 file would be changed.
This is a dry run. Use -confirm to make it happen.
`

	b := new(bytes.Buffer)
	s := replaceCli(b, false, true, []string{`\bno planet`, `planetoid`})
	assert.Equal(t, subcommands.ExitSuccess, s)
	assert.Equal(t, r, b.String())
}
