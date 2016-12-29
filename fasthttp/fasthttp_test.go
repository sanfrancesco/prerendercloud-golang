package fasthttp_test

import (
	"bufio"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"testing"

	prerendercloud "github.com/sanfrancesco/prerendercloud-golang"
	"gopkg.in/jarcoal/httpmock.v1"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
)

var listener *fasthttputil.InmemoryListener

func makeRequest(url string) ([]byte, error) {
	req, _ := http.NewRequest("GET", url, nil)
	dump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		return nil, err
	}

	c, err := listener.Dial()
	if err != nil {
		return nil, err
	}

	if _, err = c.Write(dump); err != nil {
		return nil, err
	}

	br := bufio.NewReader(c)
	var resp fasthttp.Response
	if err = resp.Read(br); err != nil {
		return nil, err
	}

	var body []byte
	if string(resp.Header.Peek("Content-Encoding")) == "gzip" {
		body, err = resp.BodyGunzip()
		if err != nil {
			return nil, err
		}
	} else {
		body = resp.Body()
	}

	return body, nil
}

func TestMain(m *testing.M) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	prerenderCloudOptions := prerendercloud.NewOptions()
	prerenderCloud := prerenderCloudOptions.NewPrerender()

	server := &fasthttp.Server{
		Handler: func(ctx *fasthttp.RequestCtx) {
			if prerenderCloud.ShouldPrerenderFastHttp(ctx) {
				prerenderCloud.PreRenderHandlerFastHttp(ctx)
			} else {
				ctx.SetContentType("text/html")
				fmt.Fprintf(ctx, `origin`)
			}
		},
		Name: "test server",
	}

	listener = fasthttputil.NewInmemoryListener()

	serverCh := make(chan struct{})
	go func() {

		if err := server.Serve(listener); err != nil {
			panic(err)
		}
		close(serverCh)

	}()

	os.Exit(m.Run())
}

func Test_fasthttp(t *testing.T) {
	httpmock.RegisterResponder("GET", "https://service.prerender.cloud/http://www.example.com/deep/path", httpmock.NewStringResponder(200, `prerendered response`))

	body, err := makeRequest("http://www.example.com/deep/path")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if string(body) != "prerendered response" {
		t.Error("expected prerendered response")
	}
}

func Test_fasthttpHtmlExtension(t *testing.T) {
	httpmock.RegisterResponder("GET", "https://service.prerender.cloud/http://www.example.com/deep/path.html", httpmock.NewStringResponder(200, `prerendered response`))

	body, err := makeRequest("http://www.example.com/deep/path.html")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if string(body) != "prerendered response" {
		t.Error("expected prerendered response")
	}
}

func Test_fasthttpNoExtension(t *testing.T) {
	httpmock.RegisterResponder("GET", "https://service.prerender.cloud/http://www.example.com/deep/path", httpmock.NewStringResponder(200, `prerendered response`))

	body, err := makeRequest("http://www.example.com/deep/path")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if string(body) != "prerendered response" {
		t.Error("expected prerendered response")
	}
}

func Test_fasthttpFontExtensionStaticResource(t *testing.T) {
	body, err := makeRequest("http://www.example.com/deep/path.woff")
	if err != nil {
		panic(err)
		t.Fatalf("unexpected error: %s", err)
	}

	if string(body) != "origin" {
		t.Error("expected prerendered response")
	}
}
