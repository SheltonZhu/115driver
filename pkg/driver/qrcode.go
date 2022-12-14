package driver

import (
	"strconv"

	qrcode "github.com/skip2/go-qrcode"
)

type QRCodeSession struct {
	// The raw data of QRCode, caller should use third-party tools/libraries
	// to convert it into QRCode matrix or image.
	QrcodeContent string `json:"qrcode"`
	Sign          string `json:"sign"`
	Time          int64  `json:"time"`
	UID           string `json:"uid"`
}

// QRCode get QRCode matrix or image.
func (s *QRCodeSession) QRCode() ([]byte, error) {
	return qrcode.Encode(s.QrcodeContent, qrcode.Medium, 256)
}

// QrcodeStart starts a QRCode login session.
func (c *Pan115Client) QRCodeStart() (*QRCodeSession, error) {
	result := QRCodeTokenResp{}
	resp, err := c.NewRequest().
		SetResult(&result).
		ForceContentType("application/json;charset=UTF-8").
		Get(ApiQrcodeToken)

	if err = CheckErr(err, &result, resp); err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// QrcodeLogin logins user through QRCode.
// You SHOULD call this method ONLY when `QRCodeStatus.IsAllowed()` is true.
func (c *Pan115Client) QRCodeLogin(s *QRCodeSession) (*Credential, error) {
	result := QRCodeLoginResp{}
	req := c.NewRequest().
		SetFormData(map[string]string{
			"account": s.UID,
			"app":     "web",
		}).
		ForceContentType("application/json;charset=UTF-8").
		SetResult(&result)
	resp, err := req.Post(ApiQrcodeLogin)
	if err = CheckErr(err, &result, resp); err != nil {
		return nil, err
	}

	return &result.Data.Credential, nil
}

type QRCodeStatus struct {
	Msg     string `json:"msg"`
	Status  int    `json:"status"`
	Version string `json:"version"`
}

func (s *QRCodeStatus) IsWaiting() bool {
	return s.Status == 0
}

func (s *QRCodeStatus) IsScanned() bool {
	return s.Status == 1
}

func (s *QRCodeStatus) IsAllowed() bool {
	return s.Status == 2
}

func (s *QRCodeStatus) IsCanceled() bool {
	return s.Status == -2
}

/*
	QRCodeStatus

There will be 4 kinds of status:
  - Waiting
  - Scanned
  - Allowed
  - Canceled
*/
func (c *Pan115Client) QRCodeStatus(s *QRCodeSession) (*QRCodeStatus, error) {
	result := QRCodeStatusResp{}
	req := c.NewRequest().
		SetQueryParams(map[string]string{
			"uid":  s.UID,
			"time": strconv.FormatInt(s.Time, 10),
			"sign": s.Sign,
			"_":    Now().String(),
		}).
		ForceContentType("application/json;charset=UTF-8").
		SetResult(&result)

	resp, err := req.Get(ApiQrcodeStatus)
	if err = CheckErr(err, &result, resp); err != nil {
		return nil, err
	}
	return &result.Data, nil
}
