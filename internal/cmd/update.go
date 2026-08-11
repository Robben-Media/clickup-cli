package cmd

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/builtbyrobben/clickup-cli/internal/outfmt"
	"github.com/builtbyrobben/clickup-cli/internal/selfupdate"
)

const (
	updateDefaultRepository       = "Robben-Media/clickup-cli"
	updateDefaultLatestReleaseURL = "https://api.github.com/repos/Robben-Media/clickup-cli/releases/latest"
	updateDefaultLatestWebURL     = "https://github.com/Robben-Media/clickup-cli/releases/latest"
	updateDefaultTimeout          = 10 * time.Second
	updateInstallUnknown          = "unknown"
)

var (
	errInvalidUpdateRepository = errors.New("invalid update repository")
	updateHTTPClient           = http.DefaultClient
	updateLatestReleaseURL     = updateDefaultLatestReleaseURL
	updateLatestWebURL         = updateDefaultLatestWebURL
	applySelfUpdate            = selfupdate.Apply
)

type UpdateCmd struct {
	Action      string        `arg:"" optional:"" help:"Optional read-only action: status or check"`
	Check       bool          `help:"Only check for an update; do not install"`
	ForceBinary bool          `name:"force-binary" help:"Allow replacing a development or dirty binary"`
	Timeout     time.Duration `name:"timeout" help:"HTTP timeout for GitHub release requests" default:"10s"`
}

type UpdateStatusCmd struct {
	Timeout time.Duration
}

type updateApplyReport struct {
	CurrentVersion  string `json:"current_version"`
	LatestVersion   string `json:"latest_version"`
	UpdateAvailable bool   `json:"update_available"`
	Applied         bool   `json:"applied"`
	PlatformAsset   string `json:"platform_asset,omitempty"`
}

type updateStatusReport struct {
	CurrentVersion      string   `json:"current_version"`
	CurrentCommit       string   `json:"current_commit,omitempty"`
	CurrentDate         string   `json:"current_date,omitempty"`
	LatestVersion       string   `json:"latest_version,omitempty"`
	LatestURL           string   `json:"latest_url,omitempty"`
	UpdateAvailable     bool     `json:"update_available"`
	Platform            string   `json:"platform"`
	PlatformAsset       string   `json:"platform_asset,omitempty"`
	PlatformAssetURL    string   `json:"platform_asset_url,omitempty"`
	ChecksumAvailable   bool     `json:"checksum_available"`
	ChecksumsURL        string   `json:"checksums_url,omitempty"`
	PlatformAssetSHA256 string   `json:"platform_asset_sha256,omitempty"`
	InstallMethod       string   `json:"install_method"`
	Executable          string   `json:"executable,omitempty"`
	Warnings            []string `json:"warnings,omitempty"`
	VersionsComparable  bool     `json:"-"`
}

type githubRelease struct {
	TagName         string               `json:"tag_name"`
	HTMLURL         string               `json:"html_url"`
	Assets          []githubReleaseAsset `json:"assets"`
	Repository      string               `json:"-"`
	SyntheticAssets bool                 `json:"-"`
}

type githubReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func (cmd *UpdateCmd) Run(ctx context.Context) error {
	if cmd.Action != "" && cmd.Action != "status" && cmd.Action != "check" {
		return newUsageError(fmt.Errorf("unknown update action %q (expected status or check)", cmd.Action))
	}

	readOnly := cmd.Check || cmd.Action != ""
	if readOnly {
		if cmd.ForceBinary {
			return newUsageError(fmt.Errorf("--force-binary cannot be used with check mode"))
		}
		return (&UpdateStatusCmd{Timeout: cmd.Timeout}).Run(ctx)
	}

	result, err := applySelfUpdate(ctx, selfupdate.ApplyOptions{
		Client:     newSelfUpdateClient(cmd.Timeout),
		CurrentVer: VersionString(),
		Force:      cmd.ForceBinary,
	})
	if err != nil {
		return fmt.Errorf("update clickup-cli: %w", err)
	}

	report := updateApplyReport{
		CurrentVersion:  result.Current,
		LatestVersion:   result.Latest,
		UpdateAvailable: result.Update,
		Applied:         result.Applied,
		PlatformAsset:   result.Asset,
	}
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, report)
	}
	if outfmt.IsPlain(ctx) {
		return writeUpdateApplyPlain(report)
	}

	if result.Applied {
		_, err = fmt.Fprintf(os.Stderr, "Updated clickup-cli: %s -> %s\n", result.Current, result.Latest)
	} else {
		_, err = fmt.Fprintf(os.Stderr, "No update needed: current %s, latest %s\n", result.Current, result.Latest)
	}
	if err != nil {
		return fmt.Errorf("write update result: %w", err)
	}
	return nil
}

