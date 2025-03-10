package main

import (
	"bytes"
	"github.com/google/subcommands"
	"github.com/stretchr/testify/assert"
	"os"
	"regexp"
	"testing"
)

func TestExportCmd(t *testing.T) {
	b := new(bytes.Buffer)
	s := exportCli(b, "feed.html", minimalIndex(t))
	assert.Equal(t, subcommands.ExitSuccess, s)
	assert.Contains(t, b.String(), "<title>Oddμ: A minimal wiki</title>")
	assert.Contains(t, b.String(), "<title>Welcome to Oddμ</title>")
}

func TestExportCmdLanguage(t *testing.T) {
	os.Setenv("ODDMU_LANGUAGES", "de,en")
	loadLanguages()
	p := Page{Body: []byte("This is an English text. All right then!")}
	it := Item{Page: p}
	assert.Equal(t, "en", it.Language())
}

func TestExportCmdJsonFeed(t *testing.T) {
	cleanup(t, "testdata/json")
	os.Mkdir("testdata/json", 0755)
	assert.NoError(t, os.WriteFile("testdata/json/template.json", []byte(`{
  "version": "https://jsonfeed.org/version/1.1",
  "title": "{{.Title}}",
  "home_page_url": "https://alexschroeder.ch",
  "others": [],
  "items": [{{range .Items}}
    {
      "id": "{{.Name}}",
      "url": "https://alexschroeder.ch/view/{{.Name}}",
      "title": "{{.Title}}",
      "language": "{{.Language}}"
      "date_modified": "{{.Date}}",
      "content_html": "{{.Html}}",
      "tags": [{{range .Hashtags}}"{{.}}",{{end}}""],
    },{{end}}
    {}
  ]
}
`), 0644))
	b := new(bytes.Buffer)
	s := exportCli(b, "testdata/json/template.json", minimalIndex(t))
	assert.Equal(t, subcommands.ExitSuccess, s)
	assert.Contains(t, b.String(), `"title": "Oddμ: A minimal wiki"`)
	assert.Regexp(t, regexp.MustCompile("&lt;h1.*&gt;Welcome to Oddμ&lt;/h1&gt;"), b.String()) // skip id
}
