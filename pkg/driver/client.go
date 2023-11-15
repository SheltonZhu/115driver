package driver

import (
	"net/http"

	"github.com/go-resty/resty/v2"
)

// Pan115Client driver client
type Pan115Client struct {
	Client            *resty.Client
	Request           *resty.Request
	UserID            int64
	Userkey           string
	UploadMetaInfo    *UploadMetaInfo
	UseInternalUpload bool
}

// New creates Client with customized options.
func New(opts ...Option) *Pan115Client {
	c := &Pan115Client{
		Client: resty.New(),
	}
	if len(opts) > 0 {
		for _, optFunc := range opts {
			optFunc(c)
		}
	}
	return c
}

// Defalut creates an Client with default settings.
func Defalut() *Pan115Client {
	return New(UA())
}

func (c *Pan115Client) SetHttpClient(httpClient *http.Client) *Pan115Client {
	c.Client = resty.NewWithClient(httpClient)
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
