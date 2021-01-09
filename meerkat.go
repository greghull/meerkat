package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
)

type MenuItem struct {
	Text string
	ID   string
}

type PageData struct {
	Menu []MenuItem
	Body template.HTML
}

var (
	addr  = flag.String("addr", "0.0.0.0:8080", "Address for listening")
	root  = flag.String("root", ".", "Root directory for serving files")
	templ = flag.String("layout", "layout.html", "Layout template for Markdown pages")

	MarkdownSuffix = regexp.MustCompile(`(?i).*\.md$`)

	htmlText = `
<!DOCTYPE html>
<html>
<head>
<link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.5.2/css/bootstrap.min.css">
<script src="https://stackpath.bootstrapcdn.com/bootstrap/4.5.2/js/bootstrap.min.js"></script>
</head>
<body>
<nav class="navbar fixed-top navbar-light bg-light">
{{range .Menu}}
<a class="nav-item nav-link" href="#{{.ID}}">{{.Text}}</a>
{{end}}
</nav>
<div class="container pt-8" style="margin-top:50px;">
{{.Body}}
</div>
</body>
</html>
`
	defaultTemplate = template.Must(template.New("html").Parse(htmlText))
)

func markdownHandler(w http.ResponseWriter, r *http.Request) {
	var markdown = goldmark.New(
		goldmark.WithExtensions(extension.GFM,
			extension.Typographer,
			NewHeaderDivExtension(),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)
	var page PageData

	source, err := ioutil.ReadFile(*root + r.URL.Path)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Render the Markdown to HTML in the buffer
	var buf bytes.Buffer
	n := markdown.Parser().Parse(text.NewReader(source))
	if err := markdown.Renderer().Render(&buf, source, n); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	page.Body = template.HTML(buf.Bytes())

	// Walk nodes to find Headers.  Add Headers to the menu
	n = n.FirstChild()
	for n != nil {
		h, ok := n.(*ast.Heading)
		if ok && h.Level == 1 {
			id, found := h.Attribute([]byte("id"))
			if found {
				page.Menu = append(page.Menu, MenuItem{
					string(h.Text(source)),
					string(id.([]byte)),
				})
			}
		}

		n = n.NextSibling()
	}

	t, err := template.ParseGlob(*templ)
	if err != nil {
		fmt.Println("Using built-in template.")
		t = defaultTemplate
	}
	// Finally write it all to the client
	if err := t.Execute(w, page); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func router(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "/") {
		r.URL.Path += "index.md"
	}

	if MarkdownSuffix.MatchString(r.URL.Path) {
		markdownHandler(w, r)
	} else {
		http.FileServer(http.Dir(*root)).ServeHTTP(w, r)
	}
}

func main() {
	flag.Parse()
	if *root == "" {
		*root = "."
	}

	http.HandleFunc("/", router)

	log.Fatal(http.ListenAndServe(*addr, nil))
}
