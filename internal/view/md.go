package view

// SPDX-License-Identifier: EUPL-1.2

import (
	"bytes"
	"errors"

	mermaid "github.com/abhinav/goldmark-mermaid"
	chromahtml "github.com/alecthomas/chroma/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

// TODO: make it settings
var (
	readmeMD goldmark.Markdown
	mdPath   = "./static/md/"
	tmplPath = "./frontend/template/"
	chroma   = "monokai"

	ErrNoTmplReference = errors.New("page template value is nil")
)

func init() {
	readmeMD = goldmark.New(
		goldmark.WithExtensions(
			meta.Meta,
			highlighting.NewHighlighting(
				highlighting.WithStyle(chroma),
				highlighting.WithFormatOptions(
					chromahtml.WithLineNumbers(true),
				),
			),
			extension.GFM,
			extension.DefinitionList,
			&mermaid.Extender{},
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

// RenderReadme renders a Markdown document to a string
func RenderReadme(s string) (string, error) {
	var buf bytes.Buffer

	if err := readmeMD.Convert([]byte(s), &buf); err != nil {
		return "", err
	}

	return buf.String(), nil
}
