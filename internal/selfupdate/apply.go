package selfupdate

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

const maxArchiveRemainder = 1 << 20

var (
	errUncomparableVersion = errors.New("current version is not comparable")
	errLatestUncomparable  = errors.New("latest release version is not comparable")
	errChecksumRequired    = errors.New("checksums.txt is required")
	errChecksumInvalid     = errors.New("invalid checksums.txt")
	errChecksumMismatch    = errors.New("checksum mismatch")
	errArchiveNoBinary     = errors.New("archive missing clickup-cli binary")
	errArchiveTooLarge     = errors.New("archive trailing data is too large")
	renameFile             = os.Rename
	removeFile             = os.Remove
)

// ApplyOptions controls release selection and executable replacement.
type ApplyOptions struct {
	Client     *Client
	CurrentVer string
	Force      bool
	DestPath   string
	GOOS       string
	GOARCH     string
}

// CheckResult describes the latest release for the current version.
type CheckResult struct {
	Current    string
	Latest     string
	Update     bool
	Applied    bool
	Comparable bool
	Asset      string
}

// Check reports whether the latest release is newer than current.
func Check(ctx context.Context, client *Client, current string) (CheckResult, error) {
	return checkPlatform(ctx, client, current, runtime.GOOS, runtime.GOARCH)
}

func checkPlatform(ctx context.Context, client *Client, current, goos, goarch string) (CheckResult, error) {
	client = defaultClient(client)

	release, err := client.LatestRelease(ctx)
	if err != nil {
		return CheckResult{}, err
	}

	result := CheckResult{
		Current: current,
		Latest:  NormalizeVersion(release.TagName),
		Asset:   AssetNameForPlatform(release.TagName, goos, goarch),
	}
	comparison, versionsComparable := compareVersions(current, release.TagName)
	result.Comparable = versionsComparable
	result.Update = versionsComparable && comparison < 0

	return result, nil
}

// Apply downloads, verifies, extracts, and installs the latest release.
func Apply(ctx context.Context, opts ApplyOptions) (CheckResult, error) {
	goos, goarch := opts.GOOS, opts.GOARCH
	if goos == "" {
		goos = runtime.GOOS
	}

	if goarch == "" {
		goarch = runtime.GOARCH
	}

	if _, ok := parseVersion(opts.CurrentVer); !ok && !opts.Force {
		return CheckResult{Current: opts.CurrentVer}, fmt.Errorf("%w: %q", errUncomparableVersion, opts.CurrentVer)
	}

	client := defaultClient(opts.Client)

	release, err := client.LatestRelease(ctx)
	if err != nil {
		return CheckResult{}, err
	}

	result := CheckResult{
		Current: opts.CurrentVer,
		Latest:  NormalizeVersion(release.TagName),
		Asset:   AssetNameForPlatform(release.TagName, goos, goarch),
	}

	latestVersion, latestOK := parseVersion(release.TagName)
	if !latestOK {
		return result, fmt.Errorf("%w: %q", errLatestUncomparable, release.TagName)
	}

	currentVersion, currentOK := parseVersion(opts.CurrentVer)
	if currentOK {
		result.Comparable = true
		comparison := compareParsedVersions(currentVersion, latestVersion)

		result.Update = comparison < 0
		if comparison >= 0 {
			return result, nil
		}
	} else {
		result.Update = true
	}

	archiveAsset, checksumAsset, err := findAssets(release, goos, goarch)
	if err != nil {
		return result, err
	}

	if checksumAsset.Name == "" || strings.TrimSpace(checksumAsset.BrowserDownloadURL) == "" {
		return result, fmt.Errorf("%w: release %s has no checksums.txt asset", errChecksumRequired, release.TagName)
	}

	var manifest bytes.Buffer

	err = client.Download(ctx, checksumAsset.BrowserDownloadURL, &manifest)
	if err != nil {
		return result, fmt.Errorf("download checksums.txt: %w", err)
	}

	expectedDigest, err := checksumFor(manifest.String(), archiveAsset.Name)
	if err != nil {
		return result, err
	}

	var archive bytes.Buffer

	err = client.Download(ctx, archiveAsset.BrowserDownloadURL, &archive)
	if err != nil {
		return result, fmt.Errorf("download release asset: %w", err)
	}

	digest := sha256.Sum256(archive.Bytes())
	if !strings.EqualFold(expectedDigest, hex.EncodeToString(digest[:])) {
		return result, fmt.Errorf("%w for %s", errChecksumMismatch, archiveAsset.Name)
	}

	binary, err := extractBinary(archive.Bytes(), archiveAsset.Name)
	if err != nil {
		return result, err
	}

	destination, err := destinationPath(opts.DestPath)
	if err != nil {
		return result, err
	}

	if err := replaceExecutable(destination, binary, goos); err != nil {
		return result, err
	}

	result.Applied = true

	return result, nil
}

