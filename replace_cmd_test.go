package main

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/google/subcommands"
	"os"
	"testing"
)

// wipes testdata
func TestReplaceCmd(t *testing.T) {
	_ = os.RemoveAll("testdata")
	index.load()
	p := &Page{Name: "testdata/pluto", Body: []byte(`# Pluto
Out there is a rock
And more rocks uncountable
You are no planet`)}
	p.save()

	r := `--- testdata/pluto.md~
+++ testdata/pluto.md
@@ -1,4 +1,4 @@
 # Pluto
 Out there is a rock
 And more rocks uncountable
-You are no planet
\ No newline at end of file
+You are planetoid
\ No newline at end of file

1 change was made.
This is a dry run. Use -confirm to make it happen.
`
	
	b := new(bytes.Buffer)
	s := replaceCli(b, false, []string{`\bno planet`, `planetoid`})
	assert.Equal(t, subcommands.ExitSuccess, s)
	assert.Equal(t, r, b.String())

	t.Cleanup(func() {
		_ = os.RemoveAll("testdata")
	})
}
