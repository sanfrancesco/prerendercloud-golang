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

	prerenderCloud := prerenderCloudOptions.NewPrerender()

	n := negroni.New()
	n.Use(negroni.NewLogger())
	n.Use(prerenderCloud)
	n.Use(negroni.NewStatic(http.Dir(".")))
	n.Run(":8080")
}
