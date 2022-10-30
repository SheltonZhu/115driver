package driver

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	hash "github.com/SheltonZhu/115driver/pkg/crypto"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/orzogc/fake115uploader/cipher"
	"github.com/pkg/errors"
)

// GetDigestResult get digest of file or stream
func (c *Pan115Client) GetDigestResult(r io.Reader) (*hash.DigestResult, error) {
	d := hash.DigestResult{}
	return &d, hash.Digest(r, &d)
}

// GetUploadInfo get some info for upload
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

// UploadAvailable check and prepare to upload
func (c *Pan115Client) UploadAvailable() (bool, error) {
	if c.UserID != 0 && len(c.Userkey) > 0 {
		return true, nil
	}
	if err := c.GetUploadInfo(); err != nil {
		return false, err
	}
	return true, nil
}

// UploadFastOrByOSS check sha1 then upload by oss
func (c *Pan115Client) UploadFastOrByOSS(dirID, fileName string, fileSize int64, r io.ReadSeeker) error {
	var (
		err      error
		digest   *hash.DigestResult
		fastInfo *UploadInitResp
	)

	if ok, err := c.UploadAvailable(); err != nil || !ok {
		return err
	}
	if fileSize > c.UploadMetaInfo.SizeLimit {
		return ErrUploadTooLarge
	}
	if digest, err = c.GetDigestResult(r); err != nil {
		return err
	}
	// 闪传
	if fastInfo, err = c.UploadSHA1(
		digest.Size, fileName, dirID, digest.PreID, digest.QuickID,
	); err != nil {
		return err
	}
	if ok, err := fastInfo.Ok(); err != nil {
		return err
	} else if ok {
		return nil
	}
	if _, err = r.Seek(0, io.SeekStart); err != nil {
		return err
	}
	// 闪传失败，普通上传
	return c.UploadByOSS(&fastInfo.UploadOSSParams, r, dirID)
}

// UploadByOSS use aliyun sdk to upload
func (c *Pan115Client) UploadByOSS(params *UploadOSSParams, r io.Reader, dirID string) error {
	ossToken, err := c.GetOSSToken()
	if err != nil {
		return err
	}
	ossClient, err := oss.New(OSSEndpoint, ossToken.AccessKeyID, ossToken.AccessKeySecret)
	if err != nil {
		return err
	}
	bucket, err := ossClient.Bucket(params.Bucket)
	if err != nil {
		return err
	}

	if err = bucket.PutObject(params.Object, r, OssOption(params, ossToken)...); err != nil {
		return err
	}

	return c.checkUploadStatus(dirID, params.SHA1)
}

func (c *Pan115Client) checkUploadStatus(dirID, sha1 string) error {
	// 验证上传是否成功
	req := c.NewRequest().ForceContentType("application/json;charset=UTF-8")
	opts := []GetFileOptions{
		WithOrder(FileOrderByTime),
		WithShowDirEnable(false),
		WithAsc(false),
		WithLimit(500),
	}
	fResp, err := GetFiles(req, dirID, opts...)
	if err != nil {
		return err
	}
	for _, fileInfo := range fResp.Files {
		if fileInfo.Sha1 == sha1 {
			return nil
		}
	}
	return ErrUploadFailed
}

// GetOSSToken get oss token for oss upload
func (c *Pan115Client) GetOSSToken() (*UploadOSSTokenResp, error) {
	result := UploadOSSTokenResp{}
	req := c.NewRequest().
		ForceContentType("application/json;charset=UTF-8").
		SetResult(&result)

	resp, err := req.Get(ApiUploadOSSToken)
	return &result, CheckErr(err, &result, resp)
}

// UploadSHA1 upload a sha1 for upload
func (c *Pan115Client) UploadSHA1(fileSize int64, fileName, dirID, preID, fileID string) (*UploadInitResp, error) {
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
	result.SHA1 = fileID
	return &result, err
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
	userIDMd5 := md5.Sum([]byte(userID))
	tokenMd5 := md5.Sum([]byte(md5Salt + fileID + fileSize + preID + userID + timeStamp + hex.EncodeToString(userIDMd5[:]) + appVer))
	return hex.EncodeToString(tokenMd5[:])
}

// UploadFastOrByMultipart  check sha1 then upload by mutipart blocks
func (c *Pan115Client) UploadFastOrByMultipart(dirID, fileName string, fileSize int64, r *os.File, opts ...UploadMultipartOption) error {
	var (
		err      error
		digest   *hash.DigestResult
		fastInfo *UploadInitResp
	)

	if ok, err := c.UploadAvailable(); err != nil || !ok {
		return err
	}
	if fileSize > c.UploadMetaInfo.SizeLimit {
		return ErrUploadTooLarge
	}
	if digest, err = c.GetDigestResult(r); err != nil {
		return err
	}
	// 闪传
	if fastInfo, err = c.UploadSHA1(
		digest.Size, fileName, dirID, digest.PreID, digest.QuickID,
	); err != nil {
		return err
	}
	if ok, err := fastInfo.Ok(); err != nil {
		return err
	} else if ok {
		return nil
	}
	if _, err = r.Seek(0, io.SeekStart); err != nil {
		return err
	}

	// 闪传失败，上传
	if digest.Size <= KB { // 文件大小小于1KB，改用普通模式上传
		return c.UploadByOSS(&fastInfo.UploadOSSParams, r, dirID)
	}
	// 分片上传
	return c.UploadByMultipart(&fastInfo.UploadOSSParams, digest.Size, r, dirID, opts...)
}

