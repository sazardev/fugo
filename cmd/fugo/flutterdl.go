package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// flutterClientReleaseRepo is where release-flutter-client.yml publishes the
// precompiled bundles, one per tagged release.
const flutterClientReleaseRepo = "sazardev/fugo"

// flutterClientDownloadURL builds the release asset URL for version+asset.
// It's a package variable (not a plain function) so tests can point it at a
// local httptest server instead of the real GitHub release endpoint.
var flutterClientDownloadURL = func(version, asset string) string {
	return fmt.Sprintf("https://github.com/%s/releases/download/v%s/%s", flutterClientReleaseRepo, version, asset)
}

// flutterClientAssetName returns the release asset name for the host OS, or
// "" if there is no precompiled client for it yet (only linux/x64 and
// windows/x64 are built today — see release-flutter-client.yml).
func flutterClientAssetName() string {
	switch runtime.GOOS {
	case "linux":
		return "fugo_flutter_client_linux_x64.tar.gz"
	case osWindows:
		return "fugo_flutter_client_windows_x64.tar.gz"
	default:
		return ""
	}
}

// flutterClientCacheDir is where a downloaded client bundle for a given fugo
// version + host platform is cached, so repeat runs never re-download.
func flutterClientCacheDir(version string) (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(base, "fugo", "flutter_client", version, runtime.GOOS+"_"+runtime.GOARCH), nil
}

// downloadedFlutterClientDir returns the cache dir for version if it's
// already populated, "" otherwise.
func downloadedFlutterClientDir(version string) string {
	dir, err := flutterClientCacheDir(version)
	if err != nil {
		return ""
	}

	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) == 0 {
		return ""
	}

	return dir
}

// isReleaseVersion reports whether v looks like a tagged release (X.Y.Z,
// optionally with a leading "v") rather than a placeholder default or a
// pseudo-version — only real releases have matching GitHub Release assets.
func isReleaseVersion(v string) bool {
	v = strings.TrimPrefix(v, "v")
	nums := strings.Split(strings.SplitN(v, "-", 2)[0], ".")
	if len(nums) != 3 {
		return false
	}

	for _, n := range nums {
		if n == "" {
			return false
		}
		for _, r := range n {
			if r < '0' || r > '9' {
				return false
			}
		}
	}

	return true
}

// downloadFlutterClient fetches and caches the precompiled Flutter client
// bundle for version + the host OS from the fugo GitHub Release, returning
// the extracted directory. On any failure it returns a plain error and
// leaves no partial cache directory behind — callers fall back to a local
// 'flutter build'.
func downloadFlutterClient(ctx context.Context, version string) (string, error) {
	if !isReleaseVersion(version) {
		return "", fmt.Errorf("not a tagged release version (%q) — no matching prebuilt client", version)
	}

	asset := flutterClientAssetName()
	if asset == "" {
		return "", fmt.Errorf("no precompiled Flutter client published for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	if dir := downloadedFlutterClientDir(version); dir != "" {
		return dir, nil
	}

	dir, err := flutterClientCacheDir(version)
	if err != nil {
		return "", fmt.Errorf("resolve cache dir: %w", err)
	}

	url := flutterClientDownloadURL(version, asset)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	client := &http.Client{Timeout: 2 * time.Minute}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("download %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download %s: HTTP %d (no prebuilt client published for this release/platform)", url, resp.StatusCode)
	}

	// Extract into a sibling .partial dir and rename into place atomically,
	// so a failed/interrupted download never leaves a cache dir that looks
	// populated (downloadedFlutterClientDir would treat it as usable).
	partial := dir + ".partial"
	if err := os.RemoveAll(partial); err != nil {
		return "", err
	}
	if err := os.MkdirAll(partial, 0o755); err != nil {
		return "", err
	}

	if err := extractTarGz(resp.Body, partial); err != nil {
		_ = os.RemoveAll(partial)

		return "", fmt.Errorf("extract %s: %w", asset, err)
	}

	if err := os.RemoveAll(dir); err != nil {
		_ = os.RemoveAll(partial)

		return "", err
	}
	if err := os.Rename(partial, dir); err != nil {
		_ = os.RemoveAll(partial)

		return "", fmt.Errorf("finalize cache dir: %w", err)
	}

	return dir, nil
}

// extractTarGz extracts a gzip-compressed tar stream into destDir, rejecting
// any entry that would escape destDir (zip-slip) and preserving each
// regular file's mode bits (so the render client binary keeps +x).
func extractTarGz(r io.Reader, destDir string) error {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer func() { _ = gz.Close() }()

	cleanDest := filepath.Clean(destDir)
	tr := tar.NewReader(gz)

	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}

		target := filepath.Join(cleanDest, filepath.Clean(hdr.Name))
		if target != cleanDest && !strings.HasPrefix(target, cleanDest+string(os.PathSeparator)) {
			return fmt.Errorf("tar entry escapes destination: %s", hdr.Name)
		}

		if err := extractTarEntry(tr, hdr, target); err != nil {
			return err
		}
	}
}

func extractTarEntry(tr *tar.Reader, hdr *tar.Header, target string) error {
	switch hdr.Typeflag {
	case tar.TypeDir:
		return os.MkdirAll(target, 0o755)
	case tar.TypeReg:
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}

		mode := os.FileMode(hdr.Mode) & 0o777 //nolint:gosec // mode bits from our own CI-built archive, not attacker input
		if mode == 0 {
			mode = 0o644
		}

		f, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
		if err != nil {
			return err
		}

		if _, err := io.Copy(f, tr); err != nil { //nolint:gosec // bundle comes from our own CI-built release asset, not arbitrary input
			_ = f.Close()

			return err
		}

		return f.Close()
	default:
		return nil // skip symlinks/etc — the client bundle only ever has dirs + regular files
	}
}
