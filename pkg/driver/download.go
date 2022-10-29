package driver

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	crypto "github.com/gaoyb7/115drive-webdav/115"
	"github.com/go-resty/resty/v2"
)

type FileDownloadUrl struct {
	Client int    `json:"client"`
	OssId  string `json:"oss_id"`
	Url    string `json:"url"`
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

// Download get download info with pickcode
func (c *Pan115Client) Download(pickCode string) (*DownloadInfo, error) {
	key := crypto.GenerateKey()

	result := DownloadReap{}
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
	resp, err := req.Post(ApiDownloadGetUrl)

	if err := CheckErr(err, &result, resp); err != nil {
		return nil, err
	}
	bytes, err := crypto.Decode(result.EncodedData, key)
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