func defaultClient(client *Client) *Client {
	if client == nil {
		return &Client{}
	}

	return client
}

func destinationPath(destination string) (string, error) {
	if destination != "" {
		return destination, nil
	}

	executable, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("resolve executable: %w", err)
	}

	executable, err = filepath.EvalSymlinks(executable)
	if err != nil {
		return "", fmt.Errorf("resolve executable path: %w", err)
	}

	return executable, nil
}

type semanticVersion struct {
	major      string
	minor      string
	patch      string
	prerelease []string
}

func compareVersions(left, right string) (int, bool) {
	leftVersion, leftOK := parseVersion(left)

	rightVersion, rightOK := parseVersion(right)
	if !leftOK || !rightOK {
		return 0, false
	}

	return compareParsedVersions(leftVersion, rightVersion), true
}

func parseVersion(value string) (semanticVersion, bool) {
	value = strings.TrimSpace(value)

	lower := strings.ToLower(value)
	if value == "" || strings.HasPrefix(lower, "dev") || strings.Contains(lower, "dirty") {
		return semanticVersion{}, false
	}
	value = strings.TrimPrefix(value, "v")
	value, _, _ = strings.Cut(value, "+")

	core, prerelease, hasPrerelease := strings.Cut(value, "-")
	if isGitDescribe(value) {
		return semanticVersion{}, false
	}

	parts := strings.Split(core, ".")
	if len(parts) != 3 {
		return semanticVersion{}, false
	}

	for _, part := range parts {
		if !validNumericIdentifier(part) {
			return semanticVersion{}, false
		}
	}

	version := semanticVersion{major: parts[0], minor: parts[1], patch: parts[2]}
	if !hasPrerelease {
		return version, true
	}

	version.prerelease = strings.Split(prerelease, ".")
	for _, identifier := range version.prerelease {
		if !validPrereleaseIdentifier(identifier) {
			return semanticVersion{}, false
		}
	}

	return version, true
}

func validNumericIdentifier(value string) bool {
	if value == "" || (len(value) > 1 && value[0] == '0') {
		return false
	}

	for _, character := range value {
		if character < '0' || character > '9' {
			return false
		}
	}

	return true
}

func validPrereleaseIdentifier(value string) bool {
	if value == "" {
		return false
	}
	numeric := true

	for _, character := range value {
		if (character < '0' || character > '9') && (character < 'A' || character > 'Z') &&
			(character < 'a' || character > 'z') && character != '-' {
			return false
		}

		if character < '0' || character > '9' {
			numeric = false
		}
	}

	return !numeric || len(value) == 1 || value[0] != '0'
}

func isGitDescribe(value string) bool {
	parts := strings.Split(value, "-")
	if len(parts) < 3 {
		return false
	}
	count := parts[len(parts)-2]

	hash := parts[len(parts)-1]
	if _, err := strconv.ParseUint(count, 10, 64); err != nil || len(hash) < 2 || hash[0] != 'g' {
		return false
	}

	for _, character := range hash[1:] {
		if !strings.ContainsRune("0123456789abcdefABCDEF", character) {
			return false
		}
	}

	return true
}

func compareParsedVersions(left, right semanticVersion) int {
	for _, pair := range [][2]string{{left.major, right.major}, {left.minor, right.minor}, {left.patch, right.patch}} {
		if comparison := compareNumericIdentifiers(pair[0], pair[1]); comparison != 0 {
			return comparison
		}
	}

	return comparePrerelease(left.prerelease, right.prerelease)
}

func compareNumericIdentifiers(left, right string) int {
	if len(left) < len(right) {
		return -1
	}

	if len(left) > len(right) {
		return 1
	}

	return strings.Compare(left, right)
}

func comparePrerelease(left, right []string) int {
	if len(left) == 0 && len(right) == 0 {
		return 0
	}

	if len(left) == 0 {
		return 1
	}

	if len(right) == 0 {
		return -1
	}

	for index := 0; index < len(left) && index < len(right); index++ {
		leftNumeric := validNumericIdentifier(left[index])

		rightNumeric := validNumericIdentifier(right[index])
		switch {
		case leftNumeric && rightNumeric:
			if comparison := compareNumericIdentifiers(left[index], right[index]); comparison != 0 {
				return comparison
			}
		case leftNumeric:
			return -1
		case rightNumeric:
			return 1
		default:
			if comparison := strings.Compare(left[index], right[index]); comparison != 0 {
				return comparison
			}
		}
	}

	if len(left) < len(right) {
		return -1
	}

	if len(left) > len(right) {
		return 1
	}

	return 0
}

