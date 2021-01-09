package main

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

// Option is a functional option type for this extension.
type Option func(*HeaderDivExtension)

type HeaderDivExtension struct{}

// New returns a new Hashtag extension.
func NewHeaderDivExtension(opts ...Option) goldmark.Extender {
	e := &HeaderDivExtension{}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// Extend adds a hashtag parser to a Goldmark parser
func (e *HeaderDivExtension) Extend(m goldmark.Markdown) {
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(NewHTMLRenderer(), 500),
		),
	)
}

// HTMLRenderer struct is a renderer.NodeRenderer implementation for the extension.
type HTMLRenderer struct {
	headerCnt int
}

// NewHTMLRenderer builds a new HTMLRenderer with given options and returns it.
func NewHTMLRenderer() renderer.NodeRenderer {
	return &HTMLRenderer{}
}

// RegisterFuncs implements NodeRenderer.RegisterFuncs.
func (r *HTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindHeading, r.renderHeading)
	reg.Register(ast.KindDocument, r.renderDocument)
}

func (r *HTMLRenderer) renderHeading(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Heading)
	if entering {
		if n.Level == 1 {
			if r.headerCnt > 0 {
				_, _ = w.WriteString("</div>")
			}
			_, _ = w.WriteString(`<div class="container pt-3 section" `)
			if n.Attributes() != nil {
				html.RenderAttributes(w, node, html.HeadingAttributeFilter)
			}
			w.WriteString(">\n")
			r.headerCnt += 1
		}
		_, _ = w.WriteString("<h")
		_ = w.WriteByte("0123456"[n.Level])
		if n.Attributes() != nil && n.Level > 1 {
			html.RenderAttributes(w, node, html.HeadingAttributeFilter)
		}
		_ = w.WriteByte('>')
	} else {
		_, _ = w.WriteString("</h")
		_ = w.WriteByte("0123456"[n.Level])
		_, _ = w.WriteString(">\n")
	}
	return ast.WalkContinue, nil
}

func (r *HTMLRenderer) renderDocument(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	// If we are finishing the document, and there is an open <div> from an H1 Header, then close it now
	if !entering && r.headerCnt > 1 {
		_, _ = w.WriteString("</div>")
	}
	return ast.WalkContinue, nil
}
