package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/builtbyrobben/clickup-cli/internal/outfmt"
)

func TestUpdateStatusReportsAvailableReleaseAsJSON(t *testing.T) {
	assetName := platformAssetName("v1.2.0", runtime.GOOS, runtime.GOARCH)
	checksum := strings.Repeat("a", 64)

	var serverURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/latest":
			_, _ = fmt.Fprintf(w, `{
				"tag_name":"v1.2.0",
				"html_url":"https://github.com/Robben-Media/clickup-cli/releases/tag/v1.2.0",
				"assets":[
					{"name":%q,"browser_download_url":%q},
					{"name":"checksums.txt","browser_download_url":%q}
				]
			}`, assetName, serverURL+"/download/"+assetName, serverURL+"/checksums.txt")
		case "/checksums.txt":
			_, _ = fmt.Fprintf(w, "%s  %s\n", checksum, assetName)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	serverURL = server.URL

	restore := setUpdateTestState(t, server.Client(), server.URL+"/latest", server.URL+"/releases/latest")
	defer restore()

	ctx := outfmt.WithMode(context.Background(), outfmt.Mode{JSON: true})
	stdout := captureStdout(t, func() error {
		return (&UpdateStatusCmd{Timeout: time.Second}).Run(ctx)
	})

	var report updateStatusReport
	if err := json.Unmarshal([]byte(stdout), &report); err != nil {
		t.Fatalf("decode JSON output: %v\noutput=%s", err, stdout)
	}
	if report.CurrentVersion != "v1.1.0" {
		t.Fatalf("current_version = %q, want v1.1.0", report.CurrentVersion)
	}
	if report.LatestVersion != "v1.2.0" {
		t.Fatalf("latest_version = %q, want v1.2.0", report.LatestVersion)
	}
	if !report.UpdateAvailable {
		t.Fatal("update_available = false, want true")
	}
	if report.PlatformAsset != assetName {
		t.Fatalf("platform_asset = %q, want %q", report.PlatformAsset, assetName)
	}
	if report.PlatformAssetSHA256 != checksum {
		t.Fatalf("platform_asset_sha256 = %q, want %q", report.PlatformAssetSHA256, checksum)
	}
}

func setUpdateTestState(t *testing.T, client *http.Client, latestURL, webURL string) func() {
	t.Helper()

	oldClient := updateHTTPClient
	oldLatestURL := updateLatestReleaseURL
	oldWebURL := updateLatestWebURL
	oldVersion := version
	oldCommit := commit
	oldDate := date

	updateHTTPClient = client
	updateLatestReleaseURL = latestURL
	updateLatestWebURL = webURL
	version = "v1.1.0"
	commit = "abc1234"
	date = "2026-08-10T00:00:00Z"

	return func() {
		updateHTTPClient = oldClient
		updateLatestReleaseURL = oldLatestURL
		updateLatestWebURL = oldWebURL
		version = oldVersion
		commit = oldCommit
		date = oldDate
	}
}

func TestUpdateStatusCheckAlias(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/latest" {
			http.NotFound(w, r)
			return
		}
		_, _ = fmt.Fprint(w, `{"tag_name":"v1.1.0","assets":[]}`)
	}))
	defer server.Close()

	restore := setUpdateTestState(t, server.Client(), server.URL+"/latest", server.URL+"/releases/latest")
	defer restore()

	stdout := captureStdout(t, func() error {
		return Execute([]string{"--json", "update", "check"})
	})
	if !strings.Contains(stdout, `"update_available": false`) {
		t.Fatalf("check alias output = %q", stdout)
	}
}

func TestUpdateStatusSupportsPlainAndHumanOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, `{"tag_name":"v1.1.0","html_url":"https://github.com/Robben-Media/clickup-cli/releases/tag/v1.1.0","assets":[]}`)
	}))
	defer server.Close()

	restore := setUpdateTestState(t, server.Client(), server.URL, server.URL)
	defer restore()

	plainCtx := outfmt.WithMode(context.Background(), outfmt.Mode{Plain: true})
	plain := captureStdout(t, func() error {
		return (&UpdateStatusCmd{Timeout: time.Second}).Run(plainCtx)
	})
	if !strings.HasPrefix(plain, "CURRENT_VERSION\t") || !strings.Contains(plain, "v1.1.0") {
		t.Fatalf("plain output = %q", plain)
	}

	human := captureStdout(t, func() error {
		return (&UpdateStatusCmd{Timeout: time.Second}).Run(context.Background())
	})
	if !strings.Contains(human, "clickup-cli is up to date (v1.1.0)") {
		t.Fatalf("human output = %q", human)
	}
}

