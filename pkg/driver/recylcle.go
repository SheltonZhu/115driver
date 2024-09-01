package driver

import (
	"fmt"
	"net/url"
	"strconv"
)

// CleanRecycleBin clean the recycle bin
func (c *Pan115Client) CleanRecycleBin(password string, rIDs ...string) error {
	form := url.Values{}
	form.Set("password", password)
	for idx, rID := range rIDs {
		form.Add(fmt.Sprintf("rid[%d]", idx), rID)
	}
	result := BasicResp{}
	req := c.NewRequest().
		SetFormDataFromValues(form).
		SetResult(&result).
		ForceContentType("application/json;charset=UTF-8")

	resp, err := req.Post(ApiRecycleClean)
	return CheckErr(err, &result, resp)
}

// ListRecycleBin list the recycle bin
func (c *Pan115Client) ListRecycleBin(offset, limit int) ([]RecycleBinItem, error) {
	result := RecycleListResponse{}
	req := c.NewRequest().
		SetQueryParams(map[string]string{
			"aid":    "7",
			"cid":    "0",
			"format": "json",
			"offset": strconv.Itoa(offset),
			"limit":  strconv.Itoa(limit),
		}).
		SetResult(&result).
		ForceContentType("application/json;charset=UTF-8")

	resp, err := req.Get(ApiRecycleList)
	err = CheckErr(err, &result, resp)
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}

type RecycleListResponse struct {
	BasicResp
	Data []RecycleBinItem `json:"data"`
}

type RecycleBinItem struct {
	FileId     string      `json:"id"`
	FileName   string      `json:"file_name"`
	FileSize   StringInt64 `json:"file_size"`
	ParentId   IntString   `json:"cid"`
	ParentName string      `json:"parent_name"`
	DeleteTime StringInt64 `json:"dtime"`
}

// RevertRecycleBin revert the recycle bin
func (c *Pan115Client) RevertRecycleBin(rIDs ...string) error {
	form := url.Values{}
	for idx, rID := range rIDs {
		form.Add(fmt.Sprintf("rid[%d]", idx), rID)
	}
	result := BasicResp{}
	req := c.NewRequest().
		SetFormDataFromValues(form).
		SetResult(&result).
		ForceContentType("application/json;charset=UTF-8")

	resp, err := req.Post(ApiRecycleRevert)
	return CheckErr(err, &result, resp)
}
