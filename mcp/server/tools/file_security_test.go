package tools

import (
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestValidateUploadURLRejectsUnsafeTargets(t *testing.T) {
	tests := []string{
		"file:///etc/passwd",
		"ftp://example.com/file.bin",
		"http://127.0.0.1/file.bin",
		"http://localhost/file.bin",
		"http://[::1]/file.bin",
	}

	for _, rawURL := range tests {
		t.Run(rawURL, func(t *testing.T) {
			if _, err := validateUploadURL(rawURL); err == nil {
				t.Fatalf("expected %q to be rejected", rawURL)
			}
		})
	}
}

func TestValidateUploadURLAcceptsHTTPSURL(t *testing.T) {
	got, err := validateUploadURL("https://example.com/path/file.bin?token=abc")
	if err != nil {
		t.Fatalf("expected URL to be accepted: %v", err)
	}
	if got.String() != "https://example.com/path/file.bin?token=abc" {
		t.Fatalf("unexpected parsed URL: %s", got.String())
	}
}

func TestValidateLocalPathRequiresConfiguredRoot(t *testing.T) {
	if _, err := validateLocalPath("", "/tmp/out.bin", false); err == nil {
		t.Fatal("expected empty local root to reject local file access")
	}
}

func TestValidateLocalPathRejectsPathOutsideRoot(t *testing.T) {
	root := t.TempDir()
	if _, err := validateLocalPath(root, root+"/../outside.bin", false); err == nil {
		t.Fatal("expected path outside root to be rejected")
	}
}

func TestValidateLocalPathAcceptsPathInsideRoot(t *testing.T) {
	root := t.TempDir()
	got, err := validateLocalPath(root, root+"/nested/out.bin", false)
	if err != nil {
		t.Fatalf("expected path inside root to be accepted: %v", err)
	}
	if !strings.HasPrefix(got, root) {
		t.Fatalf("expected %q to stay under %q", got, root)
	}
}

func TestValidateLocalPathRejectsExistingSymlinkFileOutsideRoot(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.txt")
	if err := os.WriteFile(outside, []byte("outside"), 0600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "link")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatal(err)
	}

	if _, err := validateLocalPath(root, link, false); err == nil {
		t.Fatal("expected symlink target outside root to be rejected")
	}
}

func TestCopyHTTPResponseRequiresStatusOK(t *testing.T) {
	var out strings.Builder
	resp := &http.Response{
		StatusCode: http.StatusForbidden,
		Body:       io.NopCloser(strings.NewReader("denied")),
	}

	err := copyHTTPResponse(&out, resp, 1024)
	if err == nil {
		t.Fatal("expected non-200 response to fail")
	}
	if !errors.Is(err, errUnexpectedHTTPStatus) {
		t.Fatalf("expected errUnexpectedHTTPStatus, got %v", err)
	}
	if out.Len() != 0 {
		t.Fatalf("expected response body not to be copied, wrote %d bytes", out.Len())
	}
}

func TestCopyHTTPResponseEnforcesSizeLimit(t *testing.T) {
	var out strings.Builder
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("abcdef")),
	}

	err := copyHTTPResponse(&out, resp, 3)
	if err == nil {
		t.Fatal("expected oversized response to fail")
	}
	if !errors.Is(err, errResponseTooLarge) {
		t.Fatalf("expected errResponseTooLarge, got %v", err)
	}
}

func TestCopyHTTPResponseAllowsUnlimitedSizeWhenLimitIsZero(t *testing.T) {
	var out strings.Builder
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("abcdef")),
	}

	if err := copyHTTPResponse(&out, resp, 0); err != nil {
		t.Fatalf("expected zero limit to allow response: %v", err)
	}
	if out.String() != "abcdef" {
		t.Fatalf("unexpected copied body: %q", out.String())
	}
}

func TestMCPDefaultDownloadSizeAllowsLargeDownloads(t *testing.T) {
	if defaultMCPDownloadMaxBytes != 0 {
		t.Fatalf("expected default MCP download size to be unlimited, got %d", defaultMCPDownloadMaxBytes)
	}
}

func TestMCPDefaultURLUploadSizeRemainsBounded(t *testing.T) {
	if defaultMCPURLUploadMaxBytes <= 0 {
		t.Fatalf("expected URL upload default size to be bounded, got %d", defaultMCPURLUploadMaxBytes)
	}
}

func TestMCPHTTPClientUsesConfiguredTimeout(t *testing.T) {
	client := newMCPHTTPClient(90 * time.Second)
	if client.Timeout != 90*time.Second {
		t.Fatalf("expected configured timeout, got %s", client.Timeout)
	}
}

func TestMCPHTTPClientRejectsUnsafeRedirect(t *testing.T) {
	client := newMCPHTTPClient(90 * time.Second)
	req := &http.Request{URL: mustParseURL(t, "http://127.0.0.1/private")}
	if err := client.CheckRedirect(req, nil); err == nil {
		t.Fatal("expected redirect to unsafe host to be rejected")
	}
}

func TestValidateResolvedIPsRejectsPrivateAddress(t *testing.T) {
	if err := validateResolvedIPs("example.com", []net.IP{net.ParseIP("10.0.0.1")}); err == nil {
		t.Fatal("expected private resolved IP to be rejected")
	}
}

func TestSaveHTTPResponseToFilePreservesExistingFileOnFailure(t *testing.T) {
	target := filepath.Join(t.TempDir(), "target.txt")
	if err := os.WriteFile(target, []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("abcdef")),
	}

	if err := saveHTTPResponseToFile(target, resp, 3); err == nil {
		t.Fatal("expected oversized response to fail")
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "old" {
		t.Fatalf("expected existing file to be preserved, got %q", string(got))
	}
}

func TestMCPDefaultDownloadTimeoutAllowsLargeDownloads(t *testing.T) {
	if defaultMCPDownloadTimeout < time.Hour {
		t.Fatalf("expected default timeout to allow large downloads, got %s", defaultMCPDownloadTimeout)
	}
}

func mustParseURL(t *testing.T, rawURL string) *url.URL {
	t.Helper()
	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}