func writeUpdateApplyPlain(report updateApplyReport) error {
	headers := []string{"CURRENT_VERSION", "LATEST_VERSION", "UPDATE_AVAILABLE", "APPLIED", "PLATFORM_ASSET"}
	rows := [][]string{{
		report.CurrentVersion,
		report.LatestVersion,
		strconv.FormatBool(report.UpdateAvailable),
		strconv.FormatBool(report.Applied),
		report.PlatformAsset,
	}}
	return outfmt.WritePlain(os.Stdout, headers, rows)
}

func newSelfUpdateClient(timeout time.Duration) *selfupdate.Client {
	return &selfupdate.Client{
		HTTP:  updateClient(timeout),
		Repo:  strings.TrimSpace(os.Getenv("CLICKUP_UPDATE_REPO")),
		Token: updateGitHubToken(),
	}
}

func updateGitHubToken() string {
	if token := strings.TrimSpace(os.Getenv("GITHUB_TOKEN")); token != "" {
		return token
	}
	return strings.TrimSpace(os.Getenv("GH_TOKEN"))
}

func setUpdateRequestHeaders(request *http.Request, acceptJSON bool) {
	request.Header.Set("User-Agent", "clickup-cli/"+VersionString())
	if acceptJSON {
		request.Header.Set("Accept", "application/vnd.github+json")
	}
	if token := updateGitHubToken(); token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
}

func (cmd *UpdateStatusCmd) Run(ctx context.Context) error {
	report, err := buildUpdateStatusReport(ctx, cmd.Timeout)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, report)
	}
	if outfmt.IsPlain(ctx) {
		return writeUpdateStatusPlain(report)
	}
	return writeUpdateStatusHuman(report)
}

func buildUpdateStatusReport(ctx context.Context, timeout time.Duration) (updateStatusReport, error) {
	current := VersionString()
	installMethod, executable, installWarnings := detectUpdateInstallMethod()
	report := updateStatusReport{
		CurrentVersion: current,
		CurrentCommit:  strings.TrimSpace(commit),
		CurrentDate:    strings.TrimSpace(date),
		Platform:       runtime.GOOS + "/" + runtime.GOARCH,
		InstallMethod:  installMethod,
		Executable:     executable,
		Warnings:       installWarnings,
	}

	client := updateClient(timeout)
	latestURL, latestWebURL, repository, err := updateStatusEndpoints()
	if err != nil {
		return updateStatusReport{}, err
	}
	release, err := fetchLatestGitHubReleaseForRepo(ctx, client, latestURL, latestWebURL, repository)
	if err != nil {
		return updateStatusReport{}, err
	}

	report.LatestVersion = strings.TrimSpace(release.TagName)
	report.LatestURL = strings.TrimSpace(release.HTMLURL)
	if report.LatestVersion == "" {
		report.Warnings = append(report.Warnings, "latest release did not include tag_name")
	}

	report.UpdateAvailable, report.VersionsComparable = updateAvailable(current, report.LatestVersion)
	if !report.VersionsComparable {
		report.Warnings = append(report.Warnings, "could not compare current and latest release versions")
	}

	report.PlatformAsset = platformAssetName(report.LatestVersion, runtime.GOOS, runtime.GOARCH)
	if report.PlatformAsset != "" {
		if asset, ok := findReleaseAsset(release.Assets, report.PlatformAsset); ok {
			report.PlatformAssetURL = asset.BrowserDownloadURL
		} else if !release.SyntheticAssets {
			report.Warnings = append(report.Warnings, "no release asset found for "+report.Platform)
		}
	}

	if checksumAsset, ok := findReleaseAsset(release.Assets, "checksums.txt"); ok {
		report.ChecksumsURL = checksumAsset.BrowserDownloadURL
		if report.PlatformAsset == "" {
			report.ChecksumAvailable = true
		} else {
			sum, checksumAvailable, checksumErr := fetchAssetChecksum(ctx, client, checksumAsset.BrowserDownloadURL, report.PlatformAsset)
			report.ChecksumAvailable = checksumAvailable || !release.SyntheticAssets
			if checksumErr != nil {
				report.Warnings = append(report.Warnings, checksumErr.Error())
				if release.SyntheticAssets {
					report.Warnings = append(report.Warnings, "could not verify release asset for "+report.Platform)
				}
			} else {
				report.PlatformAssetSHA256 = sum
				if release.SyntheticAssets {
					report.PlatformAssetURL = updateReleaseAssetURLForRepo(release.Repository, release.TagName, report.PlatformAsset)
				}
			}
		}
	} else {
		report.Warnings = append(report.Warnings, "checksums.txt not found on latest release")
		if release.SyntheticAssets {
			report.Warnings = append(report.Warnings, "could not verify release asset for "+report.Platform)
		}
	}

	return report, nil
}

