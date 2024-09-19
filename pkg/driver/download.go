package driver

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	crypto "github.com/SheltonZhu/115driver/pkg/crypto/m115"
	"github.com/go-resty/resty/v2"
)

type FileDownloadUrl struct {
	Client float64 `json:"client"`
	OSSID  string  `json:"oss_id"`
	Url    string  `json:"url"`
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
	req := resty.New().R().SetHeaderMultiValues(info.Header)
	resp, err := req.Get(info.Url.Url)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(resp.Body()), nil
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
	if len(ua) > 0 {
		req = req.SetHeader("User-Agent", ua)
	}
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
		info.Header = resp.Request.Header
		return info, nil
	}
	return nil, ErrUnexpected
}

// Download get download info with pickcode
func (c *Pan115Client) Download(pickCode string) (*DownloadInfo, error) {
	return c.DownloadWithUA(pickCode, "")
}

type SharedDownloadInfo struct {
	FileID   string      `json:"fid"`
	FileName string      `json:"fn"`
	FileSize StringInt64 `json:"fs"`
	URL      struct {
		URL    string      `json:"url"`
		Client int         `json:"client"`
		Desc   interface{} `json:"desc"`
		Isp    interface{} `json:"isp"`
	} `json:"url"`
}

// DownloadByShareCode get download info with share code
func (c *Pan115Client) DownloadByShareCode(shareCode, receiveCode, fileID string) (*SharedDownloadInfo, error) {
	key := crypto.GenerateKey()

	result := DownloadResp{}
	params, err := json.Marshal(map[string]string{
		"share_code":   shareCode,
		"receive_code": receiveCode,
		"file_id":      fileID,
	})
	if err != nil {
		return nil, err
	}

	data := crypto.Encode(params, key)
	req := c.NewRequest().
		SetQueryParam("t", Now().String()).
		SetFormData(map[string]string{"data": data}).
		ForceContentType("application/json").
		SetResult(&result)
	// if len(ua) > 0 {
	// req = req.SetHeader("User-Agent", ua)
	// }
	resp, err := req.Post(ApiDownloadGetShareUrl)

	if err := CheckErr(err, &result, resp); err != nil {
		return nil, err
	}
	bytes, err := crypto.Decode(string(result.EncodedData), key)
	if err != nil {
		return nil, err
	}

	downloadInfo := SharedDownloadInfo{}
	if err := json.Unmarshal(bytes, &downloadInfo); err != nil {
		return nil, err
	}

	if downloadInfo.FileSize < 0 {
		return nil, ErrDownloadEmpty
	}

	return &downloadInfo, nil
}
