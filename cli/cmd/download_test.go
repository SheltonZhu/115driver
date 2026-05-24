package cmd

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestResolveDownloadTargetPath_UsesExplicitFilePath(t *testing.T) {
	got := resolveDownloadTargetPath(filepath.Join("/tmp", "custom-name.mp4"), "remote-name.mp4")
	want := filepath.Join("/tmp", "custom-name.mp4")
	if got != want {
		t.Fatalf("unexpected target path: got %q want %q", got, want)
	}
}

func TestResolveDownloadTargetPath_AppendsNameForDirectoryHint(t *testing.T) {
	got := resolveDownloadTargetPath("/tmp/downloads/", "remote-name.mp4")
	want := filepath.Join("/tmp/downloads", "remote-name.mp4")
	if got != want {
		t.Fatalf("unexpected target path: got %q want %q", got, want)
	}
}

func TestResolveDownloadTargetPath_TreatsNonExistingExtensionlessPathAsFile(t *testing.T) {
	got := resolveDownloadTargetPath(filepath.Join("/tmp", "LICENSE"), "remote-name.mp4")
	want := filepath.Join("/tmp", "LICENSE")
	if got != want {
		t.Fatalf("unexpected target path: got %q want %q", got, want)
	}
}

func TestDownloadHTTPClientUsesConfiguredTimeout(t *testing.T) {
	client := newDownloadHTTPClient(90 * time.Second)
	if client.Timeout != 90*time.Second {
		t.Fatalf("expected configured timeout, got %s", client.Timeout)
	}
}

func TestDownloadHTTPClientDefaultTimeoutAllowsLargeDownloads(t *testing.T) {
	if defaultDownloadTimeout < time.Hour {
		t.Fatalf("expected default timeout to allow large downloads, got %s", defaultDownloadTimeout)
	}
}

func TestSaveDownloadResponsePreservesExistingFileOnFailure(t *testing.T) {
	target := filepath.Join(t.TempDir(), "target.txt")
	if err := os.WriteFile(target, []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("new")),
	}

	if err := saveDownloadResponse(target, resp, 2); err == nil {
		t.Fatal("expected short write limit to fail")
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "old" {
		t.Fatalf("expected existing file to be preserved, got %q", string(got))
	}
}
