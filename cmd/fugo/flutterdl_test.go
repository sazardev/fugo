package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestIsReleaseVersion(t *testing.T) {
	t.Parallel()

	cases := map[string]bool{
		"1.2.3":       true,
		"v1.2.3":      true,
		"0.1.0":       true,
		"1.2.3-dev":   true, // pseudo-version suffix stripped before checking
		"":            false,
		"devel":       false,
		"1.2":         false,
		"1.2.3.4":     false,
		"v1.2.x":      false,
		"1.2.3-0.202": true,
	}

	for v, want := range cases {
		if got := isReleaseVersion(v); got != want {
			t.Errorf("isReleaseVersion(%q) = %v, want %v", v, got, want)
		}
	}
}

// buildTestTarGz creates an in-memory gzip-compressed tar archive containing
// the given files (path -> content) plus one executable "fugo_flutter_client"
// entry, for exercising extractTarGz/downloadFlutterClient without a real
// release asset.
func buildTestTarGz(t *testing.T, files map[string]string) []byte {
	t.Helper()

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)

	for name, content := range files {
		mode := int64(0o644)
		if name == "fugo_flutter_client" {
			mode = 0o755
		}

		hdr := &tar.Header{Name: name, Mode: mode, Size: int64(len(content))}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}

	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}

	return buf.Bytes()
}

func TestExtractTarGz(t *testing.T) {
	t.Parallel()

	archive := buildTestTarGz(t, map[string]string{
		"fugo_flutter_client":                    "binary-content",
		"data/flutter_assets/AssetManifest.json": "{}",
	})

	dest := t.TempDir()
	if err := extractTarGz(bytes.NewReader(archive), dest); err != nil {
		t.Fatalf("extractTarGz: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dest, "fugo_flutter_client"))
	if err != nil {
		t.Fatalf("read extracted binary: %v", err)
	}
	if string(got) != "binary-content" {
		t.Errorf("binary content = %q, want %q", got, "binary-content")
	}

	info, err := os.Stat(filepath.Join(dest, "fugo_flutter_client"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o100 == 0 {
		t.Error("extracted binary lost its executable bit")
	}

	assetContent, err := os.ReadFile(filepath.Join(dest, "data", "flutter_assets", "AssetManifest.json"))
	if err != nil {
		t.Fatalf("read nested file: %v", err)
	}
	if string(assetContent) != "{}" {
		t.Errorf("nested file content = %q, want {}", assetContent)
	}
}

func TestExtractTarGzRejectsZipSlip(t *testing.T) {
	t.Parallel()

	archive := buildTestTarGz(t, map[string]string{
		"../../etc/passwd": "pwned",
	})

	dest := t.TempDir()
	err := extractTarGz(bytes.NewReader(archive), dest)
	if err == nil {
		t.Fatal("extractTarGz did not reject a path-escaping tar entry")
	}
}

func TestDownloadFlutterClientCachesAndSkipsRedownload(t *testing.T) {
	// Not t.Parallel(): mutates the package-level flutterClientDownloadURL
	// seam and the process-wide cache dir env vars.

	requests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(buildTestTarGz(t, map[string]string{"fugo_flutter_client": "v1"}))
	}))
	defer srv.Close()

	cacheRoot := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", cacheRoot) // os.UserCacheDir() honors this on Linux
	t.Setenv("HOME", cacheRoot)           // fallback UserCacheDir() path on some platforms

	originalURL := flutterClientDownloadURL
	flutterClientDownloadURL = func(_, _ string) string { return srv.URL }
	defer func() { flutterClientDownloadURL = originalURL }()

	const version = "9.9.9"
	if !isReleaseVersion(version) {
		t.Fatal("test version must look like a release for downloadFlutterClient to attempt it")
	}
	if flutterClientAssetName() == "" {
		t.Skip("no precompiled asset name for this test's GOOS — nothing to exercise")
	}
	if downloadedFlutterClientDir(version) != "" {
		t.Fatal("cache dir reported populated before any download")
	}

	ctx := context.Background()

	dir, err := downloadFlutterClient(ctx, version)
	if err != nil {
		t.Fatalf("first download: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "fugo_flutter_client")); err != nil {
		t.Fatalf("extracted binary missing: %v", err)
	}
	if requests != 1 {
		t.Fatalf("requests = %d, want 1", requests)
	}

	if downloadedFlutterClientDir(version) != dir {
		t.Errorf("downloadedFlutterClientDir after caching = %q, want %q", downloadedFlutterClientDir(version), dir)
	}

	// A second call must hit the cache, not the server again.
	dir2, err := downloadFlutterClient(ctx, version)
	if err != nil {
		t.Fatalf("second download: %v", err)
	}
	if dir2 != dir {
		t.Errorf("second call dir = %q, want same as first %q", dir2, dir)
	}
	if requests != 1 {
		t.Errorf("requests after cached call = %d, want still 1", requests)
	}
}

func TestDownloadFlutterClientRejectsNonReleaseVersion(t *testing.T) {
	t.Parallel()

	// "devel" fails isReleaseVersion, so this returns before any network
	// call — no real HTTP request should ever happen in this test.
	if _, err := downloadFlutterClient(context.Background(), "devel"); err == nil {
		t.Error("downloadFlutterClient did not reject a non-release version")
	}
}
