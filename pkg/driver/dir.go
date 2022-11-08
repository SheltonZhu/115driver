package driver

import (
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
	return result.CategoryID, nil
}

// List list all files and directories
func (c *Pan115Client) List(dirID string) (*[]File, error) {
	var files []File
	offset := int64(0)
	for {
		req := c.NewRequest().ForceContentType("application/json;charset=UTF-8")
		result, err := GetFiles(req, dirID, WithLimit(FileListLimit), WithOffset(offset))
		if err != nil {
			return nil, err
		}
		for _, fileInfo := range result.Files {
			files = append(files, *(&File{}).from(&fileInfo))
		}
		offset = int64(result.Offset) + FileListLimit
		if offset >= int64(result.Count) {
			break
		}
	}
	return &files, nil
}

func GetFiles(req *resty.Request, dirID string, opts ...GetFileOptions) (*FileListResp, error) {
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
		"natsort":          "1",
		"record_open_time": "1",
		"format":           "json",
		"fc_mix":           "0",
	}
	req = req.SetQueryParams(params).
		SetResult(&result)
	resp, err := req.Get(ApiFileListByName)
	if err = CheckErr(err, &result, resp); err != nil {
		return &FileListResp{}, err
	}
	if dirID != string(result.CategoryID) {
		return &FileListResp{}, err
	}
	return &result, err
}
