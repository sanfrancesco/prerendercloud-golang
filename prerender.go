// Package prerender provides a Prerender.io handler implementation and a
// Negroni middleware.
package prerender

import (
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	e "github.com/jqatampa/gadget-arm/errors"
	"github.com/valyala/fasthttp"
	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
)

// Options provides you with the ability to specify a custom Prerender.io URL
// as well as a Prerender.io Token to include as an X-Prerender-Token header
// to the upstream server.
type Options struct {
	PrerenderURL   *url.URL
	Token          string
	BlackList      []regexp.Regexp
	WhiteList      []regexp.Regexp
	UsingAppEngine bool
}

// NewOptions generates a default Options struct pointing to the Prerender.io
// service, obtaining a Token from the environment variable PRERENDER_TOKEN.
// No blacklist/whitelist is created.
func NewOptions() *Options {
	url, _ := url.Parse("https://service.prerender.cloud/")
	return &Options{
		PrerenderURL:   url,
		Token:          os.Getenv("PRERENDER_TOKEN"),
		BlackList:      nil,
		WhiteList:      nil,
		UsingAppEngine: false,
	}
}

// Prerender exposes methods to validate and serve content from a Prerender.io
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
	fmt.Println("Prerender")
	if p.ShouldPrerender(r) {
		p.PreRenderHandler(rw, r)
	} else if next != nil {
		next(rw, r)
	}
}

// ShouldPrerender analyzes the request to determine whether it should be routed
// to a Prerender.io upstream server.
func (p *Prerender) ShouldPrerender(or *http.Request) bool {
	fmt.Println(or)
	userAgent := strings.ToLower(or.Header.Get("User-Agent"))
	bufferAgent := or.Header.Get("X-Bufferbot")
	isRequestingPrerenderedPage := false
	reqURL := strings.ToLower(or.URL.String())

	// No user agent, don't prerender
	if userAgent == "" {
		return false
	}

	// No user agent, don't prerender
	if userAgent == "prerendercloud" {
		return false
	}

	// Not a GET or HEAD request, don't prerender
	if or.Method != "GET" && or.Method != "HEAD" {
		return false
	}

	// Static resource, don't prerender
	for _, extension := range skippedTypes {
		if strings.HasSuffix(reqURL, strings.ToLower(extension)) {
			return false
		}
	}

	// Buffer Agent or requesting an excaped fragment, request prerender
	if bufferAgent != "" || or.URL.Query().Get("_escaped_fragment_") != "" {
		isRequestingPrerenderedPage = true
	}

	// Cralwer, request prerender
	for _, crawlerAgent := range crawlerUserAgents {
		if strings.Contains(crawlerAgent, strings.ToLower(userAgent)) {
			isRequestingPrerenderedPage = true
			break
		}
	}

	// If it's a bot/crawler/escaped fragment request apply Blacklist/Whitelist logic
	if isRequestingPrerenderedPage {
		if p.Options.WhiteList != nil {
			matchFound := false
			for _, val := range p.Options.WhiteList {
				if val.MatchString(reqURL) {
					matchFound = true
					break
				}
			}
			if !matchFound {
				return false
			}
		}

		if p.Options.BlackList != nil {
			matchFound := false
			for _, val := range p.Options.BlackList {
				if val.MatchString(reqURL) {
					matchFound = true
					break
				}
			}
			if matchFound {
				return false
			}
		}
	}

	return isRequestingPrerenderedPage
}

func (p *Prerender) buildURLforFastHttp(ctx *fasthttp.RequestCtx) string {
	url := p.Options.PrerenderURL

	if !strings.HasSuffix(url.String(), "/") {
		url.Path = url.Path + "/"
	}

	protocol := string(ctx.URI().Scheme())

	if len(protocol) == 0 {
		protocol = "http"
	}

	if fp := string(ctx.Request.Header.Peek("X-Forwarded-Proto")); fp != "" {
		protocol = strings.Split(fp, ",")[0]
	}

	return url.String() + protocol + "://" + string(ctx.Host()) + string(ctx.Path()) + "?" + string(ctx.URI().QueryString())

}

func (p *Prerender) buildURL(or *http.Request) string {
	url := p.Options.PrerenderURL

	if !strings.HasSuffix(url.String(), "/") {
		url.Path = url.Path + "/"
	}

	var protocol = or.URL.Scheme

	if cf := or.Header.Get("CF-Visitor"); cf != "" {
		match := cfSchemeRegex.FindStringSubmatch(cf)
		if len(match) > 1 {
			protocol = match[1]
		}
	}

	if len(protocol) == 0 {
		protocol = "http"
	}

	if fp := or.Header.Get("X-Forwarded-Proto"); fp != "" {
		protocol = strings.Split(fp, ",")[0]
	}

	apiURL := url.String() + protocol + "://" + or.Host + or.URL.Path + "?" +
		or.URL.RawQuery
	return apiURL
}

func (p *Prerender) PreRenderHandlerFastHttp(ctx *fasthttp.RequestCtx) {
	fasthttp.CompressHandler(func(ctx *fasthttp.RequestCtx) {
		client := &http.Client{}
		req, err := http.NewRequest("GET", p.buildURLforFastHttp(ctx), nil)
		e.Check(err)

		if p.Options.Token != "" {
			ctx.Response.Header.Set("X-Prerender-Token", p.Options.Token)
		}

		res, err := client.Do(req)

		e.Check(err)

		defer res.Body.Close()

		body, err := ioutil.ReadAll(res.Body)
		e.Check(err)

		if len(res.Header["Content-Type"]) > 0 {
			ctx.SetContentType(res.Header["Content-Type"][0])
		}

		ctx.SetBody(body)
	})(ctx)
}

// PreRenderHandler is a net/http compatible handler that proxies a request to
// the configured Prerender.io URL.  All upstream requests are made with an
// Accept-Encoding=gzip header.  Responses are provided either uncompressed or
// gzip compressed based on the downstream requests Accept-Encoding header
func (p *Prerender) PreRenderHandler(rw http.ResponseWriter, or *http.Request) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", p.buildURL(or), nil)
	e.Check(err)

	if p.Options.Token != "" {
		req.Header.Set("X-Prerender-Token", p.Options.Token)
	}
	req.Header.Set("User-Agent", or.Header.Get("User-Agent"))
	req.Header.Set("Content-Type", or.Header.Get("Content-Type"))
	req.Header.Set("Accept-Encoding", "gzip")

	if p.Options.UsingAppEngine {
		ctx := appengine.NewContext(or)
		client = urlfetch.Client(ctx)
	}

	res, err := client.Do(req)

	fmt.Println(res)
	e.Check(err)

	rw.Header().Set("Content-Type", res.Header.Get("Content-Type"))

	defer res.Body.Close()

	//Figure out whether the client accepts gzip responses
	doGzip := strings.Contains(or.Header.Get("Accept-Encoding"), "gzip")
	isGzip := strings.Contains(res.Header.Get("Content-Encoding"), "gzip")

	if doGzip && !isGzip {
		// gzip raw response
		rw.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(rw)
		defer gz.Close()
		io.Copy(gz, res.Body)
		gz.Flush()

	} else if !doGzip && isGzip {
		// gunzip response
		gz, err := gzip.NewReader(res.Body)
		e.Check(err)
		defer gz.Close()
		io.Copy(rw, gz)
	} else {
		// Pass through, gzip/gzip or raw/raw
		rw.Header().Set("Content-Encoding", res.Header.Get("Content-Encoding"))
		io.Copy(rw, res.Body)

	}
}
