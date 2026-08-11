package selfupdate

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var (
	errTestReplace  = errors.New("test replace failure")
	errTestRollback = errors.New("test rollback failure")
)

func TestApplyInstallsTarAndZipReleaseAssets(t *testing.T) {
	tests := []struct {
		name      string
		goos      string
		goarch    string
		separator string
		archive   func(*testing.T, []byte) []byte
	}{
		{name: "tar.gz", goos: "linux", goarch: "amd64", separator: "  ", archive: buildTarGz},
		{name: "zip", goos: windows, goarch: "arm64", separator: " *", archive: buildZip},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			binary := []byte("replacement executable")
			archive := test.archive(t, binary)
			assetName := AssetNameForPlatform("v2.0.0", test.goos, test.goarch)
			server, requests := releaseServer(t, releaseFixture{
				tag:          "v2.0.0",
				assetName:    assetName,
				archive:      archive,
				checksumBody: sha256Hex(archive) + test.separator + assetName + "\n",
				token:        "secret-token",
			})

			destination := filepath.Join(t.TempDir(), "clickup-cli")
			if err := os.WriteFile(destination, []byte("old executable"), 0o755); err != nil {
				t.Fatal(err)
			}

			result, err := Apply(context.Background(), ApplyOptions{
				Client:     &Client{HTTP: server.Client(), BaseURL: server.URL, Repo: "example/repo", Token: "secret-token"},
				CurrentVer: "v1.0.0",
				DestPath:   destination,
				GOOS:       test.goos,
				GOARCH:     test.goarch,
			})
			if err != nil {
				t.Fatalf("Apply: %v", err)
			}

			if !result.Update || !result.Applied || !result.Comparable || result.Asset != assetName || result.Latest != "2.0.0" {
				t.Fatalf("result = %#v", result)
			}

			installed, err := os.ReadFile(destination)
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(installed, binary) {
				t.Fatalf("installed = %q, want %q", installed, binary)
			}

			if requests.release != 1 || requests.checksum != 1 || requests.archive != 1 {
				t.Fatalf("requests = %#v", requests)
			}
		})
	}
}

func TestApplyAlreadyCurrentOrNewerNeverDownloads(t *testing.T) {
	for _, test := range []struct {
		name    string
		current string
		force   bool
	}{
		{name: "equal", current: "v2.0.0"},
		{name: "equal forced", current: "2.0.0+local", force: true},
		{name: "newer", current: "v3.0.0"},
		{name: "newer forced", current: "v3.0.0-rc.1", force: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			server, requests := releaseServer(t, releaseFixture{
				tag:       "v2.0.0",
				assetName: AssetNameForPlatform("v2.0.0", "linux", "amd64"),
				archive:   []byte("must not download"),
			})

			destination := filepath.Join(t.TempDir(), "clickup-cli")
			if err := os.WriteFile(destination, []byte("original"), 0o755); err != nil {
				t.Fatal(err)
			}

			result, err := Apply(context.Background(), ApplyOptions{
				Client:     &Client{HTTP: server.Client(), BaseURL: server.URL},
				CurrentVer: test.current,
				Force:      test.force,
				DestPath:   destination,
				GOOS:       "linux",
				GOARCH:     "amd64",
			})
			if err != nil {
				t.Fatalf("Apply: %v", err)
			}

			if result.Update || result.Applied {
				t.Fatalf("result = %#v, want no update", result)
			}

			if requests.release != 1 || requests.checksum != 0 || requests.archive != 0 {
				t.Fatalf("requests = %#v, downloads must not occur", requests)
			}

			installed, readErr := os.ReadFile(destination)
			if readErr != nil || string(installed) != "original" {
				t.Fatalf("installed = %q, err = %v", installed, readErr)
			}
		})
	}
}

