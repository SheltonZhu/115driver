package driver

import (
	"strconv"

	"github.com/go-resty/resty/v2"
)

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

const (
	FileOrderByTime = "user_ptime"
	FileOrderByType = "file_type"
	FileOrderBySize = "file_size"
	FileOrderByName = "file_name"

	FileListLimit = int64(56)
)

func (c *Pan115Client) List(dirID string) (*[]File, error) {
	var files []File
	offset := int64(0)
	req := c.NewRequest().ForceContentType("application/json;charset=UTF-8")
	for {
		result, err := GetFiles(req, dirID, FileListLimit, offset)
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

func GetFiles(req *resty.Request, dirID string, pageSize, offset int64) (*FileListResponse, error) {
	result := FileListResponse{}
	params := map[string]string{
		"aid":              "1",
		"cid":              dirID,
		"o":                FileOrderByName,
		"asc":              "1",
		"offset":           strconv.FormatInt(offset, 10),
		"show_dir":         "1",
		"limit":            strconv.FormatInt(pageSize, 10),
		"snap":             "0",
		"natsort":          "1",
		"record_open_time": "1",
		"format":           "json",
		"fc_mix":           "0",
	}
	req = req.SetQueryParams(params).SetResult(&result)
	resp, err := req.Get(ApiFileListByName)
	err = CheckErr(err, &result, resp)
	if dirID != string(result.CategoryID) {
		return &FileListResponse{}, nil
	}
	return &result, err
}
