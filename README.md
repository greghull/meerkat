# Meerkat

Meerkat is a webserver for Markdown files.


# Usage

Meerkat can be called with no arguments to immediately begin serving markdown files from the 
current directory on port 8080 with a pleasing default layout.

~~~
Usage of meerkat:
  -addr string
      Address for listening (default "0.0.0.0:8080")
  -layout string
      Layout template for Markdown pages (default "layout.html")
  -root string
      Root directory for serving files (default ".")
~~~
