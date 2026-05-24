package cmd

import (
	"path/filepath"
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
