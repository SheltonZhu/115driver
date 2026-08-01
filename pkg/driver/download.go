package driver

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	crypto "github.com/SheltonZhu/115driver/pkg/crypto/m115"
	"github.com/go-resty/resty/v2"
)

type FileDownloadUrl struct {
	Client float64 `json:"client"`
	OSSID  string  `json:"oss_id"`
	Url    string  `json:"url"`
	Valid  bool    `json:"-"` // false when API returned false/null
}

// UnmarshalJSON handles both object and bool (false) responses from the API.
func (f *FileDownloadUrl) UnmarshalJSON(b []byte) error {
	// Handle false/null/empty cases
	if len(b) == 0 || string(b) == "false" || string(b) == "null" {
		*f = FileDownloadUrl{}
		return nil
	}
	// Handle object case
	type alias FileDownloadUrl
	var a alias
	if err := json.Unmarshal(b, &a); err != nil {
		return err
	}
	*f = FileDownloadUrl(a)
	f.Valid = true
	return nil
}

type DownloadInfo struct {
	FileName string          `json:"file_name"`
	FileSize StringInt64     `json:"file_size"`
	PickCode string          `json:"pick_code"`
	Url      FileDownloadUrl `json:"url"`
	Header   http.Header
}

// Get Download file from download info url
func (info *DownloadInfo) Get() (io.ReadSeeker, error) {
	req, err := http.NewRequest(http.MethodGet, info.Url.Url, nil)
	if err != nil {
		return nil, err
	}
	req.Header = info.Header.Clone()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(body), nil
}

type DownloadData map[string]*DownloadInfo

// DownloadWithUA get download info with pickcode and user agent
func (c *Pan115Client) DownloadWithUA(pickCode, ua string) (*DownloadInfo, error) {
	key := crypto.GenerateKey()

	result := DownloadResp{}
	params, err := json.Marshal(map[string]string{"pickcode": pickCode})
	if err != nil {
		return nil, err
	}

	data := crypto.Encode(params, key)
	req := c.NewRequest().
		SetQueryParam("t", Now().String()).
		SetFormData(map[string]string{"data": data}).
		ForceContentType("application/json").
		SetResult(&result)
	req = req.SetHeader("User-Agent", ua)
	resp, err := req.Post(ApiDownloadGetUrl)

	if err := CheckErr(err, &result, resp); err != nil {
		return nil, err
	}
	bytes, err := crypto.Decode(string(result.EncodedData), key)
	if err != nil {
		return nil, err
	}

	downloadInfo := DownloadData{}
	if err := json.Unmarshal(bytes, &downloadInfo); err != nil {
		return nil, err
	}

	for _, info := range downloadInfo {
		if info.FileSize < 0 {
			return nil, ErrDownloadEmpty
		}
		info.Header = buildDownloadHeaders(sentRequestHeaders(resp), resp.Cookies())
		return info, nil
	}
	return nil, ErrUnexpected
}

// DownloadWithUAByAndroidAPI get download info with pickcode and user agent
func (c *Pan115Client) DownloadWithUAByAndroidAPI(pickCode string, ua string) (*DownloadInfo, error) {
	key := crypto.GenerateKey()

	result := DownloadResp{}
	params, err := json.Marshal(map[string]string{"pick_code": pickCode})
	if err != nil {
		return nil, err
	}

	data := crypto.Encode(params, key)
	req := c.NewRequest().
		SetQueryParam("t", Now().String()).
		SetFormData(map[string]string{"data": data}).
		ForceContentType("application/json").
		SetResult(&result)
	req = req.SetHeader("User-Agent", ua)
	resp, err := req.Post(AndroidApiDownloadGetUrl)

	if err := CheckErr(err, &result, resp); err != nil {
		return nil, err
	}
	bytes, err := crypto.Decode(string(result.EncodedData), key)
	if err != nil {
		return nil, err
	}

	infoResp := struct {
		URL string `json:"url"`
	}{}
	if err := json.Unmarshal(bytes, &infoResp); err != nil {
		return nil, err
	}

	info := DownloadInfo{
		Url: FileDownloadUrl{
			Url: infoResp.URL,
		},
		PickCode: pickCode,
		Header:   buildDownloadHeaders(sentRequestHeaders(resp), resp.Cookies()),
	}

	return &info, nil
}

// sentRequestHeaders returns the request headers that were actually sent on
// the wire. resty keeps its own header map (exposed as resp.Request.Header)
// separate from RawRequest.Header — a deep copy made by createHTTPRequest —
// and the empty-UA sentinel is stripped from RawRequest only. Reading the
// raw request headers here is the source of truth for what the peer
// received, and matches the empty-UA handling in applyEmptyUAHandling.
func sentRequestHeaders(resp *resty.Response) http.Header {
	if resp == nil || resp.Request == nil || resp.Request.RawRequest == nil {
		return nil
	}
	return resp.Request.RawRequest.Header
}

// Download get download info with pickcode
func (c *Pan115Client) Download(pickCode string) (*DownloadInfo, error) {
	return c.DownloadWithUA(pickCode, "")
}

func buildDownloadHeaders(requestHeaders http.Header, responseCookies []*http.Cookie) http.Header {
	if requestHeaders == nil {
		requestHeaders = http.Header{}
	}
	headers := requestHeaders.Clone()
	if len(responseCookies) == 0 {
		return headers
	}

	cookies := make([]string, 0, len(responseCookies)+1)
	if existing := strings.TrimSpace(headers.Get("Cookie")); existing != "" {
		cookies = append(cookies, existing)
	}
	for _, cookie := range responseCookies {
		if cookie == nil {
			continue
		}
		cookies = append(cookies, cookie.String())
	}
	if len(cookies) > 0 {
		headers.Set("Cookie", strings.Join(cookies, "; "))
	}
	return headers
}

type SharedDownloadInfo struct {
	FileID   string      `json:"fid"`
	FileName string      `json:"fn"`
	FileSize StringInt64 `json:"fs"`
	URL      struct {
		URL    string `json:"url"`
		Client int    `json:"client"`
		Desc   any    `json:"desc"`
		Isp    any    `json:"isp"`
		OSSID  string `json:"oss_id"`
		OOID   string `json:"ooid"`
	} `json:"url"`
}

// DownloadByShareCode get download info with share code
func (c *Pan115Client) DownloadByShareCode(shareCode, receiveCode, fileID string) (*SharedDownloadInfo, error) {
	return c.DownloadByShareCodeWithUA("", shareCode, receiveCode, fileID)
}

func (c *Pan115Client) DownloadByShareCodeWithUA(ua, shareCode, receiveCode, fileID string) (*SharedDownloadInfo, error) {
	if isCalledByAlistV3() {
		return nil, ErrorNotSupportAlist
	}
	result := DownloadShareResp{}
	params := map[string]string{
		"share_code":   shareCode,
		"receive_code": receiveCode,
		"file_id":      fileID,
		"dl":           "1",
	}

	req := c.NewRequest().
		SetQueryParams(params).
		ForceContentType("application/json").
		SetHeader("referer", BuildShareReferer(shareCode, receiveCode)).
		SetHeader("User-Agent", ua).
		SetResult(&result)

	resp, err := req.Get(ApiDownloadGetShareUrl)

	if err := CheckErr(err, &result, resp); err != nil {
		return nil, err
	}

	downloadInfo := result.Data
	return &downloadInfo, nil
}
