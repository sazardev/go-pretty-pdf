package chromemgr

import (
	"archive/zip"
	"bytes"
	"context"
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
	if info.Mode()&0o111 == 0 && runtime.GOOS != "windows" {
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
		{"linux", "amd64", "linux64", false},
		{"linux", "arm64", "", true}, // no Chrome for Testing build published
		{"darwin", "amd64", "mac-x64", false},
		{"darwin", "arm64", "mac-arm64", false},
		{"windows", "amd64", "win64", false},
		{"windows", "arm64", "", true},
		{"freebsd", "amd64", "", true},
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