func checksumFor(manifest, assetName string) (string, error) {
	lines := strings.Split(manifest, "\n")
	matches := 0
	matchingDigest := ""

	for index, line := range lines {
		if line == "" && index == len(lines)-1 {
			continue
		}

		if len(line) < sha256.Size*2+3 {
			return "", fmt.Errorf("%w: malformed line %d", errChecksumInvalid, index+1)
		}
		digest := line[:sha256.Size*2]

		separator := line[sha256.Size*2 : sha256.Size*2+2]
		if separator != "  " && separator != " *" {
			return "", fmt.Errorf("%w: malformed line %d", errChecksumInvalid, index+1)
		}

		decoded, err := hex.DecodeString(digest)
		if err != nil || len(decoded) != sha256.Size {
			return "", fmt.Errorf("%w: malformed digest on line %d", errChecksumInvalid, index+1)
		}

		filename := line[sha256.Size*2+2:]
		if filename == "" || strings.ContainsAny(filename, "\r\t") {
			return "", fmt.Errorf("%w: malformed filename on line %d", errChecksumInvalid, index+1)
		}

		if filename == assetName {
			matches++
			matchingDigest = digest
		}
	}

	if matches != 1 {
		return "", fmt.Errorf("%w: expected exactly one checksum for %s, found %d", errChecksumInvalid, assetName, matches)
	}

	return matchingDigest, nil
}

func extractBinary(archive []byte, assetName string) ([]byte, error) {
	if strings.HasSuffix(assetName, ".zip") {
		return extractZipBinary(archive)
	}

	return extractTarGzBinary(archive)
}

func extractTarGzBinary(archive []byte) ([]byte, error) {
	gzipReader, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return nil, fmt.Errorf("open gzip archive: %w", err)
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)
	var binary []byte

	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			_, finishErr := io.CopyN(io.Discard, gzipReader, maxArchiveRemainder+1)
			if finishErr == nil {
				return nil, errArchiveTooLarge
			}

			if !errors.Is(finishErr, io.EOF) {
				return nil, fmt.Errorf("finish gzip archive: %w", finishErr)
			}

			if binary == nil {
				return nil, errArchiveNoBinary
			}

			return binary, nil
		}

		if err != nil {
			return nil, fmt.Errorf("read tar archive: %w", err)
		}

		if header.Typeflag != tar.TypeReg || !isBinaryName(filepath.Base(header.Name)) {
			continue
		}

		if binary != nil {
			continue
		}

		binary, err = io.ReadAll(tarReader)
		if err != nil {
			return nil, fmt.Errorf("read binary from tar archive: %w", err)
		}
	}
}

func extractZipBinary(archive []byte) ([]byte, error) {
	zipReader, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		return nil, fmt.Errorf("open zip archive: %w", err)
	}

	for _, file := range zipReader.File {
		if !file.Mode().IsRegular() || !isBinaryName(filepath.Base(file.Name)) {
			continue
		}

		reader, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("open zip binary: %w", err)
		}
		binary, readErr := io.ReadAll(reader)
		closeErr := reader.Close()

		if readErr != nil {
			return nil, fmt.Errorf("read zip binary: %w", readErr)
		}

		if closeErr != nil {
			return nil, fmt.Errorf("close zip binary: %w", closeErr)
		}

		return binary, nil
	}

	return nil, errArchiveNoBinary
}

func isBinaryName(name string) bool {
	return name == "clickup-cli" || name == "clickup-cli.exe"
}

func replaceExecutable(destination string, binary []byte, goos string) error {
	temporary, err := os.CreateTemp(filepath.Dir(destination), ".clickup-cli-update-*")
	if err != nil {
		return fmt.Errorf("create temporary executable: %w", err)
	}
	temporaryName := temporary.Name()

	defer func() { _ = removeFile(temporaryName) }()

	if _, err := temporary.Write(binary); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("write temporary executable: %w", err)
	}

	if err := temporary.Chmod(0o755); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("chmod temporary executable: %w", err)
	}

	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary executable: %w", err)
	}

	if goos != windows {
		if err := renameFile(temporaryName, destination); err != nil {
			return fmt.Errorf("replace executable: %w", err)
		}

		return nil
	}

	return replaceExecutableWindows(temporaryName, destination)
}

func replaceExecutableWindows(temporaryName, destination string) error {
	backup := destination + ".bak"
	if err := removeFile(backup); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove stale executable backup: %w", err)
	}

	hasBackup := true

	if err := renameFile(destination, backup); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("backup current executable: %w", err)
		}
		hasBackup = false
	}

	if err := renameFile(temporaryName, destination); err != nil {
		replaceErr := fmt.Errorf("replace executable: %w", err)
		if !hasBackup {
			return replaceErr
		}

		if rollbackErr := renameFile(backup, destination); rollbackErr != nil {
			return errors.Join(replaceErr, fmt.Errorf("rollback executable: %w", rollbackErr))
		}

		return replaceErr
	}

	if hasBackup {
		_ = removeFile(backup)
	}

	return nil
}
