package prerendercloud

import "testing"

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

func Test_buildUrlWithoutProtocol(t *testing.T) {
	apiUrl := buildApiUrl("https://service.headless-render-api.com", "", "example.org", "", "")
	if apiUrl != "https://service.headless-render-api.com/http://example.org" {
		t.Error("malformed API URL")
	}
}

func Test_buildUrlWithProtocol(t *testing.T) {
	apiUrl := buildApiUrl("https://service.headless-render-api.com", "https", "example.org", "", "")
	if apiUrl != "https://service.headless-render-api.com/https://example.org" {
		t.Error("malformed API URL")
	}
}

func Test_buildUrlWithQuery(t *testing.T) {
	apiUrl := buildApiUrl("https://service.headless-render-api.com", "https", "example.org", "", "val=true")
	if apiUrl != "https://service.headless-render-api.com/https://example.org?val=true" {
		t.Error("malformed API URL")
	}
}

func Test_buildUrlWithQueryAndPath(t *testing.T) {
	apiUrl := buildApiUrl("https://service.headless-render-api.com", "https", "example.org", "/", "val=true")
	if apiUrl != "https://service.headless-render-api.com/https://example.org/?val=true" {
		t.Error("malformed API URL")
	}
}
