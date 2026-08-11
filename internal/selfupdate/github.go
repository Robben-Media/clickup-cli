package selfupdate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"strings"
)

const (
	defaultRepo    = "Robben-Media/clickup-cli"
	defaultBaseURL = "https://api.github.com"
	windows        = "windows"
)

var (
	errReleaseMissingTag = errors.New("release missing tag_name")
	errNoReleaseAsset    = errors.New("no release asset for platform")
	errGitHubRequest     = errors.New("github request failed")
	errDownloadFailed    = errors.New("download failed")
	errInvalidRepository = errors.New("invalid update repository")
)

// Release is the subset of a GitHub release needed for self-update.
type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

// Asset is a downloadable GitHub release asset.
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// Client fetches release metadata and assets.
type Client struct {
	HTTP    *http.Client
	Repo    string
	BaseURL string
	Token   string
}

func (c *Client) repo() string {
	if repo := strings.TrimSpace(c.Repo); repo != "" {
		return repo
	}

	return defaultRepo
}

func (c *Client) baseURL() string {
	if baseURL := strings.TrimSpace(c.BaseURL); baseURL != "" {
		return strings.TrimRight(baseURL, "/")
	}

	return defaultBaseURL
}

func (c *Client) httpClient() *http.Client {
	if c.HTTP != nil {
		return c.HTTP
	}

	return http.DefaultClient
}

func repositoryPath(repository string) (string, error) {
	parts := strings.Split(repository, "/")
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return "", fmt.Errorf("%w %q (expected owner/repository)", errInvalidRepository, repository)
	}

	return url.PathEscape(parts[0]) + "/" + url.PathEscape(parts[1]), nil
}

// LatestRelease fetches the repository's latest release.
func (c *Client) LatestRelease(ctx context.Context) (Release, error) {
	repository, err := repositoryPath(c.repo())
	if err != nil {
		return Release{}, err
	}
	endpoint := fmt.Sprintf("%s/repos/%s/releases/latest", c.baseURL(), repository)

	req, err := c.request(ctx, endpoint)
	if err != nil {
		return Release{}, fmt.Errorf("build release request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return Release{}, fmt.Errorf("fetch release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return Release{}, responseError(errGitHubRequest, resp)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return Release{}, fmt.Errorf("decode release: %w", err)
	}

	if strings.TrimSpace(release.TagName) == "" {
		return Release{}, errReleaseMissingTag
	}

	return release, nil
}

// Download writes a release asset to dst.
func (c *Client) Download(ctx context.Context, assetURL string, dst io.Writer) error {
	req, err := c.request(ctx, assetURL)
	if err != nil {
		return fmt.Errorf("build download request: %w", err)
	}

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return fmt.Errorf("download asset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return responseError(errDownloadFailed, resp)
	}

	if _, err := io.Copy(dst, resp.Body); err != nil {
		return fmt.Errorf("copy download body: %w", err)
	}

	return nil
}

func (c *Client) request(ctx context.Context, endpoint string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create HTTP request: %w", err)
	}

	req.Header.Set("User-Agent", "clickup-cli-selfupdate")

	if token := strings.TrimSpace(c.Token); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return req, nil
}

func responseError(kind error, resp *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))

	message := strings.TrimSpace(string(body))
	if message == "" {
		return fmt.Errorf("%w: %s", kind, resp.Status)
	}

	return fmt.Errorf("%w: %s: %s", kind, resp.Status, message)
}

// NormalizeVersion removes a release tag's optional leading v.
func NormalizeVersion(version string) string {
	return strings.TrimPrefix(strings.TrimSpace(version), "v")
}

// AssetNameFor returns the current platform's GoReleaser asset name.
func AssetNameFor(version string) string {
	return AssetNameForPlatform(version, runtime.GOOS, runtime.GOARCH)
}

// AssetNameForPlatform returns a platform's GoReleaser asset name.
func AssetNameForPlatform(version, goos, goarch string) string {
	extension := ".tar.gz"
	if goos == windows {
		extension = ".zip"
	}

	return fmt.Sprintf("clickup-cli_%s_%s_%s%s", NormalizeVersion(version), goos, goarch, extension)
}

func findAssets(release Release, goos, goarch string) (Asset, Asset, error) {
	wanted := AssetNameForPlatform(release.TagName, goos, goarch)
	var archive, checksums Asset

	for _, asset := range release.Assets {
		switch asset.Name {
		case wanted:
			archive = asset
		case "checksums.txt":
			checksums = asset
		}
	}

	if archive.Name == "" || strings.TrimSpace(archive.BrowserDownloadURL) == "" {
		return Asset{}, Asset{}, fmt.Errorf("%w: %s/%s (looked for %s)", errNoReleaseAsset, goos, goarch, wanted)
	}

	return archive, checksums, nil
}
