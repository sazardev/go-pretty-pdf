package chromemgr

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// TestDownloadAndExtractIntegration exercises the real download+extract
// path against the live Chrome for Testing API. It touches the network
// and downloads ~100MB, so it's opt-in only (CHROMEMGR_INTEGRATION=1) and
// never runs as part of the normal `go test ./...` suite.
func TestDownloadAndExtractIntegration(t *testing.T) {
	if os.Getenv("CHROMEMGR_INTEGRATION") == "" {
		t.Skip("set CHROMEMGR_INTEGRATION=1 to run (downloads a real Chrome build)")
	}

	plat, err := platformStringFor(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		t.Skipf("no Chrome for Testing build for this platform: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	dir := t.TempDir()
	var events []string
	path, err := downloadAndExtract(ctx, dir, plat, func(msg string) { events = append(events, msg) })
	if err != nil {
		t.Fatalf("downloadAndExtract() error = %v", err)
	}
	if len(events) == 0 {
		t.Error("expected at least one progress event")
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("resolved binary does not exist: %v", err)
	}
	if info.Mode()&0o111 == 0 && runtime.GOOS != goosWindows {
		t.Errorf("resolved binary is not executable: mode=%v", info.Mode())
	}
	t.Logf("downloaded Chrome to %s (%d bytes)", path, info.Size())
}

func TestPlatformStringFor(t *testing.T) {
	tests := []struct {
		goos, goarch string
		want         string
		wantErr      bool
	}{
		{goosLinux, goarchAMD64, "linux64", false},
		{goosLinux, goarchARM64, "", true}, // no Chrome for Testing build published
		{goosDarwin, goarchAMD64, "mac-x64", false},
		{goosDarwin, goarchARM64, "mac-arm64", false},
		{goosWindows, goarchAMD64, "win64", false},
		{goosWindows, goarchARM64, "", true},
		{"freebsd", goarchAMD64, "", true},
	}

	for _, tt := range tests {
		got, err := platformStringFor(tt.goos, tt.goarch)
		if tt.wantErr {
			if err == nil {
				t.Errorf("platformStringFor(%s, %s) = %q, want error", tt.goos, tt.goarch, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("platformStringFor(%s, %s) unexpected error: %v", tt.goos, tt.goarch, err)
		}
		if got != tt.want {
			t.Errorf("platformStringFor(%s, %s) = %q, want %q", tt.goos, tt.goarch, got, tt.want)
		}
	}
}

func TestFindBinary(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "chrome-headless-shell-linux64")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	binPath := filepath.Join(nested, binaryName())
	if err := os.WriteFile(binPath, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	got := findBinary(dir)
	if got != binPath {
		t.Errorf("findBinary() = %q, want %q", got, binPath)
	}

	if got := findBinary(t.TempDir()); got != "" {
		t.Errorf("findBinary() on empty dir = %q, want empty", got)
	}
}

func buildZip(t *testing.T, entries map[string]string) string {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for name, content := range entries {
		f, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(t.TempDir(), "test.zip")
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// TestDownloadFileVerifiesGCSChecksum guards the integrity check
// downloadFile performs against GCS's X-Goog-Hash response header before
// trusting a downloaded Chrome build that's about to be chmod +x'd and
// executed: a download whose bytes match the reported MD5 must succeed.
func TestDownloadFileVerifiesGCSChecksum(t *testing.T) {
	content := []byte("fake-chrome-binary-bytes")
	sum := md5.Sum(content) //nolint:gosec
	digest := base64.StdEncoding.EncodeToString(sum[:])

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Goog-Hash", "crc32c=AAAAAA==,md5="+digest)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(content)
	}))
	defer srv.Close()

	dest := filepath.Join(t.TempDir(), "out.zip")
	if err := downloadFile(context.Background(), srv.URL, dest, nil); err != nil {
		t.Fatalf("downloadFile() error = %v", err)
	}

	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("reading downloaded file: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("downloaded content = %q, want %q", got, content)
	}
}

// TestDownloadFileRejectsChecksumMismatch is the flip side: a download
// whose bytes don't match the checksum GCS reported (corruption, or a
// tampering proxy that swaps bytes without recomputing the header) must be
// rejected and the partial/corrupt file removed rather than left behind
// for a later step to unwittingly extract and execute.
func TestDownloadFileRejectsChecksumMismatch(t *testing.T) {
	content := []byte("fake-chrome-binary-bytes")
	wrongSum := md5.Sum([]byte("something else entirely")) //nolint:gosec
	wrongDigest := base64.StdEncoding.EncodeToString(wrongSum[:])

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Goog-Hash", "md5="+wrongDigest)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(content)
	}))
	defer srv.Close()

	dest := filepath.Join(t.TempDir(), "out.zip")
	err := downloadFile(context.Background(), srv.URL, dest, nil)
	if err == nil {
		t.Fatal("expected an error for a checksum mismatch")
	}
	if _, statErr := os.Stat(dest); statErr == nil {
		t.Error("expected the corrupted/mismatched download to be removed, but it still exists")
	}
}

func TestGCSMD5FromHeader(t *testing.T) {
	validDigest := base64.StdEncoding.EncodeToString(md5.New().Sum(nil)) //nolint:gosec

	tests := []struct {
		name   string
		header http.Header
		wantOK bool
	}{
		{"present alongside crc32c", http.Header{"X-Goog-Hash": []string{"crc32c=AAAAAA==,md5=" + validDigest}}, true},
		{"crc32c only", http.Header{"X-Goog-Hash": []string{"crc32c=AAAAAA=="}}, false},
		{"absent", http.Header{}, false},
	}
	for _, tt := range tests {
		if _, ok := gcsMD5FromHeader(tt.header); ok != tt.wantOK {
			t.Errorf("%s: gcsMD5FromHeader() ok = %v, want %v", tt.name, ok, tt.wantOK)
		}
	}
}

func TestUnzipExtractsNormalEntries(t *testing.T) {
	zipPath := buildZip(t, map[string]string{
		"chrome-headless-shell-linux64/chrome-headless-shell": "binary-contents",
		"chrome-headless-shell-linux64/README":                "hello",
	})
	dest := t.TempDir()

	if err := unzip(zipPath, dest); err != nil {
		t.Fatalf("unzip() error = %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dest, "chrome-headless-shell-linux64", "chrome-headless-shell"))
	if err != nil {
		t.Fatalf("reading extracted binary: %v", err)
	}
	if string(got) != "binary-contents" {
		t.Errorf("extracted content = %q, want %q", got, "binary-contents")
	}
}

func TestUnzipBlocksZipSlip(t *testing.T) {
	zipPath := buildZip(t, map[string]string{
		"../../evil.sh":               "rm -rf /",
		"safe/../../../also-evil.txt": "still evil",
	})
	dest := t.TempDir()

	if err := unzip(zipPath, dest); err != nil {
		t.Fatalf("unzip() error = %v", err)
	}

	// Nothing should have escaped dest.
	parent := filepath.Dir(dest)
	if _, err := os.Stat(filepath.Join(parent, "evil.sh")); err == nil {
		t.Fatal("zip slip: evil.sh escaped the extraction directory")
	}
	if _, err := os.Stat(filepath.Join(parent, "also-evil.txt")); err == nil {
		t.Fatal("zip slip: also-evil.txt escaped the extraction directory")
	}
}
