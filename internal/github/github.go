package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Client struct {
	httpClient *http.Client
	token      string
	cacheDir   string
	noCache    bool
	mu         sync.Mutex
}

func NewClient(token string, noCache bool) *Client {
	cacheDir := ""
	if home, err := os.UserHomeDir(); err == nil {
		cacheDir = filepath.Join(home, ".cache", "nvimgotrack")
	}

	return &Client{
		httpClient: &http.Client{Timeout: 15 * time.Second},
		token:      token,
		cacheDir:   cacheDir,
		noCache:    noCache,
	}
}

func (c *Client) GetRepoInfo(owner, repo string) (*RepoInfo, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)
	var info RepoInfo
	if err := c.get(url, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

func (c *Client) GetReleases(owner, repo string) ([]Release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases?per_page=30", owner, repo)
	var releases []Release
	if err := c.get(url, &releases); err != nil {
		return nil, err
	}
	return releases, nil
}

func (c *Client) CompareCommits(owner, repo, base, head string) (*CompareResult, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/compare/%s...%s", owner, repo, base, head)
	var result CompareResult
	if err := c.get(url, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) get(url string, target any) error {
	// Try cache first
	if !c.noCache {
		if data, err := c.readCache(url); err == nil {
			return json.Unmarshal(data, target)
		}
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "nvimgotrack/1.0")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("not found: %s", url)
	}
	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == 429 {
		return fmt.Errorf("rate limited — set GITHUB_TOKEN env var for higher limits")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body[:min(200, len(body))]))
	}

	if !c.noCache {
		c.writeCache(url, body)
	}

	return json.Unmarshal(body, target)
}

func (c *Client) readCache(url string) ([]byte, error) {
	path := c.cachePath(url)
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if time.Since(info.ModTime()) > 1*time.Hour {
		return nil, fmt.Errorf("cache expired")
	}
	return os.ReadFile(path)
}

func (c *Client) writeCache(url string, data []byte) {
	path := c.cachePath(url)
	dir := filepath.Dir(path)
	c.mu.Lock()
	defer c.mu.Unlock()
	_ = os.MkdirAll(dir, 0755)
	_ = os.WriteFile(path, data, 0644)
}

func (c *Client) cachePath(url string) string {
	safe := ""
	for _, r := range url {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			safe += string(r)
		} else {
			safe += "_"
		}
	}
	if len(safe) > 200 {
		safe = safe[:200]
	}
	return filepath.Join(c.cacheDir, safe+".json")
}
