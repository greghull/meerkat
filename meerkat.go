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
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

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
<div class="container">
{{.}}
</div>
</body>
</html>
`
	Template = template.Must(template.New("html").Parse(htmlText))
)

func markdownHandler(w http.ResponseWriter, r *http.Request) {
	source, err := ioutil.ReadFile(*root + r.URL.Path)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	var buf bytes.Buffer
	if err := Markdown.Convert(source, &buf); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := Template.Execute(w, template.HTML(buf.Bytes())); err != nil {
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
