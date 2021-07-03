package view

// SPDX-License-Identifier: EUPL-1.2

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"text/template"

	chromahtml "github.com/alecthomas/chroma/formatters/html"
	"github.com/rs/zerolog/log"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"gitlab.com/edea-dev/edead/model"
	"gitlab.com/edea-dev/edead/util"
)

// TODO: make it settings
var (
	markdown goldmark.Markdown
	readmeMD goldmark.Markdown
	mdPath   = "./static/md/"
	tmplPath = "./static/tmpl/"
	chroma   = "monokai"

	ErrNoTmplReference = errors.New("page template value is nil")
)

func init() {
	markdown = goldmark.New(
		goldmark.WithExtensions(
			meta.Meta,
			highlighting.NewHighlighting(
				highlighting.WithStyle(chroma),
				highlighting.WithFormatOptions(
					chromahtml.WithLineNumbers(true),
				),
			),
			extension.DefinitionList,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
			parser.WithAttribute(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
			html.WithUnsafe(),
		),
	)

	readmeMD = goldmark.New(
		goldmark.WithExtensions(
			meta.Meta,
			highlighting.NewHighlighting(
				highlighting.WithStyle(chroma),
				highlighting.WithFormatOptions(
					chromahtml.WithLineNumbers(true),
				),
			),
			extension.DefinitionList,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
			parser.WithAttribute(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
		),
	)
}

// RenderMarkdown renders a markdown/go template combination
func RenderMarkdown(fn string, data map[string]interface{}, w io.Writer) {
	var buf bytes.Buffer

	source, err := ioutil.ReadFile(fmt.Sprintf("%s%s", mdPath, fn))
	if err != nil {
		log.Panic().Err(err).Msgf("could not render template: %v", err)
	}

	context := parser.NewContext()
	if err := markdown.Convert([]byte(source), &buf, parser.WithContext(context)); err != nil {
		log.Panic().Err(err).Msgf("could not render template: %v", err)
	}
	metaData := meta.Get(context)

	tmplFilename, ok := metaData["Template"]
	if tmplFilename == nil || !ok {
		log.Panic().Err(ErrNoTmplReference).Msgf("check if %s contains the proper headers: %v", fn, ErrNoTmplReference)
	}

	tmpl, err := ioutil.ReadFile(fmt.Sprintf("%s%s", tmplPath, tmplFilename.(string)))
	if err != nil {
		log.Panic().Err(err).Msgf("could not read template file referenced in markdown: %v", err)
	}

	navbarTmpl, err := ioutil.ReadFile(fmt.Sprintf("%s%s", tmplPath, "navbar.tmpl"))
	if err != nil {
		log.Panic().Err(err).Msgf("could not read navbar template file: %v", err)
	}

	metaData["Body"] = buf.String()
	metaData["Dev"] = true

	for k, v := range data {
		metaData[k] = v
	}

	sb := new(strings.Builder)

	// parse navbar
	t := template.Must(template.New("navbar").Parse(string(navbarTmpl)))

	// first pass, render markdown into outer template
	t = template.Must(t.New("pass1").Parse(string(tmpl)))
	if err := t.Execute(sb, metaData); err != nil {
		log.Panic().Err(err).Msg("failed at first pass")
	}

	// second pass, render the markdown components with the page data
	t2 := template.Must(template.New("pass2").Parse(string(sb.String())))
	if err := t2.Execute(w, data); err != nil {
		log.Panic().Err(err).Msg("failed at second pass")
	}
}

// RenderReadme renders a Markdown document to a string
func RenderReadme(s string) (string, error) {
	var buf bytes.Buffer

	if err := readmeMD.Convert([]byte(s), &buf); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// Markdown renders a markdown page without any template values
func Markdown(page string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		RenderMarkdown(page, nil, w)
	}
}

// RenderErrMarkdown renders a page with error information
func RenderErrMarkdown(ctx context.Context, w http.ResponseWriter, tmpl string, err error) {
	user := ctx.Value(util.UserContextKey).(*model.User)
	data := map[string]interface{}{
		"User":  user,
		"Error": err.Error(),
	}

	RenderMarkdown(tmpl, data, w)
}
