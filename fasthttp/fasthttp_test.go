package fasthttp_test

import (
	"bufio"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"testing"

	"gopkg.in/jarcoal/httpmock.v1"

	"github.com/sanfrancesco/prerendercloud-golang"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
)

var listener *fasthttputil.InmemoryListener

func makeRequest(url string, alreadyPrerendered bool, userAgent string) ([]byte, error) {
	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Set("User-Agent", userAgent)

	if alreadyPrerendered {
		req.Header.Set("X-PrerenderEd", "true")
	}

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

func Test_NoUserAgentRequest(t *testing.T) {
	body, err := makeRequest("http://www.example.com/", false, "")
	if err != nil {
		panic(err)
		t.Fatalf("unexpected error: %s", err)
	}

	if string(body) != "origin" {
		t.Error("expected origin response")
	}
}

func Test_WithUserAgentRequest(t *testing.T) {
	httpmock.RegisterResponder("GET", "https://service.prerender.cloud/http://www.example.com/", httpmock.NewStringResponder(200, `prerendered response`))

	body, err := makeRequest("http://www.example.com/", false, "example-user-agent")
	if err != nil {
		panic(err)
		t.Fatalf("unexpected error: %s", err)
	}

	if string(body) != "prerendered response" {
		t.Error("expected prerendered response")
	}
}

func Test_WithPrerendercloudUserAgentRequest(t *testing.T) {
	body, err := makeRequest("http://www.example.com/", false, "prerendercloud")
	if err != nil {
		panic(err)
		t.Fatalf("unexpected error: %s", err)
	}

	if string(body) != "origin" {
		t.Error("expected origin response")
	}
}

func Test_withHtmlExtension(t *testing.T) {
	httpmock.RegisterResponder("GET", "https://service.prerender.cloud/http://www.example.com/deep/path.html", httpmock.NewStringResponder(200, `prerendered response`))

	body, err := makeRequest("http://www.example.com/deep/path.html", false, "example-user-agent")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if string(body) != "prerendered response" {
		t.Error("expected prerendered response")
	}
}

func withNoExtension(t *testing.T) {
	httpmock.RegisterResponder("GET", "https://service.prerender.cloud/http://www.example.com/deep/path", httpmock.NewStringResponder(200, `prerendered response`))

	body, err := makeRequest("http://www.example.com/deep/path", false, "example-user-agent")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if string(body) != "prerendered response" {
		t.Error("expected prerendered response")
	}
}

func Test_WithUserAgentAndStaticResourceRequest(t *testing.T) {
	body, err := makeRequest("http://www.example.com/deep/path.woff", false, "example-user-agent")
	if err != nil {
		panic(err)
		t.Fatalf("unexpected error: %s", err)
	}

	if string(body) != "origin" {
		t.Error("expected origin response")
	}
}

func Test_WithUserAgentAndAlreadyPrerenderedRequest(t *testing.T) {
	body, err := makeRequest("http://www.example.com/", true, "example-user-agent")
	if err != nil {
		panic(err)
		t.Fatalf("unexpected error: %s", err)
	}

	if string(body) != "origin" {
		t.Error("expected origin response")
	}
}
