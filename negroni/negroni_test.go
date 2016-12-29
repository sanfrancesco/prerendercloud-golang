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

func Test_UserAgentRequest(t *testing.T) {
	httpmock.RegisterResponder("GET", "https://service.prerender.cloud/https://www.example.com/", httpmock.NewStringResponder(200, `prerendered response`))

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "https://www.example.com/", nil)
	req.Header.Set("User-Agent", "whatever")

	prerendercloud.NewOptions().NewPrerender().ServeHTTP(res, req, nil)

	if len(res.Body.Bytes()) == 0 {
		t.Error("Error, prerender.cloud should not have been called when the request had a user-agent present")
	}
}

func Test_UserAgentStaticResourceRequest(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "https://www.example.com/style.woff", nil)
	req.Header.Set("User-Agent", "whatever")

	prerendercloud.NewOptions().NewPrerender().ServeHTTP(res, req, nil)

	if len(res.Body.Bytes()) > 0 {
		t.Error("Error, prerender.cloud should not have been called for static resource")
	}
}

func Test_NoUserAgentRequest(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "https://www.example.com/", nil)

	prerendercloud.NewOptions().NewPrerender().ServeHTTP(res, req, nil)

	if len(res.Body.Bytes()) > 0 {
		t.Error("Error, prerender.cloud called for non-proxy request")
	}
}