func TestApplyRefusesUncomparableUnlessForced(t *testing.T) {
	for _, current := range []string{"", "dev", "v1.0.0-2-gabcdef", "v1.0.0-dirty"} {
		t.Run(fmt.Sprintf("refuse_%q", current), func(t *testing.T) {
			requests := 0
			server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { requests++ }))
			t.Cleanup(server.Close)

			_, err := Apply(context.Background(), ApplyOptions{
				Client:     &Client{HTTP: server.Client(), BaseURL: server.URL},
				CurrentVer: current,
				DestPath:   filepath.Join(t.TempDir(), "clickup-cli"),
			})
			if !errors.Is(err, errUncomparableVersion) {
				t.Fatalf("error = %v, want %v", err, errUncomparableVersion)
			}

			if requests != 0 {
				t.Fatalf("requests = %d, want none", requests)
			}
		})
	}

	for _, current := range []string{"dev", "v1.0.0-dirty"} {
		t.Run("force_"+current, func(t *testing.T) {
			binary := []byte("forced replacement")
			archive := buildTarGz(t, binary)
			assetName := AssetNameForPlatform("v2.0.0", "linux", "amd64")
			server, _ := releaseServer(t, releaseFixture{
				tag:          "v2.0.0",
				assetName:    assetName,
				archive:      archive,
				checksumBody: sha256Hex(archive) + "  " + assetName + "\n",
			})

			destination := filepath.Join(t.TempDir(), "clickup-cli")
			if err := os.WriteFile(destination, []byte("old"), 0o755); err != nil {
				t.Fatal(err)
			}

			result, err := Apply(context.Background(), ApplyOptions{
				Client:     &Client{HTTP: server.Client(), BaseURL: server.URL},
				CurrentVer: current,
				Force:      true,
				DestPath:   destination,
				GOOS:       "linux",
				GOARCH:     "amd64",
			})
			if err != nil {
				t.Fatalf("Apply: %v", err)
			}

			if result.Comparable || !result.Update {
				t.Fatalf("result = %#v", result)
			}
		})
	}
}

func TestCheckUsesSemanticVersionOrdering(t *testing.T) {
	for _, test := range []struct {
		name       string
		current    string
		latest     string
		update     bool
		comparable bool
	}{
		{name: "leading v", current: "v1.2.3", latest: "1.3.0", update: true, comparable: true},
		{name: "prerelease before release", current: "1.2.3-rc.2", latest: "v1.2.3", update: true, comparable: true},
		{name: "numeric prerelease ordering", current: "1.2.3-rc.10", latest: "1.2.3-rc.2", comparable: true},
		{name: "build metadata ignored", current: "1.2.3+local", latest: "v1.2.3+release", comparable: true},
		{name: "dev uncomparable", current: "dev", latest: "v1.2.3"},
		{name: "git describe uncomparable", current: "v1.2.3-4-gabcdef", latest: "v1.2.4"},
		{name: "dirty uncomparable", current: "v1.2.3+dirty", latest: "v1.2.4"},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/repos/Robben-Media/clickup-cli/releases/latest" {
					http.NotFound(w, r)
					return
				}
				_ = json.NewEncoder(w).Encode(Release{TagName: test.latest})
			}))
			t.Cleanup(server.Close)

			result, err := Check(context.Background(), &Client{HTTP: server.Client(), BaseURL: server.URL}, test.current)
			if err != nil {
				t.Fatalf("Check: %v", err)
			}

			if result.Update != test.update || result.Comparable != test.comparable {
				t.Fatalf("result = %#v, want update=%t comparable=%t", result, test.update, test.comparable)
			}
		})
	}
}

