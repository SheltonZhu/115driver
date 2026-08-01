package driver

import (
	"net/http"
	"strings"

	"github.com/go-resty/resty/v2"
)

// sentinelEmptyUA is a non-empty marker that signals the User-Agent header
// should be stripped before the HTTP request is sent. It prevents resty v2's
// middleware from overriding an empty UA with its default value.
//
// Do not remove or rename this without reading applyEmptyUAHandling first:
// the whole mechanism is an ad-hoc workaround for resty's User-Agent
// handling and depends on this exact marker.
const sentinelEmptyUA = "\x00__EMPTY_UA__"

// Pan115Client driver client
type Pan115Client struct {
	Client            *resty.Client
	Request           *resty.Request
	UserID            int64
	Userkey           string
	UploadMetaInfo    *UploadMetaInfo
	UseInternalUpload bool
	uaHandlingDone    bool
}

// New creates Client with customized options.
func New(opts ...Option) *Pan115Client {
	c := &Pan115Client{
		Client: resty.New(),
	}

	c.applyEmptyUAHandling()

	if len(opts) > 0 {
		for _, optFunc := range opts {
			optFunc(c)
		}
	}
	return c
}

// applyEmptyUAHandling installs the User-Agent sentinel hooks on the current
// resty client so that requests with an explicitly empty User-Agent are sent
// without any User-Agent header on the wire.
//
// Why this exists (tricky, ad-hoc workaround — read before touching):
//
// resty v2's parseRequestHeader middleware injects its default
// "go-resty/<version>" User-Agent whenever it considers the header empty
// (its IsStringEmpty check trims whitespace). There is no resty option to
// omit the User-Agent header entirely. CDN download links returned by 115
// may be bound to the UA used to fetch them, so callers need UA-free
// requests (e.g. DownloadWithUA(pickCode, "")); see
// https://github.com/SheltonZhu/115driver/issues/80.
//
// The workaround replaces the empty UA with a non-empty sentinel before
// resty's middleware runs (Hook 1), so the default UA is not injected, then
// strips the sentinel right before the HTTP request is sent (Hook 2). The
// bytes on the wire therefore carry no User-Agent header.
//
// The tricky part: resty v2.17.2's createHTTPRequest deep-copies
// r.Header into RawRequest.Header, so Hook 2 only strips the sentinel from
// the copy that is actually sent. resty's own r.Header (exposed later as
// resp.Request.Header) still contains the sentinel. Hook 3 (OnAfterResponse)
// aligns resp.Request.Header back to what was actually sent, so any API
// reading request headers after the response — now or in the future — gets
// the real value instead of the internal marker.
//
// Hooks are idempotent (uaHandlingDone). SetHttpClient and WithRestyClient
// replace the underlying resty client, so they reset the flag and re-install
// the hooks; keep that contract when adding other ways to swap the client.
func (c *Pan115Client) applyEmptyUAHandling() {
	if c.uaHandlingDone {
		return
	}

	// Hook 1: before resty's middleware — replace an explicitly empty UA
	// (request level) or client level with the sentinel. Client headers are
	// checked separately because they are merged into the request by
	// parseRequestHeader after this hook runs.
	c.Client.OnBeforeRequest(func(client *resty.Client, r *resty.Request) error {
		if isEmptyUA(r.Header) {
			r.Header.Set("User-Agent", sentinelEmptyUA)
			return nil
		}
		if isEmptyUA(client.Header) {
			client.SetHeader("User-Agent", sentinelEmptyUA)
		}
		return nil
	})

	// Hook 2: after resty's middleware — strip the sentinel from the
	// RawRequest that is actually sent, so the wire bytes have no
	// User-Agent header. Also restore the client-level header so it does
	// not leak the sentinel into subsequent requests.
	c.Client.SetPreRequestHook(func(client *resty.Client, req *http.Request) error {
		if client.Header.Get("User-Agent") == sentinelEmptyUA {
			client.SetHeader("User-Agent", "")
		}
		if req.Header.Get("User-Agent") == sentinelEmptyUA {
			req.Header.Set("User-Agent", "")
		}
		return nil
	})

	// Hook 3: after the response — resty's resp.Request.Header is its own
	// header map, not the one that went on the wire (it keeps the sentinel
	// because RawRequest.Header is a deep copy). Restore the real value so
	// every consumer of resp.Request.Header sees what was actually sent.
	c.Client.OnAfterResponse(func(client *resty.Client, resp *resty.Response) error {
		if resp != nil && resp.Request != nil &&
			resp.Request.Header.Get("User-Agent") == sentinelEmptyUA {
			resp.Request.Header.Set("User-Agent", "")
		}
		return nil
	})

	c.uaHandlingDone = true
}

// isEmptyUA reports whether the User-Agent header is explicitly present with
// an empty value: "", whitespace-only, or an empty/nil header slice. This
// mirrors resty's IsStringEmpty (trim-based) semantics for header values,
// which otherwise causes resty to inject its default User-Agent. A missing
// header key is not treated as empty UA — that is resty's default behavior
// and stays unchanged.
func isEmptyUA(h http.Header) bool {
	vals, exists := h["User-Agent"]
	return exists && (len(vals) == 0 || strings.TrimSpace(vals[0]) == "")
}

// Default creates an Client with default settings.
func Default() *Pan115Client {
	return New(UA())
}

// Defalut is deprecated: use Default instead. This function exists for backward compatibility.
func Defalut() *Pan115Client {
	return Default()
}

func (c *Pan115Client) SetHttpClient(httpClient *http.Client) *Pan115Client {
	c.Client = resty.NewWithClient(httpClient)
	c.uaHandlingDone = false
	c.applyEmptyUAHandling()
	return c
}

func (c *Pan115Client) SetUserAgent(userAgent string) *Pan115Client {
	c.Client.SetHeader("User-Agent", userAgent)
	return c
}

func (c *Pan115Client) SetCookies(cs ...*http.Cookie) *Pan115Client {
	c.Client.SetCookies(cs)
	return c
}

func (c *Pan115Client) SetDebug(d bool) *Pan115Client {
	c.Client.SetDebug(d)
	return c
}

func (c *Pan115Client) EnableTrace() *Pan115Client {
	c.Client.EnableTrace()
	return c
}

func (c *Pan115Client) SetProxy(proxy string) *Pan115Client {
	c.Client.SetProxy(proxy)
	return c
}

func (c *Pan115Client) NewRequest() *resty.Request {
	c.Request = c.Client.R()
	return c.Request
}

func (c *Pan115Client) GetRequest() *resty.Request {
	if c.Request != nil {
		return c.Request
	}
	return c.NewRequest()
}
