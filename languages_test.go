package main

import (
	"os"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestAllLanguage(t *testing.T) {
	os.Unsetenv("ODDMU_LANGUAGES")
	loadLanguages()
	l := language(`
My back hurts at night
My shoulders won't budge today
Winter bones I say`)
	assert.Equal(t, "en", l)
}

func TestSomeLanguages(t *testing.T) {
	os.Setenv("ODDMU_LANGUAGES", "en,de")
	loadLanguages()
	l := language(`
Kühle Morgenluft
Keine Amsel singt heute
Mensch im Dämmerlicht
`)
	assert.Equal(t, "de", l)
}

func TestOneLanguages(t *testing.T) {
	os.Setenv("ODDMU_LANGUAGES", "en")
	loadLanguages()
	l := language(`
Schwer wiegt die Luft hier
Atme ein, ermahn' ich mich
Erinnerungen
`)
	assert.Equal(t, "en", l)
}

func TestWrongLanguages(t *testing.T) {
	os.Setenv("ODDMU_LANGUAGES", "de,fr")
	loadLanguages()
	l := language(`
Something drifts down there
Head submerged oh god a man
Drowning as we stare
`)
	assert.NotEqual(t, "en", l)
}