func TestApplyRejectsInvalidChecksums(t *testing.T) {
	binary := []byte("replacement")
	archive := buildTarGz(t, binary)
	assetName := AssetNameForPlatform("v2.0.0", "linux", "amd64")

	validDigest := sha256Hex(archive)
	for _, test := range []struct {
		name             string
		includeChecksums bool
		manifest         string
		wantError        error
	}{
		{name: "missing checksums asset", wantError: errChecksumRequired},
		{name: "missing matching checksum", includeChecksums: true, manifest: validDigest + "  other.tar.gz\n", wantError: errChecksumInvalid},
		{name: "malformed line", includeChecksums: true, manifest: "malformed\n" + validDigest + "  " + assetName + "\n", wantError: errChecksumInvalid},
		{name: "invalid digest", includeChecksums: true, manifest: strings.Repeat("z", 64) + "  " + assetName + "\n", wantError: errChecksumInvalid},
		{name: "duplicate checksum", includeChecksums: true, manifest: validDigest + "  " + assetName + "\n" + validDigest + " *" + assetName + "\n", wantError: errChecksumInvalid},
		{name: "digest mismatch", includeChecksums: true, manifest: strings.Repeat("0", 64) + "  " + assetName + "\n", wantError: errChecksumMismatch},
	} {
		t.Run(test.name, func(t *testing.T) {
			server, _ := releaseServer(t, releaseFixture{
				tag:              "v2.0.0",
				assetName:        assetName,
				archive:          archive,
				includeChecksums: test.includeChecksums,
				checksumBody:     test.manifest,
			})

			destination := filepath.Join(t.TempDir(), "clickup-cli")
			if err := os.WriteFile(destination, []byte("original"), 0o755); err != nil {
				t.Fatal(err)
			}

			_, err := Apply(context.Background(), ApplyOptions{
				Client:     &Client{HTTP: server.Client(), BaseURL: server.URL},
				CurrentVer: "v1.0.0",
				DestPath:   destination,
				GOOS:       "linux",
				GOARCH:     "amd64",
			})
			if !errors.Is(err, test.wantError) {
				t.Fatalf("error = %v, want %v", err, test.wantError)
			}

			installed, readErr := os.ReadFile(destination)
			if readErr != nil || string(installed) != "original" {
				t.Fatalf("installed = %q, err = %v", installed, readErr)
			}
		})
	}
}

func TestApplyRejectsMissingBinaryAndMalformedArchive(t *testing.T) {
	truncatedArchive := buildTarGz(t, []byte("apparently valid binary"))
	truncatedArchive = truncatedArchive[:len(truncatedArchive)-4]

	for _, test := range []struct {
		name      string
		archive   []byte
		wantError error
	}{
		{name: "missing binary", archive: buildTarGzNamed(t, "other", []byte("not executable")), wantError: errArchiveNoBinary},
		{name: "malformed archive", archive: []byte("not a gzip archive")},
		{name: "truncated after binary", archive: truncatedArchive},
	} {
		t.Run(test.name, func(t *testing.T) {
			assetName := AssetNameForPlatform("v2.0.0", "linux", "amd64")
			server, _ := releaseServer(t, releaseFixture{
				tag:          "v2.0.0",
				assetName:    assetName,
				archive:      test.archive,
				checksumBody: sha256Hex(test.archive) + "  " + assetName + "\n",
			})

			destination := filepath.Join(t.TempDir(), "clickup-cli")
			if err := os.WriteFile(destination, []byte("original"), 0o755); err != nil {
				t.Fatal(err)
			}

			_, err := Apply(context.Background(), ApplyOptions{
				Client:     &Client{HTTP: server.Client(), BaseURL: server.URL},
				CurrentVer: "v1.0.0",
				DestPath:   destination,
				GOOS:       "linux",
				GOARCH:     "amd64",
			})
			if err == nil {
				t.Fatal("Apply succeeded")
			}

			if test.wantError != nil && !errors.Is(err, test.wantError) {
				t.Fatalf("error = %v, want %v", err, test.wantError)
			}
		})
	}
}