// UploadByMultipart upload by mutipart blocks
func (c *Pan115Client) UploadByMultipart(params *UploadOSSParams, fileSize int64, f *os.File, dirID string, opts ...UploadMultipartOption) error {
	var (
		chunks    []oss.FileChunk
		parts     []oss.UploadPart
		imur      oss.InitiateMultipartUploadResult
		ossClient *oss.Client
		bucket    *oss.Bucket
		ossToken  *UploadOSSTokenResp
		err       error
	)

	options := UploadMultipartOptions{10}
	if len(opts) > 0 {
		for _, f := range opts {
			f(options)
		}
	}

	if ossToken, err = c.GetOSSToken(); err != nil {
		return err
	}

	if ossClient, err = oss.New(OSSEndpoint, ossToken.AccessKeyID, ossToken.AccessKeySecret); err != nil {
		return err
	}

	if bucket, err = ossClient.Bucket(params.Bucket); err != nil {
		return err
	}

	// ossToken一小时后就会失效，所以每50分钟重新获取一次
	ticker := time.NewTicker(50 * time.Minute)
	defer ticker.Stop()

	if chunks, err = SplitFile(f.Name(), fileSize); err != nil {
		return err
	}

	if imur, err = bucket.InitiateMultipartUpload(params.Object,
		oss.SetHeader(OssSecurityTokenHeaderName, ossToken.SecurityToken),
		oss.UserAgentHeader(OSSUserAgent),
	); err != nil {
		return err
	}

	for _, chunk := range chunks {
		var part oss.UploadPart // 出现错误就继续尝试，共尝试3次
		for retry := 0; retry < 3; retry++ {
			select {
			case <-ticker.C:
				if ossToken, err = c.GetOSSToken(); err != nil { // 到时重新获取ossToken
					log.Printf("刷新token时出现错误：%v", err)
				}
			default:
			}

			_, _ = f.Seek(chunk.Offset, io.SeekStart)
			if part, err = bucket.UploadPart(imur, f, chunk.Size, chunk.Number, OssOption(params, ossToken)...); err == nil {
				break
			}
		}
		if err != nil {
			log.Printf("上传 %s 的第%d个分片时出现错误：%v", f.Name(), chunk.Number, err)
		}
		parts = append(parts, part)
	}

	select {
	case <-ticker.C:
		// 到时重新获取ossToken
		if ossToken, err = c.GetOSSToken(); err != nil {
			return err
		}
	default:
	}

	// EOF错误是xml的Unmarshal导致的，响应其实是json格式，所以实际上上传是成功的
	if _, err = bucket.CompleteMultipartUpload(imur, parts, OssOption(params, ossToken)...); err != nil && !errors.Is(err, io.EOF) {
		// 当文件名含有 &< 这两个字符之一时响应的xml解析会出现错误，实际上上传是成功的
		if filename := filepath.Base(f.Name()); !strings.ContainsAny(filename, "&<") {
			return err
		}
	}
	return c.checkUploadStatus(dirID, params.SHA1)
}

// SplitFile pplitFile
func SplitFile(filePath string, fileSize int64) (chunks []oss.FileChunk, err error) {
	for i := int64(1); i < 10; i++ {
		if fileSize < i*GB { // 文件大小小于iGB时分为i*1000片
			if chunks, err = oss.SplitFileByPartNum(filePath, int(i*1000)); err != nil {
				return
			}
			break
		}
	}
	if fileSize > 9*GB { // 文件大小大于9GB时分为10000片
		if chunks, err = oss.SplitFileByPartNum(filePath, 10000); err != nil {
			return
		}
	}
	// 单个分片大小不能小于100KB
	if chunks[0].Size < 100*KB {
		if chunks, err = oss.SplitFileByPartSize(filePath, 100*KB); err != nil {
			return
		}
	}
	return
}

// OssOption get options
func OssOption(params *UploadOSSParams, ossToken *UploadOSSTokenResp) []oss.Option {
	options := []oss.Option{
		oss.SetHeader(OssSecurityTokenHeaderName, ossToken.SecurityToken),
		oss.Callback(base64.StdEncoding.EncodeToString([]byte(params.Callback.Callback))),
		oss.CallbackVar(base64.StdEncoding.EncodeToString([]byte(params.Callback.CallbackVar))),
		oss.UserAgentHeader(OSSUserAgent),
	}
	return options
}
