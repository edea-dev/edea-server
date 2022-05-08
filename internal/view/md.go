package view

// SPDX-License-Identifier: EUPL-1.2

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"text/template"

	chromahtml "github.com/alecthomas/chroma/formatters/html"
	"github.com/gin-gonic/gin"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"gitlab.com/edea-dev/edea-server/internal/model"
	"go.uber.org/zap"
)

// TODO: make it settings
var (
	markdown goldmark.Markdown
	readmeMD goldmark.Markdown
	mdPath   = "./static/md/"
	tmplPath = "./frontend/template/"
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
func RenderMarkdown(c *gin.Context, fn string, data map[string]interface{}) {
	var buf bytes.Buffer

	source, err := ioutil.ReadFile(fmt.Sprintf("%s%s", mdPath, fn))
	if err != nil {
		zap.L().Panic("could not read template", zap.Error(err))
	}

	context := parser.NewContext()
	if err := markdown.Convert([]byte(source), &buf, parser.WithContext(context)); err != nil {
		zap.L().Panic("could not render template", zap.Error(err))
	}
	metaData := meta.Get(context)

	tmplFilename, ok := metaData["Template"]
	if tmplFilename == nil || !ok {
		zap.S().Panicf("check if %s contains the proper headers: %v", fn, ErrNoTmplReference)
	}

	tmpl, err := ioutil.ReadFile(fmt.Sprintf("%s%s", tmplPath, tmplFilename.(string)))
	if err != nil {
		zap.L().Panic("could not read template file referenced in markdown", zap.Error(err))
	}

	navbarTmpl, err := ioutil.ReadFile(fmt.Sprintf("%s%s", tmplPath, "navbar.tmpl"))
	if err != nil {
		zap.L().Panic("could not read navbar template file", zap.Error(err))
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
		zap.L().Panic("failed at first pass", zap.Error(err))
	}

	// second pass, render the markdown components with the page data
	t2 := template.Must(template.New("pass2").Parse(string(sb.String())))
	if err := t2.Execute(c.Writer, data); err != nil {
		zap.L().Panic("failed at second pass", zap.Error(err))
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
func Markdown(page string) gin.HandlerFunc {
	return func(c *gin.Context) {
		RenderMarkdown(c, page, nil)
	}
}

// RenderErrMarkdown renders a page with error information
func RenderErrMarkdown(c *gin.Context, tmpl string, err error) {
	user := c.Keys["user"].(*model.User)
	data := map[string]interface{}{
		"User":  user,
		"Error": err.Error(),
	}

	RenderMarkdown(c, tmpl, data)
}
