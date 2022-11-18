package driver

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
)

// Option driver client options
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

const (
	FileOrderByTime = "user_ptime"
	FileOrderByType = "file_type"
	FileOrderBySize = "file_size"
	FileOrderByName = "file_name"

	FileListLimit = int64(56)
)

// GetFileOption get file options
type GetFileOption struct {
	order    string
	asc      string
	pageSize int64
	offset   int64
	showDir  string
}

type GetFileOptions func(o *GetFileOption)

func WithLimit(pageSize int64) GetFileOptions {
	return func(o *GetFileOption) {
		o.pageSize = pageSize
	}
}

func WithOffset(offset int64) GetFileOptions {
	return func(o *GetFileOption) {
		o.offset = offset
	}
}

func WithOrder(order string) GetFileOptions {
	return func(o *GetFileOption) {
		o.order = order
	}
}

func WithShowDirEnable(e bool) GetFileOptions {
	return func(o *GetFileOption) {
		o.showDir = "0"
		if e {
			o.showDir = "1"
		}
	}
}

func WithAsc(d bool) GetFileOptions {
	return func(o *GetFileOption) {
		o.showDir = "0"
		if d {
			o.showDir = "1"
		}
	}
}

func (o *GetFileOption) GetOrder() string {
	return o.order
}

func (o *GetFileOption) GetAsc() string {
	return o.asc
}

func (o *GetFileOption) GetPageSize() string {
	return strconv.FormatInt(o.pageSize, 10)
}

func (o *GetFileOption) GetOffset() string {
	return strconv.FormatInt(o.offset, 10)
}

func (o *GetFileOption) GetshowDir() string {
	return o.showDir
}

func DefaultGetFileOptions() *GetFileOption {
	return &GetFileOption{
		order:    FileOrderByTime,
		asc:      "1",
		pageSize: int64(56),
		offset:   int64(0),
		showDir:  "1",
	}
}

type UploadMultipartOptions struct {
	ThreadsNum       int
	Timeout          time.Duration
	TokenRefreshTime time.Duration
}

func DefalutUploadMultipartOptions() *UploadMultipartOptions {
	return &UploadMultipartOptions{
		ThreadsNum:       10,
		Timeout:          time.Hour * 24,
		TokenRefreshTime: time.Minute * 50,
	}
}

type UploadMultipartOption func(o *UploadMultipartOptions)

func UploadMultipartWithThreadsNum(n int) UploadMultipartOption {
	return func(o *UploadMultipartOptions) {
		o.ThreadsNum = n
	}
}

func UploadMultipartWithTimeout(timeout time.Duration) UploadMultipartOption {
	return func(o *UploadMultipartOptions) {
		o.Timeout = timeout
	}
}

func UploadMultipartWithTokenRefreshTime(refreshTime time.Duration) UploadMultipartOption {
	return func(o *UploadMultipartOptions) {
		o.TokenRefreshTime = refreshTime
	}
}
