// Package prerender provides a Prerender.cloud handler implementation and a
// Negroni middleware.
package prerendercloud

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"

	e "github.com/jqatampa/gadget-arm/errors"
	"github.com/valyala/fasthttp"
	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
)

// Options provides you with the ability to specify a custom Prerender.cloud URL
// as well as a Prerender.cloud Token to include as an X-Prerender-Token header
// to the upstream server.
type Options struct {
	PrerenderURL   *url.URL
	Token          string
	UsingAppEngine bool
	BotsOnly       bool
}

// NewOptions generates a default Options struct pointing to the Prerender.cloud
// service, obtaining a Token from the environment variable PRERENDER_TOKEN.
func NewOptions() *Options {
	var url *url.URL

	if os.Getenv("PRERENDER_SERVICE_URL") != "" {
		url, _ = url.Parse(os.Getenv("PRERENDER_SERVICE_URL"))
	} else {
		url, _ = url.Parse("https://service.headless-render-api.com/")
	}

	return &Options{
		PrerenderURL:   url,
		Token:          os.Getenv("PRERENDER_TOKEN"),
		UsingAppEngine: false,
		BotsOnly:       false,
	}
}

// Prerender exposes methods to validate and serve content from a Prerender.cloud
// upstream server.
type Prerender struct {
	Options *Options
}

// NewPrerender generates a new Prerender instance.
func (o *Options) NewPrerender() *Prerender {
	return &Prerender{Options: o}
}

// ServeHTTP allows Prerender to act as a Negroni middleware.
func (p *Prerender) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if p.ShouldPrerender(r) {
		p.PreRenderHandler(rw, r, next)
	} else if next != nil {
		next(rw, r)
	}
}

func (p *Prerender) ShouldPrerenderFastHttp(ctx *fasthttp.RequestCtx) bool {
	userAgent := strings.ToLower(string(ctx.UserAgent()))
	method := strings.ToLower(string(ctx.Method()))

	if userAgent == "" || userAgent == "prerendercloud" {
		return false
	}

	if string(ctx.Request.Header.Peek("X-Prerendered")) != "" {
		return false
	}

	if method != "get" && method != "head" {
		return false
	}

	if !prerenderableExtension(string(ctx.Path())) {
		return false
	}

	if p.Options.BotsOnly {
		isRequestingPrerenderedPage := false
		bufferAgent := string(ctx.Request.Header.Peek("X-Bufferbot"))

		// Buffer Agent or requesting an escaped fragment, request prerender
		if bufferAgent != "" || strings.Contains(string(ctx.QueryArgs().QueryString()), "_escaped_fragment_") {
			isRequestingPrerenderedPage = true
		}

		// Crawler, request prerender
		for _, crawlerAgent := range CrawlerUserAgents {
			if strings.Contains(crawlerAgent, strings.ToLower(userAgent)) {
				isRequestingPrerenderedPage = true
				break
			}
		}

		return isRequestingPrerenderedPage
	} else {
		return true
	}
}

// ShouldPrerender analyzes the request to determine whether it should be routed
// to a Prerender.cloud upstream server.
func (p *Prerender) ShouldPrerender(or *http.Request) bool {
	userAgent := strings.ToLower(or.Header.Get("User-Agent"))
	method := strings.ToLower(or.Method)

	// No user agent, don't prerender
	if userAgent == "" || userAgent == "prerendercloud" {
		return false
	}

	if or.Header.Get("X-Prerendered") != "" {
		return false
	}

	if method != "get" && method != "head" {
		return false
	}

	if !prerenderableExtension(or.URL.EscapedPath()) {
		return false
	}

	if p.Options.BotsOnly {
		bufferAgent := or.Header.Get("X-Bufferbot")
		isRequestingPrerenderedPage := false

		// var isEscapedFragment bool
		_, isEscapedFragment := or.URL.Query()["_escaped_fragment_"]

		// Buffer Agent or requesting an escaped fragment, request prerender
		if bufferAgent != "" || isEscapedFragment {
			isRequestingPrerenderedPage = true
		}

		// Cralwer, request prerender
		for _, crawlerAgent := range CrawlerUserAgents {
			if strings.Contains(crawlerAgent, strings.ToLower(userAgent)) {
				isRequestingPrerenderedPage = true
				break
			}
		}

		return isRequestingPrerenderedPage
	} else {
		return true
	}

}

func prerenderableExtension(fullpath string) bool {
	basename := path.Base(fullpath)

	// path.Base returns "." for empty strings
	if basename == "." {
		return true
	}

	// doesn't detect index.whatever.html (multiple dots)
	hasHtmlOrNoExtension, _ := regexp.MatchString("^(([^.]|\\.html?)+)$", basename)

	if hasHtmlOrNoExtension {
		return true
	}

	// hack to handle basenames with multiple dots: index.whatever.html
	endsInHtml, _ := regexp.MatchString(".html?$", basename)

	if endsInHtml {
		return true
	}

	return false
}

