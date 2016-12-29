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

	prerenderCloud := prerenderCloudOptions.NewPrerender()

	requestHandler := func(ctx *fasthttp.RequestCtx) {
		if prerenderCloud.ShouldPrerenderFastHttp(ctx) {
			prerenderCloud.PreRenderHandlerFastHttp(ctx)
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
