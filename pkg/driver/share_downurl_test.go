package driver

import (
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestShareDownurlUA(t *testing.T) {
	cases := []struct {
		name string
		ua   string
		want string
	}{
		{"empty", "", UA115Browser},
		{"browser", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/142.0.0.0 Safari/537.36", UA115Browser},
		{"downloader", "aria2/1.36.0", UA115Browser},
		{"player", "Infuse/8.1", UA115Browser},
		{"115 browser", "Mozilla/5.0 115Browser/27.0.5.7", "Mozilla/5.0 115Browser/27.0.5.7"},
		{"115 browser newer", "Mozilla/5.0 115Browser/36.2.28", "Mozilla/5.0 115Browser/36.2.28"},
		{"115 disk", UA115Disk, UA115Disk},
		{"115 desktop", UA115Desktop, UA115Desktop},
		{"ios", UAIosApp, UAIosApp},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := shareDownurlUA(tc.ua); got != tc.want {
				t.Fatalf("shareDownurlUA(%q) = %q, want %q", tc.ua, got, tc.want)
			}
		})
	}
}

type shareDownurlCapture struct {
	method string
	path   string
	queryT string
	ua     string
	accept string
	data   string
}

func newShareDownurlTestEnv(t *testing.T, respBody string) (*Pan115Client, *shareDownurlCapture) {
	t.Helper()
	cap := &shareDownurlCapture{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		cap.method = r.Method
		cap.path = r.URL.Path
		cap.queryT = r.URL.Query().Get("t")
		cap.ua = r.Header.Get("User-Agent")
		cap.accept = r.Header.Get("Accept")
		cap.data = r.PostFormValue("data")
		_, _ = io.Copy(io.Discard, r.Body)
		_, _ = io.WriteString(w, respBody)
	}))
	t.Cleanup(server.Close)
	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	tr := &recordingTransport{base: &http.Transport{}, mockURL: u}
	return New(WithClient(&http.Client{Transport: tr})), cap
}

// TestDownloadByShareCodeWithUA_RequestShape locks the request against the
// captured vox.jar traffic: POST /app/share/downurl?t=<unix seconds> with an
// urlencoded data param carrying m115-encrypted JSON, a 115 client UA (a
// caller UA that is not a 115 client is replaced, see shareDownurlUA), and no
// referer.
func TestDownloadByShareCodeWithUA_RequestShape(t *testing.T) {
	client, cap := newShareDownurlTestEnv(t, `{"state":false,"error":"请重新登录","errno":99,"data":[]}`)

	// CheckErr surfaces the API's string error field directly (it takes
	// precedence over the errno mapping), matching the captured error body
	// {"state":false,"error":"请重新登录","errno":99}.
	_, err := client.DownloadByShareCodeWithUA("aria2/1.36.0", "swsexuo3zrk", "t58d", "3516278936403182800")
	if err == nil || !strings.Contains(err.Error(), "请重新登录") {
		t.Fatalf("err = %v, want login-expired API error", err)
	}

	if cap.method != http.MethodPost {
		t.Errorf("method = %s, want POST", cap.method)
	}
	if cap.path != "/app/share/downurl" {
		t.Errorf("path = %s, want /app/share/downurl", cap.path)
	}
	if len(cap.queryT) != 10 || strings.TrimLeft(cap.queryT, "0123456789") != "" {
		t.Errorf("t = %q, want unix-seconds timestamp", cap.queryT)
	}
	if cap.ua != UA115Browser {
		t.Errorf("UA = %q, want %q for non-115 caller UA", cap.ua, UA115Browser)
	}
	if cap.accept != "application/json;charset=UTF-8" {
		t.Errorf("Accept = %q, want application/json;charset=UTF-8", cap.accept)
	}

	raw, err := base64.StdEncoding.DecodeString(cap.data)
	if err != nil {
		t.Fatalf("data param is not valid base64: %v", err)
	}
	if len(raw) == 0 || len(raw)%128 != 0 {
		t.Errorf("data decodes to %d bytes, want a non-empty multiple of 128 (RSA blocks)", len(raw))
	}
}

func TestDownloadByShareCodeWithUA_Passthrough115UA(t *testing.T) {
	client, cap := newShareDownurlTestEnv(t, `{"state":false,"error":"请重新登录","errno":99,"data":[]}`)

	_, _ = client.DownloadByShareCodeWithUA(UA115Disk, "swsexuo3zrk", "t58d", "3516278936403182800")
	if cap.ua != UA115Disk {
		t.Fatalf("UA = %q, want 115 client UA passed through", cap.ua)
	}
}

func TestDownloadByShareCodeWithUA_StringError(t *testing.T) {
	// The legacy web endpoint answered the version gate with a bare string
	// error; make sure such a response surfaces as an error, not a panic or a
	// bogus success.
	client, _ := newShareDownurlTestEnv(t, `{"state":false,"error":"当前版本过低，请升级到最新版本下载。"}`)

	_, err := client.DownloadByShareCodeWithUA("", "swsexuo3zrk", "t58d", "3516278936403182800")
	if err == nil {
		t.Fatal("expected error for state:false response")
	}
	if !strings.Contains(err.Error(), "当前版本过低") {
		t.Fatalf("err = %v, want it to carry the API message", err)
	}
}
