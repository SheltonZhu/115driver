package driver

import (
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
)

func TestBuildDownloadHeaders_MergesResponseCookiesIntoCookieHeader(t *testing.T) {
	requestHeaders := http.Header{}
	requestHeaders.Set("User-Agent", UA115Browser)
	requestHeaders.Set("Cookie", "UID=1; CID=2")

	responseCookies := []*http.Cookie{
		{
			Name:  "download_token",
			Value: "abc123",
		},
	}

	got := buildDownloadHeaders(requestHeaders, responseCookies)

	if got.Get("User-Agent") != UA115Browser {
		t.Fatalf("unexpected user agent: %q", got.Get("User-Agent"))
	}
	if got.Get("Cookie") != "UID=1; CID=2; download_token=abc123" {
		t.Fatalf("unexpected cookie header: %q", got.Get("Cookie"))
	}
}

func TestBuildDownloadHeaders_NilRequestHeaders(t *testing.T) {
	responseCookies := []*http.Cookie{{Name: "download_token", Value: "abc123"}}
	got := buildDownloadHeaders(nil, responseCookies)
	if got.Get("Cookie") != "download_token=abc123" {
		t.Fatalf("unexpected cookie header: %q", got.Get("Cookie"))
	}
}

func TestSentRequestHeadersReturnsWireHeaders(t *testing.T) {
	rawReq := &http.Request{
		Header: http.Header{
			"User-Agent": {""},
			"Referer":    {"https://115.com/"},
		},
	}
	resp := &resty.Response{Request: &resty.Request{RawRequest: rawReq}}

	got := sentRequestHeaders(resp)
	if got.Get("User-Agent") != "" {
		t.Fatalf("User-Agent = %q, want empty", got.Get("User-Agent"))
	}
	if got.Get("Referer") != "https://115.com/" {
		t.Fatalf("Referer = %q, want preserved", got.Get("Referer"))
	}

	if sentRequestHeaders(nil) != nil {
		t.Fatal("sentRequestHeaders(nil) = non-nil, want nil")
	}
	if sentRequestHeaders(&resty.Response{}) != nil {
		t.Fatal("sentRequestHeaders(response without request) = non-nil, want nil")
	}
}
