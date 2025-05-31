package driver

import (
	"strings"

	"github.com/go-resty/resty/v2"
)

// Mkdir make a new directory which name and parent directory id, return directory id
func (c *Pan115Client) Mkdir(parentID string, name string) (string, error) {
	result := MkdirResp{}
	form := map[string]string{
		"pid":   parentID,
		"cname": name,
	}
	req := c.NewRequest().
		SetFormData(form).
		SetResult(&result).
		ForceContentType("application/json;charset=UTF-8")

	resp, err := req.Post(ApiDirAdd)

	err = CheckErr(err, &result, resp)
	if err != nil {
		return "", err
	}
	return string(result.CategoryID), nil
}

// List list all files and directories
func (c *Pan115Client) List(dirID string, opts ...ListOption) (*[]File, error) {
	return c.ListWithLimit(dirID, FileListLimit, opts...)
}

const MaxDirPageLimit = 1150

// ListWithLimit list all files and directories with limit
func (c *Pan115Client) ListWithLimit(dirID string, limit int64, opts ...ListOption) (*[]File, error) {
	if limit > MaxDirPageLimit {
		limit = MaxDirPageLimit
	}

	o := DefaultListOptions()
	if len(opts) > 0 {
		for _, opt := range opts {
			opt(o)
		}
	}

	apiURLs := o.ApiURLs
	var files []File
	offset := int64(0)
	for i := 0; ; i++ {
		apiURL := apiURLs[i%len(apiURLs)]
		req := c.NewRequest().ForceContentType("application/json;charset=UTF-8")
		getFilesOpts := []GetFileOptions{
			WithApiURL(apiURL),
			WithLimit(limit),
			WithOffset(offset),
		}
		result, err := GetFiles(req, dirID, getFilesOpts...)
		if err != nil {
			return nil, err
		}
		for _, fileInfo := range result.Files {
			files = append(files, *(&File{}).from(&fileInfo))
		}
		offset = int64(result.Offset) + limit
		if offset >= int64(result.Count) {
			break
		}
	}
	return &files, nil
}

// ListPage list files and directories with page
func (c *Pan115Client) ListPage(dirID string, offset, limit int64, opts ...ListOption) (*[]File, error) {
	o := DefaultListOptions()
	if len(opts) > 0 {
		for _, opt := range opts {
			opt(o)
		}
	}

	apiURLs := o.ApiURLs
	var files []File
	req := c.NewRequest().ForceContentType("application/json;charset=UTF-8")
	getFilesOpts := []GetFileOptions{
		WithApiURL(apiURLs[0]),
		WithLimit(limit),
		WithOffset(offset),
	}
	result, err := GetFiles(req, dirID, getFilesOpts...)
	if err != nil {
		return nil, err
	}
	if int64(result.Count) <= offset {
		return &files, nil
	}
	for _, fileInfo := range result.Files {
		files = append(files, *(&File{}).from(&fileInfo))
	}
	return &files, nil
}

func GetFiles(req *resty.Request, dirID string, opts ...GetFileOptions) (*FileListResp, error) {
	if dirID == "" {
		dirID = "0"
	}
	o := DefaultGetFileOptions()
	if len(opts) > 0 {
		for _, opt := range opts {
			opt(o)
		}
	}
	result := FileListResp{}
	params := map[string]string{
		"aid":              "1",
		"cid":              dirID,
		"o":                o.GetOrder(),
		"asc":              o.GetAsc(),
		"offset":           o.GetOffset(),
		"show_dir":         o.GetshowDir(),
		"limit":            o.GetPageSize(),
		"snap":             "0",
		"natsort":          "0",
		"record_open_time": "1",
		"format":           "json",
		"fc_mix":           "0",
	}
	req = req.SetQueryParams(params).
		SetResult(&result)
	resp, err := req.Get(o.GetApiURL())
	if err = CheckErr(err, &result, resp); err != nil {
		return &FileListResp{}, err
	}
	if dirID != string(result.CategoryID) {
		return &FileListResp{}, err
	}
	return &result, err
}

func (c *Pan115Client) DirName2CID(dir string) (*APIGetDirIDResp, error) {
	result := APIGetDirIDResp{}
	dir = strings.TrimPrefix(dir, "/")
	req := c.NewRequest().ForceContentType("application/json;charset=UTF-8")
	req.SetQueryParam("path", dir).SetResult(&result)
	resp, err := req.Get(ApiDirName2CID)
	if err = CheckErr(err, &result, resp); err != nil {
		return nil, err
	}
	return &result, err
}
