package negroni

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	prerendercloud "github.com/sanfrancesco/prerendercloud-golang"
	"gopkg.in/jarcoal/httpmock.v1"
)

func TestMain(m *testing.M) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	os.Exit(m.Run())
}

func Test_NoUserAgentRequest(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://www.example.com/", nil)

	prerendercloud.NewOptions().NewPrerender().ServeHTTP(res, req, nil)

	if len(res.Body.Bytes()) > 0 {
		t.Error("Error, prerender.cloud should not have been called when there is no user-agent")
	}
}

func Test_WithUserAgentRequest(t *testing.T) {
	httpmock.RegisterResponder("GET", "https://service.headless-render-api.com/http://www.example.com/", httpmock.NewStringResponder(201, `prerendered response`))

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://www.example.com/", nil)
	req.Header.Set("User-Agent", "example-user-agent")

	prerendercloud.NewOptions().NewPrerender().ServeHTTP(res, req, nil)

	if res.Result().StatusCode != 201 {
		fmt.Printf("actual StatusCode %#v\n", res.Result().StatusCode)
		t.Error("expected prerender.cloud statusCode to be preserved")
	}

	if len(res.Body.Bytes()) == 0 {
		t.Error("Error, prerender.cloud should have been called when the request had a user-agent present")
	}
}

func Test_WithPrerendercloudUserAgentRequest(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://www.example.com/", nil)
	req.Header.Set("User-Agent", "prerendercloud")

	prerendercloud.NewOptions().NewPrerender().ServeHTTP(res, req, nil)

	if len(res.Body.Bytes()) > 0 {
		t.Error("Error, prerender.cloud should not have been called when the request had the prerendercloud user-agent present")
	}
}

func Test_withHtmlExtension(t *testing.T) {
	httpmock.RegisterResponder("GET", "https://service.headless-render-api.com/http://www.example.com/deep/path.html", httpmock.NewStringResponder(200, `prerendered response`))

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://www.example.com/deep/path.html", nil)
	req.Header.Set("User-Agent", "example-user-agent")

	prerendercloud.NewOptions().NewPrerender().ServeHTTP(res, req, nil)

	if len(res.Body.Bytes()) == 0 {
		t.Error("Error, prerender.cloud should have been called when the request had an HTML extension")
	}
}

func withNoExtension(t *testing.T) {
	httpmock.RegisterResponder("GET", "https://service.headless-render-api.com/http://www.example.com/deep/path", httpmock.NewStringResponder(200, `prerendered response`))

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://www.example.com/deep/path", nil)
	req.Header.Set("User-Agent", "example-user-agent")

	prerendercloud.NewOptions().NewPrerender().ServeHTTP(res, req, nil)

	if len(res.Body.Bytes()) == 0 {
		t.Error("Error, prerender.cloud should have been called when the request had no extension")
	}
}

func Test_WithUserAgentAndStaticResourceRequest(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://www.example.com/style.woff", nil)
	req.Header.Set("User-Agent", "example-user-agent")

	prerendercloud.NewOptions().NewPrerender().ServeHTTP(res, req, nil)

	if len(res.Body.Bytes()) > 0 {
		t.Error("Error, prerender.cloud should not have been called for static resource")
	}
}

func Test_WithUserAgentAndAlreadyPrerenderedRequest(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://www.example.com/", nil)
	req.Header.Set("User-Agent", "example-user-agent")
	req.Header.Set("X-PrerenderEd", "true")

	prerendercloud.NewOptions().NewPrerender().ServeHTTP(res, req, nil)

	if len(res.Body.Bytes()) > 0 {
		t.Error("Error, prerender.cloud should not have been called if x-prerendered was true")
	}
}

func Test_WithServerErrorAndNextMiddleware(t *testing.T) {
	httpmock.RegisterResponder("GET", "https://service.headless-render-api.com/http://www.example.com/", httpmock.NewStringResponder(500, `server error`))

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://www.example.com/", nil)
	req.Header.Set("User-Agent", "example-user-agent")

	prerendercloud.NewOptions().NewPrerender().ServeHTTP(res, req, func(res http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(res, "next middleware")
	})

	if res.Result().StatusCode != 200 {
		fmt.Printf("actual StatusCode %#v\n", res.Result().StatusCode)
		t.Error("Error, middleware should return 200 response when server returns 500")
	}

	if string(res.Body.Bytes()) != "next middleware" {
		fmt.Printf("actual response %#v\n", string(res.Body.Bytes()))
		t.Error("Error, middleware should return response from next middleware when server returns 500")
	}
}

func Test_WithServerErrorAndNoNextMiddleware(t *testing.T) {
	httpmock.RegisterResponder("GET", "https://service.headless-render-api.com/http://www.example.com/", httpmock.NewStringResponder(501, `server error`))

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://www.example.com/", nil)
	req.Header.Set("User-Agent", "example-user-agent")

	prerendercloud.NewOptions().NewPrerender().ServeHTTP(res, req, nil)

	if res.Result().StatusCode != 501 {
		fmt.Printf("actual StatusCode %#v\n", res.Result().StatusCode)
		t.Error("Error, middleware should return server's 501 response when there's no next middleware")
	}

	if string(res.Body.Bytes()) != "server error" {
		fmt.Printf("actual response %#v\n", string(res.Body.Bytes()))
		t.Error("Error, middleware should return response from next middleware when server returns 500")
	}
}
