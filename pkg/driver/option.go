package driver

import (
	"net/http"

	"github.com/go-resty/resty/v2"
)

type Option func(c *Pan115Client)

func UA(userAgent ...string) Option {
	return func(c *Pan115Client) {
		if len(userAgent) > 0 {
			c.SetUserAgent(userAgent[0])
		} else {
			c.SetUserAgent(UADefalut)
		}
	}
}

func WithClient(hc *http.Client) Option {
	return func(c *Pan115Client) {
		c.SetHttpClient(hc)
	}
}

func WithRestyClient(resty *resty.Client) Option {
	return func(c *Pan115Client) {
		c.Client = resty
	}
}

func WithDebug() Option {
	return func(c *Pan115Client) {
		c.SetDebug(true)
	}
}

func WithTrace() Option {
	return func(c *Pan115Client) {
		c.EnableTrace()
	}
}

func WithProxy(proxy string) Option {
	return func(c *Pan115Client) {
		c.SetProxy(proxy)
	}
}
