package driver

import (
	"sync"

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
func (c *Pan115Client) List(dirID string) (*[]File, error) {
	return c.ListWithLimit(dirID, FileListLimit)
}

// ListWithMultiThreads list all files and directories more faster
func (c *Pan115Client) ListWithMultiThreads(dirID string, limit int64) (*[]File, error) {
	var files []File
	offset := int64(0)
	req := c.NewRequest().ForceContentType("application/json;charset=UTF-8")
	result, err := GetFiles(req, dirID, WithLimit(limit), WithOffset(offset))
	if err != nil {
		return nil, err
	}

	for _, fileInfo := range result.Files {
		files = append(files, *(&File{}).from(&fileInfo))
	}
	offset = int64(result.Offset) + limit

	if offset >= int64(result.Count) {
		return &files, nil
	}

	remainingCount := (int64(result.Count) - limit)
	jobCount := remainingCount / limit
	lastPageLimit := remainingCount % limit

	if lastPageLimit != 0 {
		jobCount += 1
	}

	var wg sync.WaitGroup
	wg.Add(int(jobCount))

	noOrderFiles := make(map[int]*[]File)

	for i := 0; i < int(jobCount); i++ {
		go func(_offset int64, jobIndex int, done func()) {
			defer done()

			pageFiles, err := c.ListPage(dirID, _offset, limit)
			if err != nil {
				return
			}
			noOrderFiles[jobIndex] = pageFiles

		}(offset+limit*int64(i), i, wg.Done)
	}

	wg.Wait()

	for i := 0; i < int(jobCount); i++ {
		pageFiles, exists := noOrderFiles[i]
		if exists {
			files = append(files, *pageFiles...)
		}
	}

	return &files, nil
}

// ListWithLimit list all files and directories with limit
func (c *Pan115Client) ListWithLimit(dirID string, limit int64) (*[]File, error) {
	var files []File
	offset := int64(0)
	for {
		req := c.NewRequest().ForceContentType("application/json;charset=UTF-8")
		result, err := GetFiles(req, dirID, WithLimit(limit), WithOffset(offset))
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
func (c *Pan115Client) ListPage(dirID string, offset, limit int64) (*[]File, error) {
	var files []File
	req := c.NewRequest().ForceContentType("application/json;charset=UTF-8")
	result, err := GetFiles(req, dirID, WithLimit(limit), WithOffset(offset))
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
	resp, err := req.Get(ApiFileList)
	if err = CheckErr(err, &result, resp); err != nil {
		return &FileListResp{}, err
	}
	if dirID != string(result.CategoryID) {
		return &FileListResp{}, err
	}
	return &result, err
}