func TestUpdateStatusWarnsForMalformedChecksumAndMissingAsset(t *testing.T) {
	assetName := platformAssetName("v1.2.0", runtime.GOOS, runtime.GOARCH)
	var serverURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/latest":
			_, _ = fmt.Fprintf(w, `{"tag_name":"v1.2.0","assets":[{"name":"checksums.txt","browser_download_url":%q}]}`, serverURL+"/checksums.txt")
		case "/checksums.txt":
			_, _ = fmt.Fprintf(w, "not-a-checksum  %s\n", assetName)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	serverURL = server.URL

	restore := setUpdateTestState(t, server.Client(), server.URL+"/latest", server.URL)
	defer restore()

	report, err := buildUpdateStatusReport(context.Background(), time.Second)
	if err != nil {
		t.Fatalf("build report: %v", err)
	}
	warnings := strings.Join(report.Warnings, "\n")
	if !strings.Contains(warnings, "no release asset found") {
		t.Fatalf("warnings = %q, want missing asset warning", warnings)
	}
	if !strings.Contains(warnings, "invalid checksum") {
		t.Fatalf("warnings = %q, want invalid checksum warning", warnings)
	}
	if !report.ChecksumAvailable {
		t.Fatal("checksum_available = false although the release lists checksums.txt")
	}
}

func TestUpdateStatusWarnsWhenChecksumsAssetIsMissing(t *testing.T) {
	assetName := platformAssetName("v1.2.0", runtime.GOOS, runtime.GOARCH)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprintf(w, `{"tag_name":"v1.2.0","assets":[{"name":%q,"browser_download_url":"https://example.test/archive"}]}`, assetName)
	}))
	defer server.Close()

	restore := setUpdateTestState(t, server.Client(), server.URL, server.URL)
	defer restore()

	report, err := buildUpdateStatusReport(context.Background(), time.Second)
	if err != nil {
		t.Fatalf("build report: %v", err)
	}
	if report.ChecksumAvailable {
		t.Fatal("checksum_available = true without checksums.txt")
	}
	if !strings.Contains(strings.Join(report.Warnings, "\n"), "checksums.txt not found") {
		t.Fatalf("warnings = %q", report.Warnings)
	}
}

func TestFetchLatestGitHubReleaseFallsBackToSafeRedirect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/latest":
			http.Error(w, "rate limited", http.StatusForbidden)
		case "/releases/latest":
			http.Redirect(w, r, "https://github.com/Robben-Media/clickup-cli/releases/tag/v1.2.0", http.StatusFound)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	oldWebURL := updateLatestWebURL
	updateLatestWebURL = server.URL + "/releases/latest"
	defer func() { updateLatestWebURL = oldWebURL }()

	release, err := fetchLatestGitHubRelease(context.Background(), server.Client(), server.URL+"/api/latest")
	if err != nil {
		t.Fatalf("fetch latest release: %v", err)
	}
	if release.TagName != "v1.2.0" || !release.SyntheticAssets {
		t.Fatalf("release = %#v", release)
	}
}

func TestFetchLatestGitHubReleaseRejectsUnsafeRedirect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://example.com/Robben-Media/clickup-cli/releases/tag/v1.2.0", http.StatusFound)
	}))
	defer server.Close()

	_, err := fetchLatestGitHubReleaseRedirect(context.Background(), server.Client(), server.URL)
	if err == nil || !strings.Contains(err.Error(), "unexpected github redirect host") {
		t.Fatalf("error = %v, want unsafe redirect rejection", err)
	}
}

func TestUpdateStatusHonorsTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(50 * time.Millisecond)
		_, _ = fmt.Fprint(w, `{"tag_name":"v1.2.0"}`)
	}))
	defer server.Close()

	client := updateClientForTest(server.Client(), 5*time.Millisecond)
	var release githubRelease
	if err := fetchUpdateJSON(context.Background(), client, server.URL, &release); err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestUpdateFallbackDoesNotClaimUnavailableAssets(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		switch request.URL.String() {
		case "https://api.example/latest":
			return testHTTPResponse(http.StatusForbidden, "rate limited", ""), nil
		case "https://web.example/latest":
			return testHTTPResponse(http.StatusFound, "", "https://github.com/Robben-Media/clickup-cli/releases/tag/v1.2.0"), nil
		case "https://github.com/Robben-Media/clickup-cli/releases/download/v1.2.0/checksums.txt":
			return testHTTPResponse(http.StatusNotFound, "not found", ""), nil
		default:
			return nil, fmt.Errorf("unexpected HTTP request: %s", request.URL)
		}
	})}
	restore := setUpdateTestState(t, client, "https://api.example/latest", "https://web.example/latest")
	defer restore()

	report, err := buildUpdateStatusReport(context.Background(), time.Second)
	if err != nil {
		t.Fatalf("build report: %v", err)
	}
	if report.ChecksumAvailable {
		t.Fatal("checksum_available = true after checksums.txt returned 404")
	}
	if report.PlatformAssetURL != "" {
		t.Fatalf("platform_asset_url = %q without verified asset", report.PlatformAssetURL)
	}
	if !strings.Contains(strings.Join(report.Warnings, "\n"), "could not verify release asset") {
		t.Fatalf("warnings = %q", report.Warnings)
	}
}