func writeUpdateStatusPlain(report updateStatusReport) error {
	headers := []string{
		"CURRENT_VERSION", "CURRENT_COMMIT", "CURRENT_DATE", "LATEST_VERSION", "LATEST_URL",
		"UPDATE_AVAILABLE", "PLATFORM", "PLATFORM_ASSET", "PLATFORM_ASSET_URL",
		"CHECKSUM_AVAILABLE", "CHECKSUMS_URL", "PLATFORM_ASSET_SHA256", "INSTALL_METHOD", "EXECUTABLE", "WARNINGS",
	}
	rows := [][]string{{
		report.CurrentVersion, report.CurrentCommit, report.CurrentDate, report.LatestVersion, report.LatestURL,
		strconv.FormatBool(report.UpdateAvailable), report.Platform, report.PlatformAsset, report.PlatformAssetURL,
		strconv.FormatBool(report.ChecksumAvailable), report.ChecksumsURL, report.PlatformAssetSHA256,
		report.InstallMethod, report.Executable, strings.Join(report.Warnings, "; "),
	}}
	return outfmt.WritePlain(os.Stdout, headers, rows)
}

func writeUpdateStatusHuman(report updateStatusReport) error {
	var output bytes.Buffer
	switch {
	case !report.VersionsComparable:
		fmt.Fprintf(&output, "Could not determine update status for %s\n", report.CurrentVersion)
	case report.UpdateAvailable:
		fmt.Fprintf(&output, "Update available: %s -> %s\n", report.CurrentVersion, report.LatestVersion)
	default:
		fmt.Fprintf(&output, "clickup-cli is up to date (%s)\n", report.CurrentVersion)
	}
	fmt.Fprintf(&output, "  Latest:  %s\n", report.LatestURL)
	fmt.Fprintf(&output, "  Platform: %s\n", report.Platform)
	fmt.Fprintf(&output, "  Install:  %s\n", report.InstallMethod)
	if report.PlatformAsset != "" {
		fmt.Fprintf(&output, "  Asset:    %s\n", report.PlatformAsset)
	}
	if report.PlatformAssetSHA256 != "" {
		fmt.Fprintf(&output, "  SHA-256:  %s\n", report.PlatformAssetSHA256)
	}
	if _, err := io.Copy(os.Stdout, &output); err != nil {
		return fmt.Errorf("write update status: %w", err)
	}

	var warnings bytes.Buffer
	for _, warning := range report.Warnings {
		fmt.Fprintf(&warnings, "Warning: %s\n", warning)
	}
	if _, err := io.Copy(os.Stderr, &warnings); err != nil {
		return fmt.Errorf("write update warnings: %w", err)
	}
	return nil
}

