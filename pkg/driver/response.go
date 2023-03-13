package driver

import (
	"time"
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
	UploadOSSParams

	// Useless fields
	FileID   int    `json:"fileid"`
	FileInfo string `json:"fileinfo"`

	// New fields in upload v4.0
	SignKey   string `json:"sign_key"`
	SignCheck string `json:"sign_check"`
}

type UploadOSSParams struct {
	SHA1     string `json:"-"`
	Bucket   string `json:"bucket"`
	Object   string `json:"object"`
	Callback struct {
		Callback    string `json:"callback"`
		CallbackVar string `json:"callback_var"`
	} `json:"callback"`
}

func (r *UploadInitResp) Err(respBody ...string) error {
	if r.ErrorCode == 0 || r.ErrorCode == 701 {
		return nil
	}
	return GetErr(r.ErrorCode, r.ErrorMsg)
}

// Ok if fastupload is successful will return true, otherwise return false
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

type UploadOSSTokenResp struct {
	AccessKeyID     string    `json:"AccessKeyID"`
	AccessKeySecret string    `json:"AccessKeySecret"`
	Expiration      time.Time `json:"Expiration"`
	SecurityToken   string    `json:"SecurityToken"`
	StatusCode      string    `json:"StatusCode"`
}

func (r *UploadOSSTokenResp) Err(respBody ...string) error {
	if r.StatusCode == "200" {
		return nil
	}
	if len(respBody) > 0 {
		return GetErr(0, respBody[0])
	}
	return ErrUnexpected
}

type DownloadReap struct {
	BasicResp
	EncodedData string `json:"data,omitempty"`
}

type UserInfoResp struct {
	BasicResp
	UserInfo UserInfo `json:"data"`
}
type UserInfo struct {
	Device      int           `json:"device"`
	Rank        int           `json:"rank"`
	Liang       int           `json:"liang"`
	Mark        int           `json:"mark"`
	Mark1       int           `json:"mark1"`
	Vip         int           `json:"vip"`
	Expire      int           `json:"expire"`
	Global      int           `json:"global"`
	Forever     int           `json:"forever"`
	IsPrivilege bool          `json:"is_privilege"`
	Privilege   []interface{} `json:"privilege"`
	UserName    string        `json:"user_name"`
	Face        string        `json:"face"`
	UserID      int64         `json:"user_id"`
}

type FileStatResponse struct {
	FileCount   StringInt         `json:"count"`
	Size        string            `json:"size"`
	FolderCount StringInt         `json:"folder_count"`
	CreateTime  StringInt64       `json:"ptime"`
	UpdateTime  StringInt64       `json:"utime"`
	IsShare     StringInt         `json:"is_share"`
	FileName    string            `json:"file_name"`
	PickCode    string            `json:"pick_code"`
	Sha1        string            `json:"sha1"`
	IsMark      StringInt         `json:"is_mark"`
	OpenTime    int64             `json:"open_time"`
	IsFile      StringInt         `json:"file_category"`
	Paths       []*FileParentInfo `json:"paths"`
}
type FileParentInfo struct {
	FileID   int    `json:"file_id"`
	FileName string `json:"file_name"`
}

func (r *FileStatResponse) Err(respBody ...string) error {
	return nil
}

type GetFileInfoResponse struct {
	BasicResp
	Files []*FileInfo `json:"data"`
}

type QRCodeBasicResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	State   int    `json:"state"`
	Errno   int    `json:"errno"`
	Error   string `json:"error"`
}

func (resp *QRCodeBasicResp) Err(respBody ...string) error {
	if resp.State == 1 {
		return nil
	}
	if len(respBody) > 0 {
		return GetErr(resp.Code, respBody[0])
	}
	return GetErr(resp.Code)
}

type QRCodeTokenResp struct {
	QRCodeBasicResp
	Data QRCodeSession `json:"data"`
}

type QRCodeLoginResp struct {
	QRCodeBasicResp
	Data struct {
		Alert      string     `json:"alert"`
		BindMobile int        `json:"bind_mobile"`
		Credential Credential `json:"cookie"`
		Country    string     `json:"country"`
		Email      string     `json:"email"`
		Face       struct {
			FaceL string `json:"face_l"`
			FaceM string `json:"face_m"`
			FaceS string `json:"face_s"`
		} `json:"face"`
		From          string      `json:"from"`
		IsChangPasswd int         `json:"is_chang_passwd"`
		IsFirstLogin  int         `json:"is_first_login"`
		IsTrusted     interface{} `json:"is_trusted"`
		IsVip         int         `json:"is_vip"`
		Mark          int         `json:"mark"`
		Mobile        string      `json:"mobile"`
		UserID        int         `json:"user_id"`
		UserName      string      `json:"user_name"`
	} `json:"data"`
}

type QRCodeStatusResp struct {
	QRCodeBasicResp
	Data QRCodeStatus `json:"data"`
}
