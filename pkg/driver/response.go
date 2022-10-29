package driver

import (
	"time"

	"github.com/pkg/errors"
)

type LoginResp struct {
	Code     int  `json:"code"`
	CheckSsd bool `json:"check_ssd"`
	Data     struct {
		Expire int    `json:"expire"`
		Link   string `json:"link"`
		UserID int64  `json:"user_id"`
	} `json:"data"`
	Errno   int    `json:"errno"`
	Error   string `json:"error"`
	Message string `json:"message"`
	State   int    `json:"state"`
	Expire  int    `json:"expire"`
}

func (resp *LoginResp) Err(respBody ...string) error {
	if resp.State == 0 {
		return nil
	}
	if len(respBody) > 0 {
		return GetErr(resp.Code, respBody[0])
	}
	return GetErr(resp.Code)
}

type BasicResp struct {
	Errno   StringInt `json:"errno,omitempty"`
	ErrNo   int       `json:"errNo,omitempty"`
	Error   string    `json:"error,omitempty"`
	State   bool      `json:"state,omitempty"`
	Errtype string    `json:"errtype,omitempty"`
	Msg     string    `json:"msg,omitempty"`
}

func (resp *BasicResp) Err(respBody ...string) error {
	if resp.State {
		return nil
	}
	nonZeroCode := findNonZero(int(resp.Errno), resp.ErrNo)
	if len(respBody) > 0 {
		return GetErr(nonZeroCode, respBody[0])
	}
	return GetErr(nonZeroCode)
}

func findNonZero(code ...int) int {
	for _, c := range code {
		if c != 0 {
			return c
		}
	}
	return 0
}

type MkdirResp struct {
	BasicResp
	AreaID IntString `json:"aid"`

	CategoryID   string `json:"cid"`
	CategoryName string `json:"cname"`

	FileID   string `json:"file_id"`
	FileName string `json:"file_name"`
}

type FileListResp struct {
	BasicResp

	AreaID     string    `json:"aid"`
	CategoryID IntString `json:"cid"`

	Count int    `json:"count"`
	Order string `json:"order"`
	IsAsc int    `json:"is_asc"`

	Offset   int `json:"offset"`
	Limit    int `json:"limit"`
	PageSize int `json:"page_size"`

	Files []FileInfo `json:"data"`
}

type FileInfo struct {
	AreaID     IntString `json:"aid"`
	CategoryID string    `json:"cid"`
	FileID     string    `json:"fid"`
	ParentID   string    `json:"pid"`

	Name     string      `json:"n"`
	Type     string      `json:"ico"`
	Size     StringInt64 `json:"s"`
	Sha1     string      `json:"sha"`
	PickCode string      `json:"pc"`

	IsStar StringInt    `json:"m"`
	Labels []*LabelInfo `json:"fl"`

	CreateTime StringInt64 `json:"tp"`
	UpdateTime string      `json:"t"`
}
type LabelInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`

	Sort StringInt `json:"sort"`

	CreateTime int64 `json:"create_time"`
	UpdateTime int64 `json:"update_time"`
}
type UploadInfoResp struct {
	BasicResp
	UploadMetaInfo
	UserID  int64  `json:"user_id"`
	Userkey string `json:"userkey"`
}

type UploadMetaInfo struct {
	AppID            int64    `json:"app_id"`
	AppVersion       int64    `json:"app_version"`
	IspType          int64    `json:"isp_type"`
	MaxDirLevel      int64    `json:"max_dir_level"`
	MaxDirLevelYun   int64    `json:"max_dir_level_yun"`
	MaxFileNum       int64    `json:"max_file_num"`
	MaxFileNumYun    int64    `json:"max_file_num_yun"`
	SizeLimit        int64    `json:"size_limit"`
	SizeLimitYun     int64    `json:"size_limit_yun"`
	TypeLimit        []string `json:"type_limit"`
	UploadAllowed    bool     `json:"upload_allowed"`
	UploadAllowedMsg string   `json:"upload_allowed_msg"`
}

type UploadInitResp struct {
	Request   string `json:"request"`
	ErrorCode int    `json:"statuscode"`
	ErrorMsg  string `json:"statusmsg"`

	Status   BoolInt `json:"status"`
	PickCode string  `json:"pickcode"`
	Target   string  `json:"target"`
	Version  string  `json:"version"`

	// OSS upload fields
	UploadOssParams

	// Useless fields
	FileId   int    `json:"fileid"`
	FileInfo string `json:"fileinfo"`
}

type UploadOssParams struct {
	SHA1     string `json:"-"`
	Bucket   string `json:"bucket"`
	Object   string `json:"object"`
	Callback struct {
		Callback    string `json:"callback"`
		CallbackVar string `json:"callback_var"`
	} `json:"callback"`
}

func (r *UploadInitResp) Err(respBody ...string) error {
	if r.ErrorCode == 0 {
		return nil
	}
	return GetErr(r.ErrorCode, r.ErrorMsg)
}

func (r *UploadInitResp) Ok() (bool, error) {
	switch r.Status {
	case 2:
		return true, nil
	case 1:
		return false, nil
	default:
		return false, ErrUnexpected
	}
}

type UploadOssTokenResponse struct {
	AccessKeyID     string    `json:"AccessKeyId"`
	AccessKeySecret string    `json:"AccessKeySecret"`
	Expiration      time.Time `json:"Expiration"`
	SecurityToken   string    `json:"SecurityToken"`
	StatusCode      string    `json:"StatusCode"`
}

func (r *UploadOssTokenResponse) Err(respBody ...string) error {
	if r.StatusCode == "200" {
		return nil
	}
	if len(respBody) > 0 {
		return errors.Wrap(ErrUnexpected, respBody[0])
	}
	return ErrUnexpected
}
