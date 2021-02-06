package main

import (
    "fmt"
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
	hCnt      [6]int
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
	lvl := n.Level

	if entering {
		// close old divs when entering into a new heading.  Each heading needs to close all divs <= than it
		// i.e. An H1 closes divs for H1, H2, ..., H5
		// an H3 closes dives for H3, H4, H5
		for i := lvl; i <= 5; i++ {
			if r.hCnt[i] > 0 {
				_, _ = w.WriteString("</div>")
				r.hCnt[i]--
			}
		}

		w.WriteString("\n")

		// Open parent div for this header
        if lvl == 1 {
            fmt.Fprintf(w, `<div class="container section%v">`, lvl)
        } else {
            fmt.Fprintf(w, `<div class="section%v">`, lvl)
        }
		r.hCnt[lvl]++

		// a span for anchor link
		_, _ = w.WriteString(`<span class="anchor" `)
		if n.Attributes() != nil {
			html.RenderAttributes(w, node, html.HeadingAttributeFilter)
		}
		_, _ = w.WriteString("></span>\n")

		// Write the header tag
		_, _ = w.WriteString("<h")
		_ = w.WriteByte("0123456"[n.Level])
		if n.Attributes() != nil && n.Level > 1 {
			html.RenderAttributes(w, node, html.HeadingAttributeFilter)
		}
		_ = w.WriteByte('>')
	} else {
		// Close the Header tag
		_, _ = w.WriteString("</h")
		_ = w.WriteByte("0123456"[n.Level])
		_, _ = w.WriteString(">\n")
	}
	return ast.WalkContinue, nil
}

func (r *HTMLRenderer) renderDocument(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	// If we are finishing the document, and there is an open <div> from an H1 Header, then close it now
	if !entering {
		// when exiting a document, close all the open divs
		// that were created by renderHeading
		for i := 1; i < 6; i++ {
			if r.hCnt[i] > 1 {
				_, _ = w.WriteString("</div>")
				r.hCnt[i]--
			}
		}
	}
	return ast.WalkContinue, nil
}
