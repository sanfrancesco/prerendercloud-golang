package prerender

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"gopkg.in/jarcoal/httpmock.v1"
)

func TestMain(m *testing.M) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	os.Exit(m.Run())
}

func Test_prerenderableExtension(t *testing.T) {
	if prerenderableExtension("") != true {
		t.Error("empty string should be prerenderable")
	}

	if prerenderableExtension("index.html") != true {
		t.Error("/index.html should be prerenderable")
	}

	if prerenderableExtension("index.htm") != true {
		t.Error("/index.html should be prerenderable")
	}

	if prerenderableExtension("/") != true {
		t.Error("a slash should be prerenderable")
	}

	if prerenderableExtension("/index.html") != true {
		t.Error("/index.html should be prerenderable")
	}

	if prerenderableExtension("/index.htm") != true {
		t.Error("/index.html should be prerenderable")
	}

	if prerenderableExtension("root") != true {
		t.Error("root should be prerenderable")
	}

	if prerenderableExtension("font.woff") != false {
		t.Error("font.woff should not be prerenderable")
	}

	if prerenderableExtension("assets/font.woff") != false {
		t.Error("assets/font.woff should not be prerenderable")
	}
}

func Test_UserAgentRequest(t *testing.T) {
	httpmock.RegisterResponder("GET", "https://service.prerender.cloud/https://www.example.com/", httpmock.NewStringResponder(200, `prerendered response`))

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "https://www.example.com/", nil)
	req.Header.Set("User-Agent", "whatever")

	NewOptions().NewPrerender().ServeHTTP(res, req, nil)

	if len(res.Body.Bytes()) == 0 {
		t.Error("Error, prerender.cloud not called")
	}
}

func Test_UserAgentStaticResourceRequest(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "https://www.example.com/style.woff", nil)
	req.Header.Set("User-Agent", "whatever")

	NewOptions().NewPrerender().ServeHTTP(res, req, nil)

	if len(res.Body.Bytes()) > 0 {
		t.Error("Error, prerender.cloud called for non-proxy request")
	}
}

func Test_NoUserAgentRequest(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "https://www.example.com/", nil)

	NewOptions().NewPrerender().ServeHTTP(res, req, nil)

	if len(res.Body.Bytes()) > 0 {
		t.Error("Error, prerender.cloud called for non-proxy request")
	}
}
