2016-12-19 - this is a fork of https://github.com/tampajohn/goprerender that:

* changes the default API URL to prerender.cloud (a chromium alternative to prerender.io, which uses phantomJS)
* adds support for fasthttp

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

## Set your API token via env var
(get token after signing up at prerender.cloud)

```bash
PRERENDER_TOKEN="mySecretTokenFromPrerenderCloud" go run main.go
```
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

## Using it in [fasthttp](https://github.com/valyala/fasthttp)

```go
package main

import (
	"fmt"

	"github.com/sanfrancesco/goprerender"
	"github.com/valyala/fasthttp"
)

func main() {

	requestHandler := func(ctx *fasthttp.RequestCtx) {
		if string(ctx.UserAgent()) != "prerendercloud" {
			prerender.NewOptions().NewPrerender().PreRenderHandlerFastHttp(ctx)
		} else {
			ctx.SetContentType("text/html")
			fmt.Fprintf(ctx, `
        <div id='root'></div>
        <script type='text/javascript'>
          document.getElementById('root').innerHTML = "hello";
        </script>
      `)
		}
	}

	fasthttp.ListenAndServe(":8080", requestHandler)
}

```