func (p *Prerender) buildURLforFastHttp(ctx *fasthttp.RequestCtx) string {
	return buildApiUrl(
		p.Options.PrerenderURL.String(),
		string(ctx.URI().Scheme()),
		string(ctx.Host()),
		string(ctx.Path()),
		string(ctx.URI().QueryString()),
	)
}

func (p *Prerender) buildURLforHttp(or *http.Request) string {
	return buildApiUrl(p.Options.PrerenderURL.String(),
		or.URL.Scheme,
		or.Host,
		or.URL.Path,
		or.URL.RawQuery,
	)
}

func buildApiUrl(prerenderServiceUrl, protocol, host, path, rawQuery string) string {
	if !strings.HasSuffix(prerenderServiceUrl, "/") {
		prerenderServiceUrl += "/"
	}

	if len(protocol) == 0 {
		protocol = "http"
	}

	apiUrl := prerenderServiceUrl
	apiUrl += protocol
	apiUrl += "://"
	apiUrl += host
	apiUrl += path

	if len(rawQuery) > 0 {
		apiUrl += "?" + rawQuery
	}

	return apiUrl
}

func (p *Prerender) PreRenderHandlerFastHttp(ctx *fasthttp.RequestCtx) error {

	client := &http.Client{}
	req, err := http.NewRequest("GET", p.buildURLforFastHttp(ctx), nil)
	e.Check(err)

	if p.Options.Token != "" {
		req.Header.Set("X-Prerender-Token", p.Options.Token)
	}

	req.Header.Set("X-Original-User-Agent", string(ctx.Request.Header.Peek("User-Agent")))
	req.Header.Set("User-Agent", "prerender-cloud-golang-middleware")

	res, err := client.Do(req)
	e.Check(err)
	defer res.Body.Close()

	if res.StatusCode >= 500 && res.StatusCode <= 511 {
		fmt.Println("prerender.cloud server error", res.StatusCode)
		return errors.New("prerendercloud server error")
	}

	body, err := ioutil.ReadAll(res.Body)
	e.Check(err)

	fasthttp.CompressHandler(func(ctx *fasthttp.RequestCtx) {
		ctx.SetStatusCode(res.StatusCode)
		if len(res.Header["Content-Type"]) > 0 {
			ctx.SetContentType(res.Header["Content-Type"][0])
		}

		ctx.SetBody(body)
	})(ctx)

	return nil
}

// PreRenderHandler is a net/http compatible handler that proxies a request to
// the configured Prerender.cloud URL.  All upstream requests are made with an
// Accept-Encoding=gzip header.  Responses are provided either uncompressed or
// gzip compressed based on the downstream requests Accept-Encoding header
func (p *Prerender) PreRenderHandler(rw http.ResponseWriter, or *http.Request, next http.HandlerFunc) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", p.buildURLforHttp(or), nil)
	e.Check(err)

	if p.Options.Token != "" {
		req.Header.Set("X-Prerender-Token", p.Options.Token)
	}

	req.Header.Set("X-Original-User-Agent", or.Header.Get("User-Agent"))
	req.Header.Set("User-Agent", "prerender-cloud-golang-middleware")
	req.Header.Set("Content-Type", or.Header.Get("Content-Type"))
	req.Header.Set("Accept-Encoding", "gzip")

	if p.Options.UsingAppEngine {
		ctx := appengine.NewContext(or)
		client = urlfetch.Client(ctx)
	}

	res, err := client.Do(req)
	e.Check(err)
	defer res.Body.Close()

	if res.StatusCode >= 500 && res.StatusCode <= 511 && next != nil {
		fmt.Println("prerender.cloud server error", res.StatusCode)
		if next != nil {
			next(rw, or)
			return
		}
	}

	rw.Header().Set("Content-Type", res.Header.Get("Content-Type"))

	//Figure out whether the client accepts gzip responses
	doGzip := strings.Contains(or.Header.Get("Accept-Encoding"), "gzip")
	isGzip := strings.Contains(res.Header.Get("Content-Encoding"), "gzip")

	if doGzip && !isGzip {
		// gzip raw response
		rw.Header().Set("Content-Encoding", "gzip")
		rw.WriteHeader(res.StatusCode)
		gz := gzip.NewWriter(rw)
		defer gz.Close()
		io.Copy(gz, res.Body)
		gz.Flush()

	} else if !doGzip && isGzip {
		rw.WriteHeader(res.StatusCode)
		// gunzip response
		gz, err := gzip.NewReader(res.Body)
		e.Check(err)
		defer gz.Close()
		io.Copy(rw, gz)
	} else {
		// Pass through, gzip/gzip or raw/raw
		rw.Header().Set("Content-Encoding", res.Header.Get("Content-Encoding"))
		rw.WriteHeader(res.StatusCode)
		io.Copy(rw, res.Body)
	}

}
