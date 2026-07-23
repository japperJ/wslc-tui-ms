package update

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGitHubClientContract(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/releases":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`[{"tag_name":"v2.0.0","body":"notes","html_url":"https://example/release","prerelease":false,"assets":[{"name":"wslc-tui-v2.0.0-checksums.json","browser_download_url":"https://example/checksums"}]}]`))
		case "/checksums":
			w.Write([]byte(`{"assets":[{"name":"payload.zip","size":7,"sha256":"abc"}]}`))
		default:
			w.Write([]byte(`{"minimumSupportedVersion":"1.0.0","message":"upgrade"}`))
		}
	}))
	defer server.Close()
	c := GitHubClient{BaseURL: server.URL, RawBaseURL: server.URL, HTTP: server.Client()}
	if _, err := c.Releases(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestGitHubClientRejectsHTTPFailure(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()
	c := GitHubClient{BaseURL: server.URL, HTTP: server.Client()}
	if _, err := c.Releases(context.Background()); err == nil {
		t.Fatal("expected HTTP error")
	}
}
