package main

import (
	"bytes"
	"github.com/sergi/go-diff/diffmatchpatch"
	"html"
	"html/template"
	"net/http"
	"os"
	"strings"
)

func diffHandler(w http.ResponseWriter, r *http.Request, name string) {
	p, err := loadPage(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	p.handleTitle(true)
	p.renderHtml()
	renderTemplate(w, "diff", p)
}

func (p *Page) Diff() template.HTML {
	a := p.Name + ".md~"
	t1, err := os.ReadFile(a)
	if err != nil {
		return template.HTML("Cannot read " + a + ", so the page is new.")
	}
	b := p.Name + ".md"
	t2, err := os.ReadFile(b)
	if err != nil {
		return template.HTML("Cannot read " + b + ", so the page was deleted.")
	}
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(string(t1), string(t2), false)
	return template.HTML(diff2html(dmp.DiffCleanupSemantic(diffs)))
}

func diff2html(diffs []diffmatchpatch.Diff) string {
	var buff bytes.Buffer
	for _, item := range diffs {
		text := strings.ReplaceAll(html.EscapeString(item.Text), "\n", "<br>")
		switch item.Type {
		case diffmatchpatch.DiffInsert:
			_, _ = buff.WriteString("<ins>")
			_, _ = buff.WriteString(text)
			_, _ = buff.WriteString("</ins>")
		case diffmatchpatch.DiffDelete:
			_, _ = buff.WriteString("<del>")
			_, _ = buff.WriteString(text)
			_, _ = buff.WriteString("</del>")
		case diffmatchpatch.DiffEqual:
			_, _ = buff.WriteString("<span>")
			_, _ = buff.WriteString(text)
			_, _ = buff.WriteString("</span>")
		}
	}
	return buff.String()
}
