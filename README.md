![image](https://cloud.githubusercontent.com/assets/22159102/21554484/9d542f5a-cdc4-11e6-8c4c-7730a9e9e2d1.png)

# prerendercloud-golang

[![CircleCI](https://circleci.com/gh/sanfrancesco/prerendercloud-golang.svg?style=svg)](https://circleci.com/gh/sanfrancesco/prerendercloud-golang)

Includes [negroni](https://github.com/codegangsta/negroni) middleware, and a [fasthttp](https://github.com/valyala/fasthttp) handler for prerendering javascript web pages/apps (single page apps or SPA) with [https://www.prerender.cloud/](https://www.prerender.cloud/)

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
	prerendercloud "github.com/sanfrancesco/prerendercloud-golang"
)

func main() {

	// set the PRERENDER_TOKEN env var when starting this golang binary/executable
	prerenderCloudOptions := prerendercloud.NewOptions()

	// not recommended, but if you must, uncomment this to
	// restrict prerendering to bots and the _escaped_fragment_ query param
	// prerenderCloudOptions.BotsOnly = true
	// with BotsOnly enabled, we don't include googlebot by default (to reduce cloaking penality risk), this is how you could enable it
	// prerendercloud.CrawlerUserAgents = append(prerendercloud.CrawlerUserAgents, "googlebot")

	prerenderCloud := prerenderCloudOptions.NewPrerender()

	n := negroni.New()
	n.Use(negroni.NewLogger())
	n.Use(prerenderCloud)
	n.Use(negroni.NewStatic(http.Dir(".")))
	n.Run(":8080")
}

```

## Using it in [fasthttp](https://github.com/valyala/fasthttp)

```go
package main

import (
	"fmt"

	prerendercloud "github.com/sanfrancesco/prerendercloud-golang"
	"github.com/valyala/fasthttp"
)

func main() {

	// set the PRERENDER_TOKEN env var when starting this golang binary/executable
	prerenderCloudOptions := prerendercloud.NewOptions()

	// not recommended, but if you must, uncomment this to
	// restrict prerendering to bots and the _escaped_fragment_ query param
	// prerenderCloudOptions.BotsOnly = true
	// with BotsOnly enabled, we don't include googlebot by default (to reduce cloaking penality risk), this is how you could enable it
	// prerendercloud.CrawlerUserAgents = append(prerendercloud.CrawlerUserAgents, "googlebot")

	prerenderCloud := prerenderCloudOptions.NewPrerender()

	requestHandler := func(ctx *fasthttp.RequestCtx) {
		if prerenderCloud.ShouldPrerenderFastHttp(ctx) && prerenderCloud.PreRenderHandlerFastHttp(ctx) == nil {
			return
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
