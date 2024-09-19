package driver

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	hash "github.com/SheltonZhu/115driver/pkg/crypto"
	cipher "github.com/SheltonZhu/115driver/pkg/crypto/ec115"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/pkg/errors"
)

// GetDigestResult get digest of file or stream
func (c *Pan115Client) GetDigestResult(r io.Reader) (*hash.DigestResult, error) {
	d := hash.DigestResult{}
	return &d, hash.Digest(r, &d)
}

// GetUploadEndpoint get upload endPoint
func (c *Pan115Client) GetUploadEndpoint(endpoint *UploadEndpointResp) error {
	req := c.NewRequest().
		ForceContentType("application/json;charset=UTF-8").
		SetResult(&endpoint)
	_, err := req.Get(ApiGetUploadEndpoint)
	if err != nil {
		return err
	}
	return nil
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

// UploadFastOrByOSS Upload By OSS when unable to rapid upload file
// Deprecated: As of v1.0.22, this function simply calls [RapidUploadOrByOSS].
func (c *Pan115Client) UploadFastOrByOSS(dirID, fileName string, fileSize int64, r io.ReadSeeker) error {
	return c.RapidUploadOrByOSS(dirID, fileName, fileSize, r)
}

// RapidUploadOrByOSS Upload By OSS when unable to rapid upload file
func (c *Pan115Client) RapidUploadOrByOSS(dirID, fileName string, fileSize int64, r io.ReadSeeker) error {
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
	if fastInfo, err = c.RapidUpload(
		digest.Size, fileName, dirID, digest.PreID, digest.QuickID, r,
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

// getOSSEndpoint get oss endpoint 利用阿里云内网上传文件，需要在阿里云服务器上运行本程序，同时也需要115在服务器的所在地域开通了阿里云OSS
func (c *Pan115Client) getOSSEndpoint(enableInternalUpload bool) string {
	if enableInternalUpload {
		uploadEndpoint := UploadEndpointResp{}
		if err := c.GetUploadEndpoint(&uploadEndpoint); err != nil {
			// TODO warn error log
			return OSSEndpoint
		}
		i := strings.Index(uploadEndpoint.Endpoint, ".aliyuncs.com")
		if i > -1 {
			endpoint := uploadEndpoint.Endpoint[:i] + "-internal" + uploadEndpoint.Endpoint[i:]
			return endpoint
		}
	}
	return OSSEndpoint
}

// GetOSSEndpoint get oss endpoint 利用阿里云内网上传文件，需要在阿里云服务器上运行本程序，同时也需要115在服务器的所在地域开通了阿里云OSS
func (c *Pan115Client) GetOSSEndpoint(enableInternalUpload bool) string {
	return c.getOSSEndpoint(enableInternalUpload)
}

// UploadByOSS use aliyun sdk to upload
func (c *Pan115Client) UploadByOSS(params *UploadOSSParams, r io.Reader, dirID string) error {
	ossToken, err := c.GetOSSToken()
	if err != nil {
		return err
	}
	ossClient, err := oss.New(c.getOSSEndpoint(c.UseInternalUpload), ossToken.AccessKeyID, ossToken.AccessKeySecret)
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

// UploadSHA1 upload a sha1, alias of RapidUpload
// Deprecated: As of v1.0.22, this function simply calls [RapidUpload].
func (c *Pan115Client) UploadSHA1(fileSize int64, fileName, dirID, preID, fileID string, r io.ReadSeeker) (*UploadInitResp, error) {
	return c.RapidUpload(fileSize, fileName, dirID, preID, fileID, r)
}

// RapidUpload rapid upload
func (c *Pan115Client) RapidUpload(fileSize int64, fileName, dirID, preID, fileID string, r io.ReadSeeker) (*UploadInitResp, error) {
	var (
		ecdhCipher   *cipher.EcdhCipher
		encrypted    []byte
		decrypted    []byte
		encodedToken string
		err          error
		target       = "U_1_" + dirID
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

	userID := strconv.FormatInt(c.UserID, 10)
	form := url.Values{}
	form.Set("appid", "0")
	form.Set("appversion", appVer)
	form.Set("userid", userID)
	form.Set("filename", fileName)
	form.Set("filesize", fileSizeStr)
	form.Set("fileid", fileID)
	form.Set("target", target)
	form.Set("sig", c.GenerateSignature(fileID, target))

	signKey, signVal := "", ""
	for retry := true; retry; {
		t := NowMilli()

		if encodedToken, err = ecdhCipher.EncodeToken(t.ToInt64()); err != nil {
			return nil, err
		}

		params := map[string]string{
			"k_ec": encodedToken,
		}

		form.Set("t", t.String())
		form.Set("token", c.GenerateToken(fileID, preID, t.String(), fileSizeStr, signKey, signVal))
		if signKey != "" && signVal != "" {
			form.Set("sign_key", signKey)
			form.Set("sign_val", signVal)
		}
		if encrypted, err = ecdhCipher.Encrypt([]byte(form.Encode())); err != nil {
			return nil, err
		}

		req := c.NewRequest().
			SetQueryParams(params).
			SetBody(encrypted).
			SetHeaderVerbatim("Content-Type", "application/x-www-form-urlencoded").
			SetDoNotParseResponse(true)
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
		if err = CheckErr(json.Unmarshal(decrypted, &result), &result, resp); err != nil {
			return nil, err
		}
		if result.Status == 7 {
			// Update signKey & signVal
			signKey = result.SignKey
			signVal, _ = c.UploadDigestRange(r, result.SignCheck)
		} else {
			retry = false
		}
		result.SHA1 = fileID
	}

	return &result, nil
}

const (
	md5Salt = "Qclm8MGWUv59TnrR0XPg"
	appVer  = "27.0.3.7"
)

func (c *Pan115Client) UploadDigestRange(r io.ReadSeeker, rangeSpec string) (result string, err error) {
	var start, end int64
	if _, err = fmt.Sscanf(rangeSpec, "%d-%d", &start, &end); err != nil {
		return
	}
	h := sha1.New()
	_, err = r.Seek(start, io.SeekStart)
	if err != nil {
		return
	}
	if _, err = io.CopyN(h, r, end-start+1); err == nil {
		result = strings.ToUpper(hex.EncodeToString(h.Sum(nil)))
	}

	return
}

func (c *Pan115Client) GenerateSignature(fileID, target string) string {
	sh1hash := sha1.Sum([]byte(strconv.FormatInt(c.UserID, 10) + fileID + target + "0"))
	sigStr := c.Userkey + hex.EncodeToString(sh1hash[:]) + "000000"
	sh1Sig := sha1.Sum([]byte(sigStr))
	return strings.ToUpper(hex.EncodeToString(sh1Sig[:]))
}

func (c *Pan115Client) GenerateToken(fileID, preID, timeStamp, fileSize, signKey, signVal string) string {
	userID := strconv.FormatInt(c.UserID, 10)
	userIDMd5 := md5.Sum([]byte(userID))
	tokenMd5 := md5.Sum([]byte(md5Salt + fileID + fileSize + signKey + signVal + userID + timeStamp + hex.EncodeToString(userIDMd5[:]) + appVer))
	return hex.EncodeToString(tokenMd5[:])
}

// UploadFastOrByMultipart upload by mutipart blocks when unable to rapid upload
// Deprecated: As of v1.0.22, this function simply calls [RapidUploadOrByMultipart].
func (c *Pan115Client) UploadFastOrByMultipart(dirID, fileName string, fileSize int64, r *os.File, opts ...UploadMultipartOption) error {
	return c.RapidUploadOrByMultipart(dirID, fileName, fileSize, r, opts...)
}

// RapidUploadOrByMultipart upload by mutipart blocks when unable to rapid upload
func (c *Pan115Client) RapidUploadOrByMultipart(dirID, fileName string, fileSize int64, r *os.File, opts ...UploadMultipartOption) error {
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
	if fastInfo, err = c.RapidUpload(
		digest.Size, fileName, dirID, digest.PreID, digest.QuickID, r,
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

	options := DefalutUploadMultipartOptions()
	if len(opts) > 0 {
		for _, f := range opts {
			f(options)
		}
	}

	if ossToken, err = c.GetOSSToken(); err != nil {
		return err
	}

	if ossClient, err = oss.New(c.getOSSEndpoint(c.UseInternalUpload), ossToken.AccessKeyID, ossToken.AccessKeySecret); err != nil {
		return err
	}

	if bucket, err = ossClient.Bucket(params.Bucket); err != nil {
		return err
	}

	// ossToken一小时后就会失效，所以每50分钟重新获取一次
	ticker := time.NewTicker(options.TokenRefreshTime)
	defer ticker.Stop()
	// 设置超时
	timeout := time.NewTimer(options.Timeout)

	if chunks, err = SplitFile(f.Name(), fileSize); err != nil {
		return err
	}

	if imur, err = bucket.InitiateMultipartUpload(params.Object,
		oss.SetHeader(OssSecurityTokenHeaderName, ossToken.SecurityToken),
		oss.UserAgentHeader(OSSUserAgent),
	); err != nil {
		return err
	}

	wg := sync.WaitGroup{}
	wg.Add(len(chunks))

	chunksCh := make(chan oss.FileChunk)
	errCh := make(chan error)
	UploadedPartsCh := make(chan oss.UploadPart)
	quit := make(chan struct{})

	// producter
	go chunksProducer(chunksCh, chunks)
	go func() {
		wg.Wait()
		quit <- struct{}{}
	}()

	// consumers
	for i := 0; i < options.ThreadsNum; i++ {
		go func(threadId int) {
			defer func() {
				if r := recover(); r != nil {
					errCh <- fmt.Errorf("recovered in %v", r)
				}
			}()
			for chunk := range chunksCh {
				var part oss.UploadPart // 出现错误就继续尝试，共尝试3次
				for retry := 0; retry < 3; retry++ {
					select {
					case <-ticker.C:
						if ossToken, err = c.GetOSSToken(); err != nil { // 到时重新获取ossToken
							errCh <- errors.Wrap(err, "刷新token时出现错误")
						}
					default:
					}

					buf := make([]byte, chunk.Size)
					if _, err = f.ReadAt(buf, chunk.Offset); err != nil && !errors.Is(err, io.EOF) {
						continue
					}

					b := bytes.NewBuffer(buf)
					if part, err = bucket.UploadPart(imur, b, chunk.Size, chunk.Number, OssOption(params, ossToken)...); err == nil {
						break
					}
				}
				if err != nil {
					errCh <- errors.Wrap(err, fmt.Sprintf("上传 %s 的第%d个分片时出现错误：%v", f.Name(), chunk.Number, err))
				}
				UploadedPartsCh <- part
			}
		}(i)
	}

	go func() {
		for part := range UploadedPartsCh {
			parts = append(parts, part)
			wg.Done()
		}
	}()
LOOP:
	for {
		select {
		case <-ticker.C:
			// 到时重新获取ossToken
			if ossToken, err = c.GetOSSToken(); err != nil {
				return err
			}
		case <-quit:
			break LOOP
		case <-errCh:
			return err
		case <-timeout.C:
			return fmt.Errorf("time out")
		}
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

func chunksProducer(ch chan oss.FileChunk, chunks []oss.FileChunk) {
	for _, chunk := range chunks {
		ch <- chunk
	}
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
