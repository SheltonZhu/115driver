package driver

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"net/url"
	"strconv"
	"strings"

	"io"

	hash "github.com/SheltonZhu/115driver/pkg/crypto"
	"github.com/orzogc/fake115uploader/cipher"
)

func (c *Pan115Client) GetUploadInfo() error {
	result := UploadInfoResp{}
	req := c.NewRequest().
		ForceContentType("application/json;charset=UTF-8").
		SetResult(&result)
	resp, err := req.Post(ApiUploadInfo)
	if err = CheckErr(err, &result, resp); err != nil {
		return err
	}
	c.Userkey = result.Userkey
	c.UserID = result.UserID
	c.UploadMetaInfo = &result.UploadMetaInfo
	return nil
}

func (c *Pan115Client) UploadAvailable() (bool, error) {
	if c.UserID != 0 && len(c.Userkey) > 0 {
		return true, nil
	}
	if err := c.GetUploadInfo(); err != nil {
		return false, err
	}
	return true, nil
}

func (c *Pan115Client) Upload() error {
	return nil
}

const (
	md5Salt = "Qclm8MGWUv59TnrR0XPg"
	appVer  = "30.1.0"
)

func (c *Pan115Client) GenerateSignature(fileID, target string) string {
	sh1hash := sha1.Sum([]byte(strconv.FormatInt(c.UserID, 10) + fileID + fileID + target + "0"))
	sigStr := c.Userkey + hex.EncodeToString(sh1hash[:]) + "000000"
	sh1Sig := sha1.Sum([]byte(sigStr))
	return strings.ToUpper(hex.EncodeToString(sh1Sig[:]))
}

func (c *Pan115Client) GenerateToken(fileID, preID, timeStamp, fileSize string) string {
	userID := strconv.FormatInt(c.UserID, 10)
	userIdMd5 := md5.Sum([]byte(userID))
	tokenMd5 := md5.Sum([]byte(md5Salt + fileID + fileSize + preID + userID + timeStamp + hex.EncodeToString(userIdMd5[:]) + appVer))
	return hex.EncodeToString(tokenMd5[:])
}

func (c *Pan115Client) GetDigestResult(r io.Reader) (*hash.DigestResult, error) {
	d := hash.DigestResult{}
	return &d, hash.Digest(r, &d)
}

func (c *Pan115Client) UploadSH1(fileSize int64, fileName, dirID, preID, fileID string) (*UploadInitResp, error) {
	var (
		ecdhCipher   *cipher.EcdhCipher
		encrypted    []byte
		decrypted    []byte
		encodedToken string
		err          error
		target       = "U_1_" + dirID
		t            = Now()
		bodyBytes    []byte
		result       = UploadInitResp{}
		fileSizeStr  = strconv.FormatInt(fileSize, 10)
	)

	if ecdhCipher, err = cipher.NewEcdhCipher(); err != nil {
		return nil, err
	}

	if ok, err := c.UploadAvailable(); !ok || err != nil {
		return nil, err
	}

	if encodedToken, err = ecdhCipher.EncodeToken(t.ToInt64()); err != nil {
		return nil, err
	}

	params := map[string]string{
		"isp":        strconv.FormatInt(c.UploadMetaInfo.IspType, 10),
		"appid":      strconv.FormatInt(c.UploadMetaInfo.AppID, 10),
		"t":          t.String(),
		"token":      c.GenerateToken(fileID, preID, t.String(), fileSizeStr),
		"appversion": appVer,
		"format":     "json",
		"sig":        c.GenerateSignature(fileID, target),
		"k_ec":       encodedToken,
	}

	userID := strconv.FormatInt(c.UserID, 10)
	form := url.Values{}
	form.Set("preid", preID)
	form.Set("filename", fileName)
	form.Set("quickid", fileID)
	form.Set("user_id", userID)
	form.Set("app_ver", appVer)
	form.Set("filesize", fileSizeStr)
	form.Set("userid", userID)
	form.Set("exif", "")
	form.Set("target", target)
	form.Set("fileid", fileID)

	if encrypted, err = ecdhCipher.Encrypt([]byte(form.Encode())); err != nil {
		return nil, err
	}

	req := c.NewRequest().
		SetQueryParams(params).
		SetBody(encrypted).
		SetHeaderVerbatim("Content-Type", "application/x-www-form-urlencoded").
		SetDoNotParseResponse(true)
	defer func() {
		req.SetDoNotParseResponse(false)
	}()
	resp, err := req.Post(ApiUploadInit)
	if err != nil {
		return nil, err
	}
	data := resp.RawBody()
	defer data.Close()

	if bodyBytes, err = io.ReadAll(data); err != nil {
		return nil, err
	}
	if decrypted, err = ecdhCipher.Decrypt(bodyBytes); err != nil {
		return nil, err
	}
	err = CheckErr(json.Unmarshal(decrypted, &result), &result, resp)
	return &result, err
}
