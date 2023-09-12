package main

import (
	"os"
	"errors"
	"github.com/pemistahl/lingua-go"
	"strings"
)

// getLangauges returns the environment variable ODDMU_LANGUAGES or
// all languages.
func getLanguages() ([]lingua.Language, error) {
	v := os.Getenv("ODDMU_LANGUAGES")
	if v == "" {
		return lingua.AllLanguages(), nil
	}
	codes := strings.Split(v, ",")
	if len(codes) == 1 {
		return nil, errors.New("Detection unnecessary")
	}
	
	var langs []lingua.Language
	for _, lang := range codes {
		langs = append(langs, lingua.GetLanguageFromIsoCode639_1(lingua.GetIsoCode639_1FromValue(lang)))
	}
	return langs, nil
}

// detector is the LanguageDetector initialized at startup by loadLanguages.
var detector lingua.LanguageDetector

// loadLanguages initializes the detector using the languages returned
// by getLanguages.
func loadLanguages() {
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
