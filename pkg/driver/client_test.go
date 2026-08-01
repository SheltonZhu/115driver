package driver

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-resty/resty/v2"
)

// recordingTransport records the User-Agent of each request as seen at the
// transport layer (i.e. what net/http would write to the wire) and can
// rewrite the request URL to a mock server, which is needed because the
// 115 download API endpoints are absolute URLs.
type recordingTransport struct {
	base    http.RoundTripper
	mockURL *url.URL
	wireUAs []string
}

func (t *recordingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.wireUAs = append(t.wireUAs, req.Header.Get("User-Agent"))
	if t.mockURL != nil {
		req.URL.Scheme = t.mockURL.Scheme
		req.URL.Host = t.mockURL.Host
	}
	return t.base.RoundTrip(req)
}

// newUATestEnv starts a mock server that records the received User-Agent and
// returns it together with a recording transport and a client wired to it.
func newUATestEnv(t *testing.T) (*httptest.Server, *recordingTransport, *[]string) {
	t.Helper()
	var serverUAs []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverUAs = append(serverUAs, r.Header.Get("User-Agent"))
		_, _ = io.Copy(io.Discard, r.Body)
		_, _ = w.Write([]byte(`{"state":false,"error":"mock API"}`))
	}))
	tr := &recordingTransport{base: &http.Transport{}}
	return server, tr, &serverUAs
}

func assertNoUA(t *testing.T, wireUAs, serverUAs []string) {
	t.Helper()
	if len(wireUAs) == 0 {
		t.Fatal("no request observed at transport layer")
	}
	for i, ua := range wireUAs {
		if ua != "" {
			t.Fatalf("transport UA #%d = %q, want empty", i, ua)
		}
	}
	for i, ua := range serverUAs {
		if ua != "" {
			t.Fatalf("server UA #%d = %q, want empty", i, ua)
		}
	}
}

// TestEmptyUAHandling verifies that an explicitly empty User-Agent never
// reaches the wire as resty's default UA, and that the headers visible after
// the response reflect what was actually sent.
func TestEmptyUAHandling(t *testing.T) {
	t.Run("request-level-empty-string", func(t *testing.T) {
		server, tr, serverUAs := newUATestEnv(t)
		defer server.Close()
		client := New(WithClient(&http.Client{Transport: tr}))

		resp, err := client.NewRequest().SetHeader("User-Agent", "").Post(server.URL)
		if err != nil {
			t.Fatal(err)
		}
		assertNoUA(t, tr.wireUAs, *serverUAs)
		if got := resp.Request.Header.Get("User-Agent"); got != "" {
			t.Fatalf("resp.Request.Header UA = %q, want empty (OnAfterResponse cleanup)", got)
		}
		if got := sentRequestHeaders(resp).Get("User-Agent"); got != "" {
			t.Fatalf("sentRequestHeaders UA = %q, want empty", got)
		}
	})

	t.Run("request-level-whitespace", func(t *testing.T) {
		server, tr, serverUAs := newUATestEnv(t)
		defer server.Close()
		client := New(WithClient(&http.Client{Transport: tr}))

		if _, err := client.NewRequest().SetHeader("User-Agent", " ").Post(server.URL); err != nil {
			t.Fatal(err)
		}
		assertNoUA(t, tr.wireUAs, *serverUAs)
	})

	t.Run("request-level-nil-values", func(t *testing.T) {
		server, tr, serverUAs := newUATestEnv(t)
		defer server.Close()
		client := New(WithClient(&http.Client{Transport: tr}))

		r := client.NewRequest()
		r.Header["User-Agent"] = nil
		if _, err := r.Post(server.URL); err != nil {
			t.Fatal(err)
		}
		assertNoUA(t, tr.wireUAs, *serverUAs)
	})

	t.Run("real-ua-preserved", func(t *testing.T) {
		server, tr, _ := newUATestEnv(t)
		defer server.Close()
		client := New(WithClient(&http.Client{Transport: tr}))
		const ua = "Mozilla/5.0 115Browser/26.0.0.3"

		if _, err := client.NewRequest().SetHeader("User-Agent", ua).Post(server.URL); err != nil {
			t.Fatal(err)
		}
		if len(tr.wireUAs) != 1 || tr.wireUAs[0] != ua {
			t.Fatalf("wire UA = %v, want %q", tr.wireUAs, ua)
		}
	})

	t.Run("client-level-empty", func(t *testing.T) {
		server, tr, serverUAs := newUATestEnv(t)
		defer server.Close()
		client := New(WithClient(&http.Client{Transport: tr}))
		client.SetUserAgent("")

		if _, err := client.NewRequest().Post(server.URL); err != nil {
			t.Fatal(err)
		}
		assertNoUA(t, tr.wireUAs, *serverUAs)
	})

	t.Run("after-set-http-client", func(t *testing.T) {
		server, tr, serverUAs := newUATestEnv(t)
		defer server.Close()
		client := New()
		client.SetHttpClient(&http.Client{Transport: tr})

		if _, err := client.NewRequest().SetHeader("User-Agent", "").Post(server.URL); err != nil {
			t.Fatal(err)
		}
		assertNoUA(t, tr.wireUAs, *serverUAs)
	})

	t.Run("after-with-resty-client", func(t *testing.T) {
		server, tr, serverUAs := newUATestEnv(t)
		defer server.Close()
		client := New(WithRestyClient(resty.NewWithClient(&http.Client{Transport: tr})))

		if _, err := client.NewRequest().SetHeader("User-Agent", "").Post(server.URL); err != nil {
			t.Fatal(err)
		}
		assertNoUA(t, tr.wireUAs, *serverUAs)
	})
}

// TestDownloadWithUA_EmptyUA_SendsNoUA exercises the real DownloadWithUA call
// (with the actual absolute API URL) against a mock server through the
// recording transport. The mock returns an API error so the request follows
// the error path — the 115 response encryption cannot be mocked without the
// server's RSA private exponent, so a successful decrypted response is not
// feasible in tests.
func TestDownloadWithUA_EmptyUA_SendsNoUA(t *testing.T) {
	server, tr, serverUAs := newUATestEnv(t)
	defer server.Close()
	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	tr.mockURL = u
	client := New(WithClient(&http.Client{Transport: tr}))

	info, err := client.DownloadWithUA("pickcode", "")
	if err == nil {
		t.Fatalf("expected mock API error, got info=%v", info)
	}
	assertNoUA(t, tr.wireUAs, *serverUAs)
}
