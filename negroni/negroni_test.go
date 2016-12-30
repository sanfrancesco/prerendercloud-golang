package negroni

import (
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
	httpmock.RegisterResponder("GET", "https://service.prerender.cloud/http://www.example.com/", httpmock.NewStringResponder(200, `prerendered response`))

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://www.example.com/", nil)
	req.Header.Set("User-Agent", "example-user-agent")

	prerendercloud.NewOptions().NewPrerender().ServeHTTP(res, req, nil)

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
	httpmock.RegisterResponder("GET", "https://service.prerender.cloud/http://www.example.com/deep/path.html", httpmock.NewStringResponder(200, `prerendered response`))

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://www.example.com/deep/path.html", nil)
	req.Header.Set("User-Agent", "example-user-agent")

	prerendercloud.NewOptions().NewPrerender().ServeHTTP(res, req, nil)

	if len(res.Body.Bytes()) == 0 {
		t.Error("Error, prerender.cloud should have been called when the request had an HTML extension")
	}
}

func withNoExtension(t *testing.T) {
	httpmock.RegisterResponder("GET", "https://service.prerender.cloud/http://www.example.com/deep/path", httpmock.NewStringResponder(200, `prerendered response`))

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
