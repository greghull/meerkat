# Meerkat

Meerkat translates Markdown files to HTML files, using a Go Template.
At each H1, H2, ..., H5 element, Structure is added to the HTML output to accomodate anchor 
links and toggling the visibility of sections of the Document.

A single page website can be produced from a single Markdown file that also has nice fall-back
behavior when Javascript is not present.



# Usage

Meerkat can be called with no arguments to watch for changes to Markdown files in the source
directory.

~~~
Usage of meerkat:
  -layout string
    	Layout template for Markdown pages (default "layout.html")
  -minify
    	Minify HTML output
  -output string
    	Directory to write HTML files (default ".")
  -source string
    	Directory to watch for Markdown files (default ".")
~~~
