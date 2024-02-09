package main

import (
	"errors"
	"github.com/pemistahl/lingua-go"
	"os"
	"strings"
)

// getLanguages returns the environment variable ODDMU_LANGUAGES or all languages.
func getLanguages() ([]lingua.Language, error) {
	v := os.Getenv("ODDMU_LANGUAGES")
	if v == "" {
		return lingua.AllLanguages(), nil
	}
	codes := strings.Split(v, ",")
	if len(codes) == 1 {
		return nil, errors.New("detection unnecessary")
	}

	var langs []lingua.Language
	for _, lang := range codes {
		langs = append(langs, lingua.GetLanguageFromIsoCode639_1(lingua.GetIsoCode639_1FromValue(lang)))
	}
	return langs, nil
}

// detector is the LanguageDetector initialized at startup by loadLanguages.
var detector lingua.LanguageDetector

// loadLanguages initializes the detector using the languages returned by getLanguages and returns the number of
// languages loaded. If this is skipped, no language detection happens and the templates cannot use {{.Language}} to use
// this. Usually this is used for correct hyphenation by the browser.
func loadLanguages() int {
	langs, err := getLanguages()
	if err == nil {
		detector = lingua.NewLanguageDetectorBuilder().
			FromLanguages(langs...).
			WithPreloadedLanguageModels().
			WithLowAccuracyMode().
			Build()
	} else {
		detector = nil
	}
	return len(langs)
}

// language returns the language used for a string, as a lower case
// ISO 639-1 string, e.g. "en" or "de".
func language(s string) string {
	if detector == nil {
		return os.Getenv("ODDMU_LANGUAGES")
	}
	if language, ok := detector.DetectLanguageOf(s); ok {
		return strings.ToLower(language.IsoCode639_1().String())
	}
	return ""
}
