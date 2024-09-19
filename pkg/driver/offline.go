package driver

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	crypto "github.com/SheltonZhu/115driver/pkg/crypto/m115"
)

// OfflineTask describe an offline downloading task.
type OfflineTask struct {
	InfoHash     string  `json:"info_hash"`
	Name         string  `json:"name"`
	Size         int64   `json:"size"`
	Url          string  `json:"url"`
	AddTime      int64   `json:"add_time"`
	Peers        int64   `json:"peers"`
	RateDownload float64 `json:"rateDownload"`
	Status       int     `json:"status"`
	Percent      float64 `json:"percentDone"`
	UpdateTime   int64   `json:"last_update"`
	LeftTime     int64   `json:"left_time"`
	FileId       string  `json:"file_id"`
	DelFileId    string  `json:"delete_file_id"`
	DirId        string  `json:"wp_path_id"`
	Move         int     `json:"move"`
}

func (t *OfflineTask) IsTodo() bool {
	return t.Status == 0
}

func (t *OfflineTask) IsRunning() bool {
	return t.Status == 1
}

func (t *OfflineTask) IsDone() bool {
	return t.Status == 2
}

func (t *OfflineTask) IsFailed() bool {
	return t.Status == -1
}

func (t *OfflineTask) GetStatus() string {
	if t.IsTodo() {
		return "准备开始离线下载"
	}
	if t.IsDone() {
		return "离线下载完成"
	}
	if t.IsFailed() {
		return "离线下载失败"
	}
	if t.IsRunning() {
		return "离线任务下载中"
	}
	return fmt.Sprintf("未知状态: %d", t.Status)
}

// ListOfflineTask list tasks
func (c *Pan115Client) ListOfflineTask(page int64) (OfflineTaskResp, error) {
	result := OfflineTaskResp{}
	req := c.NewRequest().
		SetQueryParam("page", strconv.FormatInt(page, 10)).
		SetResult(&result).
		ForceContentType("application/json;charset=UTF-8")

	resp, err := req.Post(ApiListOfflineUrl)

	if err := CheckErr(err, &result, resp); err != nil {
		return OfflineTaskResp{}, err
	}
	return result, nil
}

// AddOfflineTaskURIs adds offline tasks by download URIs.
// supports http, ed2k, magent
func (c *Pan115Client) AddOfflineTaskURIs(uris []string, saveDirID string) (hashes []string, err error) {
	count := len(uris)
	if count == 0 {
		return
	}

	if c.UserID < 0 {
		if err := c.LoginCheck(); err != nil {
			return nil, err
		}
	}

	key := crypto.GenerateKey()

	result := DownloadResp{}

	params := map[string]string{
		"ac":         "add_task_urls",
		"wp_path_id": saveDirID,
		"app_ver":    appVer,
		"uid":        strconv.FormatInt(c.UserID, 10),
	}
	for i, uri := range uris {
		key := fmt.Sprintf("url[%d]", i)
		params[key] = uri
	}
	paramsBytes, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	data := crypto.Encode(paramsBytes, key)
	req := c.NewRequest().
		SetQueryParam("t", Now().String()).
		SetFormData(map[string]string{"data": data}).
		ForceContentType("application/json").
		SetResult(&result)

	resp, err := req.Post(ApiAddOfflineUrl)

	if err := CheckErr(err, &result, resp); err != nil {
		return nil, err
	}

	bytes, err := crypto.Decode(string(result.EncodedData), key)
	if err != nil {
		return nil, err
	}

	taskInfos := OfflineAddUrlResponse{}
	if err := json.Unmarshal(bytes, &taskInfos); err != nil {
		return nil, err
	}

	hashes = make([]string, count)
	for i, task := range taskInfos.Result {
		hashes[i] = task.InfoHash
	}
	return hashes, nil
}

// DeleteOfflineTasks deletes tasks.
func (c *Pan115Client) DeleteOfflineTasks(hashes []string, deleteFiles bool) error {
	form := url.Values{}
	for _, hash := range hashes {
		form.Add("hash", hash)
	}

	form.Set("flag", "0")
	if deleteFiles {
		form.Set("flag", "1")
	}

	result := MkdirResp{}
	req := c.NewRequest().
		SetFormDataFromValues(form).
		SetResult(&result).
		ForceContentType("application/json;charset=UTF-8")

	resp, err := req.Post(ApiDelOfflineUrl)
	return CheckErr(err, &result, resp)
}

// ClearOfflineTasks deletes tasks.
func (c *Pan115Client) ClearOfflineTasks(clearFlag int64) error {
	form := url.Values{}
	form.Set("flag", strconv.FormatInt(int64(clearFlag), 10))

	result := MkdirResp{}
	req := c.NewRequest().
		SetFormDataFromValues(form).
		SetResult(&result).
		ForceContentType("application/json;charset=UTF-8")

	resp, err := req.Post(ApiClearOfflineUrl)
	return CheckErr(err, &result, resp)
}

type OfflineAddUrlResponse struct {
	BasicResp
	Result []OfflineTaskResponse `json:"result"`
}
type OfflineTaskResponse struct {
	InfoHash string `json:"info_hash"`
	Url      string `json:"url"`
}

type OfflineTaskResp struct {
	BasicResp
	Total     int64          `json:"total"`
	Count     int64          `json:"count"`
	PageRow   int64          `json:"page_row"`
	PageCount int64          `json:"page_count"`
	Page      int64          `json:"page"`
	Quota     int64          `json:"quota"`
	Tasks     []*OfflineTask `json:"tasks"`
}