func TestUpdateStatusHumanWritesWarningsToStderr(t *testing.T) {
	report := updateStatusReport{
		CurrentVersion:     "v1.1.0",
		LatestVersion:      "v1.1.0",
		VersionsComparable: true,
		Warnings:           []string{"checksums.txt not found"},
	}
	stderr := captureStderr(t, func() error {
		_ = captureStdout(t, func() error {
			return writeUpdateStatusHuman(report)
		})
		return nil
	})
	if !strings.Contains(stderr, "Warning: checksums.txt not found") {
		t.Fatalf("stderr = %q", stderr)
	}
}

func TestUpdateStatusHumanDoesNotCallUncomparableBuildCurrent(t *testing.T) {
	report := updateStatusReport{
		CurrentVersion:     "dev (no version)",
		LatestVersion:      "v1.2.0",
		LatestURL:          "https://github.com/Robben-Media/clickup-cli/releases/tag/v1.2.0",
		Platform:           runtime.GOOS + "/" + runtime.GOARCH,
		InstallMethod:      "source",
		VersionsComparable: false,
	}
	stdout := captureStdout(t, func() error {
		return writeUpdateStatusHuman(report)
	})
	if strings.Contains(stdout, "up to date") {
		t.Fatalf("human output makes unsupported claim: %q", stdout)
	}
	if !strings.Contains(stdout, "Could not determine update status") {
		t.Fatalf("human output = %q", stdout)
	}
}

func TestPlatformAssetNameMatchesGoReleaserArchives(t *testing.T) {
	if got := platformAssetName("v1.2.0", "linux", "amd64"); got != "clickup-cli_1.2.0_linux_amd64.tar.gz" {
		t.Fatalf("linux asset = %q", got)
	}
	if got := platformAssetName("v1.2.0", "windows", "arm64"); got != "clickup-cli_1.2.0_windows_arm64.zip" {
		t.Fatalf("windows asset = %q", got)
	}
}

func TestUpdateVersionComparison(t *testing.T) {
	tests := []struct {
		current string
		latest  string
		want    bool
		ok      bool
	}{
		{current: "v1.1.0", latest: "v1.2.0", want: true, ok: true},
		{current: "1.2.0", latest: "v1.2.0", want: false, ok: true},
		{current: "v1.3.0-dev", latest: "v1.2.0+build", want: false, ok: true},
		{current: "v1.2.0-rc.1", latest: "v1.2.0", want: true, ok: true},
		{current: "v1.2.0", latest: "v1.2.0-rc.1", want: false, ok: true},
		{current: "v0.1.0-1-gd2af795-dirty", latest: "v1.2.0", want: false, ok: false},
		{current: "dev (no version)", latest: "v1.2.0", want: false, ok: false},
	}
	for _, test := range tests {
		got, ok := updateAvailable(test.current, test.latest)
		if got != test.want || ok != test.ok {
			t.Fatalf("updateAvailable(%q, %q) = (%t, %t), want (%t, %t)", test.current, test.latest, got, ok, test.want, test.ok)
		}
	}
}

func updateClientForTest(client *http.Client, timeout time.Duration) *http.Client {
	clone := *client
	clone.Timeout = timeout
	return &clone
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (roundTrip roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return roundTrip(request)
}

func testHTTPResponse(status int, body, location string) *http.Response {
	header := make(http.Header)
	if location != "" {
		header.Set("Location", location)
	}
	return &http.Response{
		StatusCode: status,
		Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
		Header:     header,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func captureStderr(t *testing.T, run func() error) string {
	t.Helper()

	oldStderr := os.Stderr
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stderr pipe: %v", err)
	}
	os.Stderr = writer
	defer func() { os.Stderr = oldStderr }()

	runErr := run()
	if closeErr := writer.Close(); closeErr != nil {
		t.Fatalf("close stderr writer: %v", closeErr)
	}
	data, readErr := io.ReadAll(reader)
	if readErr != nil {
		t.Fatalf("read stderr: %v", readErr)
	}
	if closeErr := reader.Close(); closeErr != nil {
		t.Fatalf("close stderr reader: %v", closeErr)
	}
	if runErr != nil {
		t.Fatalf("run command: %v", runErr)
	}
	return string(data)
}

func captureStdout(t *testing.T, run func() error) string {
	t.Helper()

	oldStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %v", err)
	}
	os.Stdout = writer
	defer func() { os.Stdout = oldStdout }()

	runErr := run()
	if closeErr := writer.Close(); closeErr != nil {
		t.Fatalf("close stdout writer: %v", closeErr)
	}

	data, readErr := io.ReadAll(reader)
	if readErr != nil {
		t.Fatalf("read stdout: %v", readErr)
	}
	if closeErr := reader.Close(); closeErr != nil {
		t.Fatalf("close stdout reader: %v", closeErr)
	}
	if runErr != nil {
		t.Fatalf("run command: %v", runErr)
	}
	return string(data)
}
