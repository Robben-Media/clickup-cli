package cmd

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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
)

const (
	updateDefaultLatestReleaseURL = "https://api.github.com/repos/Robben-Media/clickup-cli/releases/latest"
	updateDefaultLatestWebURL     = "https://github.com/Robben-Media/clickup-cli/releases/latest"
	updateDefaultTimeout          = 10 * time.Second
	updateInstallUnknown          = "unknown"
)

var (
	updateHTTPClient       = http.DefaultClient
	updateLatestReleaseURL = updateDefaultLatestReleaseURL
	updateLatestWebURL     = updateDefaultLatestWebURL
)

type UpdateCmd struct {
	Status UpdateStatusCmd `cmd:"" name:"status" aliases:"check" help:"Show installed and latest clickup-cli release status"`
}

type UpdateStatusCmd struct {
	Timeout time.Duration `name:"timeout" help:"HTTP timeout for GitHub release metadata" default:"10s"`
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
	SyntheticAssets bool                 `json:"-"`
}

type githubReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
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
	release, err := fetchLatestGitHubRelease(ctx, client, updateLatestReleaseURL)
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
					report.PlatformAssetURL = updateReleaseAssetURL(release.TagName, report.PlatformAsset)
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
	var release githubRelease
	apiErr := fetchUpdateJSON(ctx, client, endpoint, &release)
	if apiErr == nil {
		return release, nil
	}

	fallback, fallbackErr := fetchLatestGitHubReleaseRedirect(ctx, client, updateLatestWebURL)
	if fallbackErr != nil {
		return githubRelease{}, fmt.Errorf("fetch latest release: API: %s; web fallback: %w", apiErr.Error(), fallbackErr)
	}
	return fallback, nil
}

func fetchLatestGitHubReleaseRedirect(ctx context.Context, client *http.Client, latestURL string) (githubRelease, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, latestURL, nil)
	if err != nil {
		return githubRelease{}, err
	}
	req.Header.Set("User-Agent", "clickup-cli/"+VersionString())

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

	const tagPath = "/Robben-Media/clickup-cli/releases/tag/"
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
		SyntheticAssets: true,
		Assets: []githubReleaseAsset{{
			Name:               "checksums.txt",
			BrowserDownloadURL: updateReleaseAssetURL(tag, "checksums.txt"),
		}},
	}, nil
}

func updateReleaseAssetURL(tag, assetName string) string {
	return "https://github.com/Robben-Media/clickup-cli/releases/download/" +
		url.PathEscape(tag) + "/" + url.PathEscape(assetName)
}

func fetchUpdateJSON(ctx context.Context, client *http.Client, endpoint string, dst any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "clickup-cli/"+VersionString())

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
	req.Header.Set("User-Agent", "clickup-cli/"+VersionString())
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
	version := strings.TrimPrefix(strings.TrimSpace(tag), "v")
	if version == "" {
		return ""
	}
	extension := ".tar.gz"
	if goos == "windows" {
		extension = ".zip"
	}
	return fmt.Sprintf("clickup-cli_%s_%s_%s%s", version, goos, goarch, extension)
}

func updateAvailable(current, latest string) (bool, bool) {
	comparison, ok := compareReleaseVersions(current, latest)
	if !ok {
		return false, false
	}
	return comparison < 0, true
}

type releaseVersion struct {
	parts      []int
	prerelease []string
}

func compareReleaseVersions(current, latest string) (int, bool) {
	currentVersion, currentOK := parseReleaseVersion(current)
	latestVersion, latestOK := parseReleaseVersion(latest)
	if !currentOK || !latestOK {
		return 0, false
	}

	maxLen := len(currentVersion.parts)
	if len(latestVersion.parts) > maxLen {
		maxLen = len(latestVersion.parts)
	}
	for i := 0; i < maxLen; i++ {
		var currentPart, latestPart int
		if i < len(currentVersion.parts) {
			currentPart = currentVersion.parts[i]
		}
		if i < len(latestVersion.parts) {
			latestPart = latestVersion.parts[i]
		}
		if currentPart < latestPart {
			return -1, true
		}
		if currentPart > latestPart {
			return 1, true
		}
	}

	return comparePrerelease(currentVersion.prerelease, latestVersion.prerelease), true
}

func parseReleaseVersion(value string) (releaseVersion, bool) {
	value = strings.TrimPrefix(strings.TrimSpace(value), "v")
	if value == "" || strings.HasPrefix(value, "dev") {
		return releaseVersion{}, false
	}

	value, _, _ = strings.Cut(value, "+")
	core, prerelease, hasPrerelease := strings.Cut(value, "-")
	if hasPrerelease && isGitDescribePrerelease(prerelease) {
		return releaseVersion{}, false
	}

	fields := strings.Split(core, ".")
	parts := make([]int, 0, len(fields))
	for _, field := range fields {
		if field == "" {
			return releaseVersion{}, false
		}
		part, err := strconv.Atoi(field)
		if err != nil || part < 0 {
			return releaseVersion{}, false
		}
		parts = append(parts, part)
	}

	var prereleaseParts []string
	if hasPrerelease {
		prereleaseParts = strings.Split(prerelease, ".")
		for _, part := range prereleaseParts {
			if part == "" {
				return releaseVersion{}, false
			}
		}
	}
	return releaseVersion{parts: parts, prerelease: prereleaseParts}, true
}

func comparePrerelease(current, latest []string) int {
	switch {
	case len(current) == 0 && len(latest) == 0:
		return 0
	case len(current) == 0:
		return 1
	case len(latest) == 0:
		return -1
	}

	maxLen := len(current)
	if len(latest) > maxLen {
		maxLen = len(latest)
	}
	for i := 0; i < maxLen; i++ {
		if i >= len(current) {
			return -1
		}
		if i >= len(latest) {
			return 1
		}

		currentNumber, currentErr := strconv.Atoi(current[i])
		latestNumber, latestErr := strconv.Atoi(latest[i])
		switch {
		case currentErr == nil && latestErr == nil:
			if currentNumber < latestNumber {
				return -1
			}
			if currentNumber > latestNumber {
				return 1
			}
		case currentErr == nil:
			return -1
		case latestErr == nil:
			return 1
		default:
			if current[i] < latest[i] {
				return -1
			}
			if current[i] > latest[i] {
				return 1
			}
		}
	}
	return 0
}

func isGitDescribePrerelease(value string) bool {
	parts := strings.Split(value, "-")
	if len(parts) < 2 {
		return false
	}
	if _, err := strconv.Atoi(parts[0]); err != nil || !strings.HasPrefix(parts[1], "g") {
		return false
	}
	hash := strings.TrimPrefix(parts[1], "g")
	if hash == "" {
		return false
	}
	for _, character := range hash {
		if !strings.ContainsRune("0123456789abcdefABCDEF", character) {
			return false
		}
	}
	return len(parts) == 2 || (len(parts) == 3 && parts[2] == "dirty")
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
