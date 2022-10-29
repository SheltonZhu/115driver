package driver

import (
	"fmt"
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

// Copy files or directory into another directory with directroy id
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
