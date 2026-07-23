package update

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type GitHubClient struct {
	BaseURL, RawBaseURL string
	HTTP                *http.Client
}

func NewGitHubClient(owner, repo string, httpClient *http.Client) GitHubClient {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return GitHubClient{BaseURL: "https://api.github.com/repos/" + owner + "/" + repo, RawBaseURL: "https://raw.githubusercontent.com/" + owner + "/" + repo + "/main", HTTP: httpClient}
}

type githubRelease struct {
	Tag        string `json:"tag_name"`
	Notes      string `json:"body"`
	URL        string `json:"html_url"`
	Prerelease bool   `json:"prerelease"`
	Assets     []struct {
		Name, URL  string
		BrowserURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func (c GitHubClient) Releases(ctx context.Context) ([]Release, error) {
	var raw []githubRelease
	if err := c.getJSON(ctx, c.BaseURL+"/releases", &raw); err != nil {
		return nil, err
	}
	result := make([]Release, len(raw))
	for i, r := range raw {
		result[i] = Release{Tag: r.Tag, Notes: r.Notes, URL: r.URL, Prerelease: r.Prerelease}
		for _, a := range r.Assets {
			result[i].Assets = append(result[i].Assets, Asset{Name: a.Name, URL: a.BrowserURL})
		}
	}
	return result, nil
}
func (c GitHubClient) Policy(ctx context.Context) (Policy, error) {
	var p Policy
	err := c.getJSON(ctx, c.RawBaseURL+"/update-policy.json", &p)
	return p, err
}
func (c GitHubClient) Checksums(ctx context.Context, r Release) (map[string]Asset, error) {
	name := "wslc-tui-" + r.Tag + "-checksums.json"
	var manifest struct {
		Assets []Asset `json:"assets"`
	}
	var target Asset
	for _, a := range r.Assets {
		if a.Name == name {
			target = a
			break
		}
	}
	if target.URL == "" {
		return nil, fmt.Errorf("release %s has no checksum manifest", r.Tag)
	}
	if err := c.getJSON(ctx, target.URL, &manifest); err != nil {
		return nil, err
	}
	result := map[string]Asset{}
	for _, a := range manifest.Assets {
		result[a.Name] = a
	}
	return result, nil
}
func (c GitHubClient) getJSON(ctx context.Context, url string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("update server returned %s", resp.Status)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if strings.TrimSpace(string(b)) == "" {
		return fmt.Errorf("empty update response")
	}
	return json.Unmarshal(b, out)
}