func TestApplyUsesOnlyListedReleaseAssetURLs(t *testing.T) {
	assetName := AssetNameForPlatform("v2.0.0", "linux", "amd64")
	downloads := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/releases/latest") {
			_ = json.NewEncoder(w).Encode(Release{TagName: "v2.0.0", Assets: []Asset{{Name: "different.tar.gz", BrowserDownloadURL: "https://example.invalid/asset"}}})
			return
		}

		downloads++

		http.NotFound(w, r)
	}))
	t.Cleanup(server.Close)

	_, err := Apply(context.Background(), ApplyOptions{
		Client:     &Client{HTTP: server.Client(), BaseURL: server.URL},
		CurrentVer: "v1.0.0",
		DestPath:   filepath.Join(t.TempDir(), "clickup-cli"),
		GOOS:       "linux",
		GOARCH:     "amd64",
	})
	if !errors.Is(err, errNoReleaseAsset) {
		t.Fatalf("error = %v, want missing %s", err, assetName)
	}

	if downloads != 0 {
		t.Fatalf("downloads = %d, want none", downloads)
	}
}

func TestClientUsesSuppliedTimeoutAndContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(50 * time.Millisecond)
		_ = json.NewEncoder(w).Encode(Release{TagName: "v2.0.0"})
	}))
	t.Cleanup(server.Close)

	client := server.Client()
	client.Timeout = 5 * time.Millisecond

	_, err := (&Client{HTTP: client, BaseURL: server.URL}).LatestRelease(context.Background())
	if err == nil {
		t.Fatal("LatestRelease ignored supplied client timeout")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = (&Client{HTTP: server.Client(), BaseURL: server.URL}).LatestRelease(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context canceled", err)
	}
}

func TestUnixRenameFailurePreservesExecutableAndUsesSameDirectoryTemp(t *testing.T) {
	originalRename := renameFile

	t.Cleanup(func() { renameFile = originalRename })

	destination := filepath.Join(t.TempDir(), "clickup-cli")
	if err := os.WriteFile(destination, []byte("original"), 0o755); err != nil {
		t.Fatal(err)
	}
	var temporaryName string
	renameFile = func(oldPath, newPath string) error {
		temporaryName = oldPath

		if newPath != destination {
			t.Fatalf("rename destination = %q, want %q", newPath, destination)
		}

		return errTestReplace
	}

	err := replaceExecutable(destination, []byte("replacement"), "linux")
	if !errors.Is(err, errTestReplace) {
		t.Fatalf("error = %v, want %v", err, errTestReplace)
	}

	if filepath.Dir(temporaryName) != filepath.Dir(destination) {
		t.Fatalf("temporary path %q is not beside destination %q", temporaryName, destination)
	}

	installed, readErr := os.ReadFile(destination)
	if readErr != nil || string(installed) != "original" {
		t.Fatalf("installed = %q, err = %v", installed, readErr)
	}
}

func TestWindowsReplaceRollsBackFailedReplacement(t *testing.T) {
	originalRename := renameFile

	t.Cleanup(func() { renameFile = originalRename })

	destination := filepath.Join(t.TempDir(), "clickup-cli.exe")
	if err := os.WriteFile(destination, []byte("original"), 0o755); err != nil {
		t.Fatal(err)
	}
	calls := 0
	renameFile = func(oldPath, newPath string) error {
		calls++
		if calls == 2 {
			return errTestReplace
		}

		return os.Rename(oldPath, newPath)
	}

	err := replaceExecutable(destination, []byte("replacement"), windows)
	if !errors.Is(err, errTestReplace) {
		t.Fatalf("error = %v, want %v", err, errTestReplace)
	}

	installed, readErr := os.ReadFile(destination)
	if readErr != nil || string(installed) != "original" {
		t.Fatalf("installed = %q, err = %v", installed, readErr)
	}
}

