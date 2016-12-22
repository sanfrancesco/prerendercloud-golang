2016-12-19 - this is a fork that changes the API URL to prerender.cloud (a chromium alternative to prerender.io, which uses phantomJS)

Prerender Go
===========================

Bots are constantly hitting your site, and a lot of times they're unable to render
javascript.  Prerender.io is awesome, and allows a headless browser to render you
page.

This middleware allows you to intercept requests from crawlers and route them
to an external Prerender Service to retrieve the static HTML for the requested page.

Prerender adheres to google's `_escaped_fragment_` proposal, which we recommend you use. It's easy:
- Just add &lt;meta name="fragment" content="!"> to the &lt;head> of all of your pages
- If you use hash urls (#), change them to the hash-bang (#!)
- That's it! Perfect SEO on javascript pages.

## Features
I tried to replicate the features found in the [Prerender-node](https://github.com/prerender/prerender-node/)
middleware.

## Using it in [negroni](https://github.com/codegangsta/negroni)
``` go
package main

import (
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/sanfrancesco/goprerender"
)

func main() {
	n := negroni.New()
	n.Use(negroni.NewLogger())
	n.Use(prerender.NewOptions().NewPrerender())
	n.Use(negroni.NewStatic(http.Dir(".")))
	n.Run(":8080")
}

```
