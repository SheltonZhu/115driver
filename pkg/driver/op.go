package driver

import (
	"fmt"
	"strconv"
	"time"
)

// Delete delete files or directory from file ids
func (c *Pan115Client) Delete(fileIDs ...string) error {
	if len(fileIDs) == 0 {
		return nil
	}
	form := map[string]string{}
	for i, value := range fileIDs {
		key := fmt.Sprintf("%s[%d]", "fid", i)
		form[key] = value
	}

	result := BasicResp{}
	req := c.NewRequest().
		SetFormData(form).
		ForceContentType("application/json;charset=UTF-8").
		SetResult(&result)
	resp, err := req.Post(ApiFileDelete)
	return CheckErr(err, &result, resp)
}

// Rename rename a file or directory with file id and name
func (c *Pan115Client) Rename(fileID, newName string) error {
	form := map[string]string{
		"fid":       fileID,
		"file_name": newName,
		fmt.Sprintf("files_new_name[%s]", fileID): newName,
	}

	result := BasicResp{}
	req := c.NewRequest().
		SetFormData(form).
		ForceContentType("application/json;charset=UTF-8").
		SetResult(&result)
	resp, err := req.Post(ApiFileRename)
	return CheckErr(err, &result, resp)
}

// Move move files or directory into another directory with directroy id
func (c *Pan115Client) Move(dirID string, fileIDs ...string) error {
	if len(fileIDs) == 0 {
		return nil
	}
	form := map[string]string{
		"pid": dirID,
	}
	for i, value := range fileIDs {
		key := fmt.Sprintf("%s[%d]", "fid", i)
		form[key] = value
	}
	result := BasicResp{}
	req := c.NewRequest().
		SetFormData(form).
		ForceContentType("application/json;charset=UTF-8").
		SetResult(&result)
	resp, err := req.Post(ApiFileMove)
	return CheckErr(err, &result, resp)
}

// Copy copy files or directory into another directory with directroy id
func (c *Pan115Client) Copy(dirID string, fileIDs ...string) error {
	if len(fileIDs) == 0 {
		return nil
	}
	form := map[string]string{
		"pid": dirID,
	}
	for i, value := range fileIDs {
		key := fmt.Sprintf("%s[%d]", "fid", i)
		form[key] = value
	}
	result := BasicResp{}
	req := c.NewRequest().
		SetFormData(form).
		ForceContentType("application/json;charset=UTF-8").
		SetResult(&result)
	resp, err := req.Post(ApiFileCopy)
	return CheckErr(err, &result, resp)
}

type FileStatInfo struct {
	// Base name of the file.
	Name string
	// Identifier used for downloading or playing the file.
	PickCode string
	// SHA1 hash of file content, in HEX format.
	Sha1 string
	// Marks is file a directory.
	IsDirectory bool
	// Files count under this directory.
	FileCount int
	// Subdirectories count under this directory.
	DirCount int

	// Create time of the file.
	CreateTime time.Time
	// Last update time of the file.
	UpdateTime time.Time
	// Last access time of the file.
	// AccessTime time.Time

	// Parent directory list.
	Parents []*DirInfo
}

// DirInfo only used in FileInfo.
type DirInfo struct {
	// Directory ID.
	ID string
	// Directory Name.
	Name string
}

// Stat get statistic information of a file or directory
func (c *Pan115Client) Stat(fileID string) (*FileStatInfo, error) {
	result := FileStatResponse{}
	req := c.NewRequest().
		SetQueryParam("cid", fileID).
		ForceContentType("application/json;charset=UTF-8").
		SetResult(&result)
	resp, err := req.Get(ApiFileStat)
	if err := CheckErr(err, &result, resp); err != nil {
		return nil, err
	}
	info := &FileStatInfo{
		Name:       result.FileName,
		PickCode:   result.PickCode,
		Sha1:       result.Sha1,
		CreateTime: time.Unix(int64(result.CreateTime), 0),
		UpdateTime: time.Unix(int64(result.UpdateTime), 0),
		// AccessTime: time.Unix(result.AccessTime, 0),
	}
	// Fill parents
	info.Parents = make([]*DirInfo, len(result.Paths))
	for i, path := range result.Paths {
		info.Parents[i] = &DirInfo{
			ID:   strconv.Itoa(path.FileID),
			Name: path.FileName,
		}
	}
	// Directory info
	info.IsDirectory = result.IsFile == 0
	if info.IsDirectory {
		info.FileCount = int(result.FileCount)
		info.DirCount = int(result.FolderCount)
	}
	return info, nil
}

// GetFile gets information of a file or directory by its ID.
func (c *Pan115Client) GetFile(fileID string) (*File, error) {
	result := GetFileInfoResponse{}
	req := c.NewRequest().
		SetQueryParam("file_id", fileID).
		ForceContentType("application/json;charset=UTF-8").
		SetResult(&result)
	resp, err := req.Get(ApiFileInfo)
	if err := CheckErr(err, &result, resp); err != nil {
		return nil, err
	}
	fileInfo := &FileInfo{}
	if len(result.Files) > 0 {
		fileInfo = result.Files[0]
	}
	f := &File{}
	f.from(fileInfo)
	return f, nil
}