func TestWindowsReplaceSurfacesRollbackFailure(t *testing.T) {
	originalRename := renameFile

	t.Cleanup(func() { renameFile = originalRename })

	destination := filepath.Join(t.TempDir(), "clickup-cli.exe")
	if err := os.WriteFile(destination, []byte("original"), 0o755); err != nil {
		t.Fatal(err)
	}
	calls := 0
	renameFile = func(oldPath, newPath string) error {
		calls++
		switch calls {
		case 1:
			return os.Rename(oldPath, newPath)
		case 2:
			return errTestReplace
		case 3:
			return errTestRollback
		default:
			t.Fatalf("unexpected rename %q -> %q", oldPath, newPath)
			return nil
		}
	}

	err := replaceExecutable(destination, []byte("replacement"), windows)
	if !errors.Is(err, errTestReplace) || !errors.Is(err, errTestRollback) {
		t.Fatalf("error = %v, want replace and rollback failures", err)
	}

	if calls != 3 {
		t.Fatalf("rename calls = %d, want 3", calls)
	}
}

type releaseFixture struct {
	tag              string
	assetName        string
	archive          []byte
	checksumBody     string
	includeChecksums bool
	token            string
}

type requestCounts struct {
	release  int
	checksum int
	archive  int
}

func releaseServer(t *testing.T, fixture releaseFixture) (*httptest.Server, *requestCounts) {
	t.Helper()

	if fixture.checksumBody != "" {
		fixture.includeChecksums = true
	}
	counts := &requestCounts{}
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if fixture.token != "" && r.Header.Get("Authorization") != "Bearer "+fixture.token {
			t.Errorf("Authorization = %q", r.Header.Get("Authorization"))
			http.Error(w, "unauthorized", http.StatusUnauthorized)

			return
		}

		switch r.URL.Path {
		case "/repos/example/repo/releases/latest", "/repos/Robben-Media/clickup-cli/releases/latest":
			counts.release++

			assets := []Asset{{Name: fixture.assetName, BrowserDownloadURL: server.URL + "/asset"}}
			if fixture.includeChecksums {
				assets = append(assets, Asset{Name: "checksums.txt", BrowserDownloadURL: server.URL + "/checksums"})
			}
			_ = json.NewEncoder(w).Encode(Release{TagName: fixture.tag, Assets: assets})
		case "/checksums":
			counts.checksum++
			_, _ = io.WriteString(w, fixture.checksumBody)
		case "/asset":
			counts.archive++
			_, _ = w.Write(fixture.archive)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	return server, counts
}

func buildTarGz(t *testing.T, binary []byte) []byte {
	t.Helper()
	return buildTarGzNamed(t, "clickup-cli", binary)
}

func buildTarGzNamed(t *testing.T, name string, contents []byte) []byte {
	t.Helper()
	var output bytes.Buffer
	gzipWriter := gzip.NewWriter(&output)

	tarWriter := tar.NewWriter(gzipWriter)
	if err := tarWriter.WriteHeader(&tar.Header{Name: name, Mode: 0o755, Size: int64(len(contents)), Typeflag: tar.TypeReg}); err != nil {
		t.Fatal(err)
	}

	if _, err := tarWriter.Write(contents); err != nil {
		t.Fatal(err)
	}

	if err := tarWriter.Close(); err != nil {
		t.Fatal(err)
	}

	if err := gzipWriter.Close(); err != nil {
		t.Fatal(err)
	}

	return output.Bytes()
}

func buildZip(t *testing.T, binary []byte) []byte {
	t.Helper()
	var output bytes.Buffer
	zipWriter := zip.NewWriter(&output)
	header := &zip.FileHeader{Name: "clickup-cli.exe", Method: zip.Deflate}
	header.SetMode(0o755)

	entry, err := zipWriter.CreateHeader(header)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := entry.Write(binary); err != nil {
		t.Fatal(err)
	}

	if err := zipWriter.Close(); err != nil {
		t.Fatal(err)
	}

	return output.Bytes()
}

func sha256Hex(contents []byte) string {
	digest := sha256.Sum256(contents)
	return hex.EncodeToString(digest[:])
}