func updateStatusEndpoints() (latestURL, latestWebURL, repository string, err error) {
	repository = strings.TrimSpace(os.Getenv("CLICKUP_UPDATE_REPO"))
	if repository == "" {
		return updateLatestReleaseURL, updateLatestWebURL, updateDefaultRepository, nil
	}

	parts := strings.Split(repository, "/")
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return "", "", "", fmt.Errorf("%w %q (expected owner/repository)", errInvalidUpdateRepository, repository)
	}
	repository = parts[0] + "/" + parts[1]
	escapedRepository := url.PathEscape(parts[0]) + "/" + url.PathEscape(parts[1])

	return "https://api.github.com/repos/" + escapedRepository + "/releases/latest",
		"https://github.com/" + escapedRepository + "/releases/latest",
		repository,
		nil
}

func updateClient(timeout time.Duration) *http.Client {
	if timeout <= 0 {
		timeout = updateDefaultTimeout
	}
	if updateHTTPClient == nil {
		return &http.Client{Timeout: timeout}
	}
	if updateHTTPClient.Timeout != 0 {
		return updateHTTPClient
	}
	clone := *updateHTTPClient
	clone.Timeout = timeout
	return &clone
}

func fetchLatestGitHubRelease(ctx context.Context, client *http.Client, endpoint string) (githubRelease, error) {
	return fetchLatestGitHubReleaseForRepo(ctx, client, endpoint, updateLatestWebURL, updateDefaultRepository)
}

func fetchLatestGitHubReleaseForRepo(
	ctx context.Context,
	client *http.Client,
	endpoint,
	latestWebURL,
	repository string,
) (githubRelease, error) {
	var release githubRelease
	apiErr := fetchUpdateJSON(ctx, client, endpoint, &release)
	if apiErr == nil {
		release.Repository = repository
		return release, nil
	}

	fallback, fallbackErr := fetchLatestGitHubReleaseRedirectForRepo(ctx, client, latestWebURL, repository)
	if fallbackErr != nil {
		return githubRelease{}, fmt.Errorf("fetch latest release: API: %s; web fallback: %w", apiErr.Error(), fallbackErr)
	}
	return fallback, nil
}

func fetchLatestGitHubReleaseRedirect(ctx context.Context, client *http.Client, latestURL string) (githubRelease, error) {
	return fetchLatestGitHubReleaseRedirectForRepo(ctx, client, latestURL, updateDefaultRepository)
}

