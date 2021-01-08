package main

import (
	"bytes"
	"flag"
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
	text string
	id   string
}

type PageData struct {
	Menu []MenuItem
	Body template.HTML
}

var (
	addr = flag.String("addr", "0.0.0.0:8080", "Address for listening")
	root = flag.String("root", ".", "Root directory for serving files")

	MarkdownSuffix = regexp.MustCompile(`(?i).*\.md$`)
	Markdown       = goldmark.New(
		goldmark.WithExtensions(extension.GFM, extension.Typographer),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)

	htmlText = `
<html>
<header>
<link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.5.2/css/bootstrap.min.css">
<script src="https://stackpath.bootstrapcdn.com/bootstrap/4.5.2/js/bootstrap.min.js"></script>
</header>
<body>
<nav class="navbar fixed-top navbar-light bg-light">
  <a class="navbar-brand" href="#">Fixed top</a>
</nav>
<button class="navbar-toggler" type="button" data-toggle="collapse" data-target="#navbarNavAltMarkup" aria-controls="navbarNavAltMarkup" aria-expanded="false" aria-label="Toggle navigation">
    <span class="navbar-toggler-icon"></span>
  </button>
  <div class="collapse navbar-collapse" id="navbarNavAltMarkup">
    <div class="navbar-nav">
      <a class="nav-item nav-link active" href="#">Home <span class="sr-only">(current)</span></a>
      <a class="nav-item nav-link" href="#">Features</a>
      <a class="nav-item nav-link" href="#">Pricing</a>
      <a class="nav-item nav-link disabled" href="#">Disabled</a>
    </div>
  </div>
<div class="container pt-8">
{{.Body}}
</div>
</body>
</html>
`
	Template = template.Must(template.New("html").Parse(htmlText))
)

func markdownHandler(w http.ResponseWriter, r *http.Request) {
	var page PageData

	source, err := ioutil.ReadFile(*root + r.URL.Path)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Render the Markdown to HTML in the buffer
	var buf bytes.Buffer
	n := Markdown.Parser().Parse(text.NewReader(source))
	if err := Markdown.Renderer().Render(&buf, source, n); err != nil {
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

	// Finally write it all to the client
	if err := Template.Execute(w, page); err != nil {
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
