package driver

import (
	"fmt"
	"strconv"

	"github.com/go-resty/resty/v2"
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

// QRCodeByApi get QRCode matrix or image by api.
func (s *QRCodeSession) QRCodeByApi() ([]byte, error) {
	resp, err := resty.New().R().Get(fmt.Sprintf(ApiQrcodeImage, s.UID))
	return resp.Body(), err
}

// QRCodeStart starts a QRCode login session.
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

// QRCodeLogin logins user through QRCode with web app.
// You SHOULD call this method ONLY when `QRCodeStatus.IsAllowed()` is true.
func (c *Pan115Client) QRCodeLogin(s *QRCodeSession) (*Credential, error) {
	return c.QRCodeLoginWithApp(s, LoginAppWeb)
}

type LoginApp string

const (
	LoginAppWeb     LoginApp = "web"
	LoginAppAndroid LoginApp = "android"
	LoginAppIOS     LoginApp = "ios"
	// LoginAppLinux      LoginApp = "linux"   // disabled
	// LoginAppMac        LoginApp = "mac"     // disabled
	// LoginAppWindows    LoginApp = "windows" // disabled
	LoginAppTV         LoginApp = "tv"
	LoginAppAlipayMini LoginApp = "alipaymini"
	LoginAppWechatMini LoginApp = "wechatmini"
	LoginQAppAndroid   LoginApp = "qandroid"
)

// QRCodeLoginWithApp logins user through QRCode with specified app.
// You SHOULD call this method ONLY when `QRCodeStatus.IsAllowed()` is true.
func (c *Pan115Client) QRCodeLoginWithApp(s *QRCodeSession, app LoginApp) (*Credential, error) {
	result := QRCodeLoginResp{}
	req := c.NewRequest().
		SetFormData(map[string]string{
			"account": s.UID,
			"app":     string(app),
		}).
		ForceContentType("application/json;charset=UTF-8").
		SetResult(&result)
	resp, err := req.Post(fmt.Sprintf(ApiQrcodeLoginWithApp, app))
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

func (s *QRCodeStatus) IsExpired() bool {
	return s.Status == -1
}

func (s *QRCodeStatus) IsCanceled() bool {
	return s.Status == -2
}

/*
QRCodeStatus represents the status of a QRCode session.

There are 4 possible status values:
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