func fetchLatestGitHubReleaseRedirectForRepo(
	ctx context.Context,
	client *http.Client,
	latestURL,
	repository string,
) (githubRelease, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, latestURL, nil)
	if err != nil {
		return githubRelease{}, err
	}
	setUpdateRequestHeaders(req, false)

	redirectClient := *client
	redirectClient.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}
	resp, err := redirectClient.Do(req)
	if err != nil {
		return githubRelease{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 300 || resp.StatusCode >= 400 {
		return githubRelease{}, fmt.Errorf("github returned %s", resp.Status)
	}
	location := strings.TrimSpace(resp.Header.Get("Location"))
	if location == "" {
		return githubRelease{}, fmt.Errorf("github redirect did not include Location")
	}
	resolved, err := req.URL.Parse(location)
	if err != nil {
		return githubRelease{}, fmt.Errorf("parse github redirect: %w", err)
	}
	if !strings.EqualFold(resolved.Hostname(), "github.com") {
		return githubRelease{}, fmt.Errorf("unexpected github redirect host %q", resolved.Hostname())
	}

	parts := strings.Split(repository, "/")
	if len(parts) != 2 {
		return githubRelease{}, fmt.Errorf("%w %q", errInvalidUpdateRepository, repository)
	}
	tagPath := "/" + url.PathEscape(parts[0]) + "/" + url.PathEscape(parts[1]) + "/releases/tag/"
	if !strings.HasPrefix(resolved.EscapedPath(), tagPath) {
		return githubRelease{}, fmt.Errorf("unexpected github release redirect path %q", resolved.EscapedPath())
	}
	tag, err := url.PathUnescape(strings.TrimPrefix(resolved.EscapedPath(), tagPath))
	if err != nil || tag == "" || strings.Contains(tag, "/") {
		return githubRelease{}, fmt.Errorf("invalid github release tag in redirect")
	}

	return githubRelease{
		TagName:         tag,
		HTMLURL:         resolved.String(),
		Repository:      repository,
		SyntheticAssets: true,
		Assets: []githubReleaseAsset{{
			Name:               "checksums.txt",
			BrowserDownloadURL: updateReleaseAssetURLForRepo(repository, tag, "checksums.txt"),
		}},
	}, nil
}

func updateReleaseAssetURLForRepo(repository, tag, assetName string) string {
	parts := strings.Split(repository, "/")
	if len(parts) != 2 {
		return ""
	}
	return "https://github.com/" + url.PathEscape(parts[0]) + "/" + url.PathEscape(parts[1]) + "/releases/download/" +
		url.PathEscape(tag) + "/" + url.PathEscape(assetName)
}

func fetchUpdateJSON(ctx context.Context, client *http.Client, endpoint string, dst any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	setUpdateRequestHeaders(req, true)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		message := strings.TrimSpace(string(body))
		if message == "" {
			message = resp.Status
		}
		return fmt.Errorf("github returned %s: %s", resp.Status, message)
	}
	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func fetchAssetChecksum(ctx context.Context, client *http.Client, endpoint, assetName string) (sum string, available bool, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", false, fmt.Errorf("fetch checksums.txt: %w", err)
	}
	setUpdateRequestHeaders(req, false)
	resp, err := client.Do(req)
	if err != nil {
		return "", false, fmt.Errorf("fetch checksums.txt: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", false, fmt.Errorf("fetch checksums.txt: github returned %s", resp.Status)
	}

	scanner := bufio.NewScanner(io.LimitReader(resp.Body, 1<<20))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}
		name := strings.TrimPrefix(fields[len(fields)-1], "*")
		if name != assetName {
			continue
		}
		sum := strings.ToLower(fields[0])
		decoded, decodeErr := hex.DecodeString(sum)
		if decodeErr != nil || len(decoded) != sha256.Size {
			return "", true, fmt.Errorf("invalid checksum for %s in checksums.txt", assetName)
		}
		return sum, true, nil
	}
	if err := scanner.Err(); err != nil {
		return "", true, fmt.Errorf("read checksums.txt: %w", err)
	}
	return "", true, fmt.Errorf("checksum for %s not found in checksums.txt", assetName)
}

func findReleaseAsset(assets []githubReleaseAsset, name string) (githubReleaseAsset, bool) {
	for _, asset := range assets {
		if asset.Name == name {
			return asset, true
		}
	}
	return githubReleaseAsset{}, false
}

func platformAssetName(tag, goos, goarch string) string {
	if strings.TrimSpace(selfupdate.NormalizeVersion(tag)) == "" {
		return ""
	}
	return selfupdate.AssetNameForPlatform(tag, goos, goarch)
}

func updateAvailable(current, latest string) (bool, bool) {
	comparison, ok := selfupdate.CompareVersions(current, latest)
	if !ok {
		return false, false
	}
	return comparison < 0, true
}

func detectUpdateInstallMethod() (method, executable string, warnings []string) {
	executable, err := os.Executable()
	if err != nil {
		return updateInstallUnknown, "", []string{"could not determine executable path: " + err.Error()}
	}
	if resolved, resolveErr := filepath.EvalSymlinks(executable); resolveErr == nil {
		executable = resolved
	}
	lower := strings.ToLower(executable)
	switch {
	case isDockerRuntime():
		method = "docker"
	case strings.Contains(lower, "/cellar/") || strings.Contains(lower, "/homebrew/") || strings.Contains(lower, "/linuxbrew/"):
		method = "homebrew"
	case strings.Contains(lower, string(filepath.Separator)+"go-build") || strings.HasSuffix(lower, ".test"):
		method = "source"
	default:
		method = "standalone"
	}
	return method, executable, nil
}

func isDockerRuntime() bool {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	data, err := os.ReadFile("/proc/1/cgroup")
	if err != nil {
		return false
	}
	text := strings.ToLower(string(data))
	return strings.Contains(text, "docker") || strings.Contains(text, "kubepods") || strings.Contains(text, "containerd")
}
