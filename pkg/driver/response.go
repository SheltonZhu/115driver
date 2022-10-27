package driver

type LoginResp struct {
	Code     int  `json:"code"`
	CheckSsd bool `json:"check_ssd"`
	Data     struct {
		Expire int    `json:"expire"`
		Link   string `json:"link"`
		UserID int    `json:"user_id"`
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

type BasicResponse struct {
	Errno   StringInt `json:"errno,omitempty"`
	ErrNo   int       `json:"errNo,omitempty"`
	Error   string    `json:"error,omitempty"`
	State   bool      `json:"state,omitempty"`
	Errtype string    `json:"errtype,omitempty"`
}

func (resp *BasicResponse) Err(respBody ...string) error {
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
	BasicResponse
	AreaID IntString `json:"aid"`

	CategoryID   string `json:"cid"`
	CategoryName string `json:"cname"`

	FileID   string `json:"file_id"`
	FileName string `json:"file_name"`
}

type FileListResponse struct {
	BasicResponse

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
