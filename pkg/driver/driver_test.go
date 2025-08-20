package driver

import (
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	cookieStr = ""
	client    *Pan115Client
)

func TestMain(m *testing.M) {
	cookieStr = os.Getenv("COOKIE")
	os.Exit(m.Run())
}

func TestImportFromCookie(t *testing.T) {
	cr := &Credential{}
	assert.Nil(t, cr.FromCookie("UID=1;CID=2;SEID=3;KID=12;other=4"))
	assert.Error(t, ErrBadCookie, cr.FromCookie(""))
	assert.Error(t, ErrBadCookie, cr.FromCookie("k=a;;"))
	assert.Error(t, ErrBadCookie, cr.FromCookie("1=2;2=3;3=4"))
	assert.Error(t, ErrBadCookie, cr.FromCookie("1=2;2=3;3=4"))
}

func TestLoginErr(t *testing.T) {
	assert.Error(t, New().ImportCredential(&Credential{}).LoginCheck())
}

func TestBadCookie(t *testing.T) {
	assert.Error(t, New().ImportCredential(&Credential{}).CookieCheck())
}

func teardown(t *testing.T) func(t *testing.T) {
	cr := &Credential{}
	assert.Nil(t, cr.FromCookie(cookieStr))
	client = New(UA(UA115Browser), WithDebug(), WithTrace()).ImportCredential(cr)
	assert.Nil(t, client.CookieCheck())
	return func(t *testing.T) {}
}

func TestListRecycleBin(t *testing.T) {
	down := teardown(t)
	defer down(t)
	_, err := client.ListRecycleBin(0, 40)
	assert.Nil(t, err)
}

func TestCleanRecycleBin(t *testing.T) {
	down := teardown(t)
	defer down(t)
	err := client.CleanRecycleBin("xx", "1", "2")
	assert.NotNil(t, err)
}

func TestListOfflineTasks(t *testing.T) {
	down := teardown(t)
	defer down(t)
	_, err := client.ListOfflineTask(1)
	assert.Nil(t, err)
}

func TestRevertRecycleBin(t *testing.T) {
	down := teardown(t)
	defer down(t)
	err := client.RevertRecycleBin("xx", "1", "2")
	assert.NotNil(t, err)
}

func TestOfflineAddUri(t *testing.T) {
	down := teardown(t)
	defer down(t)

	uri := "https://x.com/Olympics/status/1820550228640203065/photo/1"
	hashs, err := client.AddOfflineTaskURIs([]string{uri}, "0")
	assert.Nil(t, err)
	assert.NotEmpty(t, hashs)
}

func TestOfflineDelUri(t *testing.T) {
	down := teardown(t)
	defer down(t)

	err := client.DeleteOfflineTasks([]string{"1123", "1231"}, true)
	assert.Nil(t, err)
}

func TestOfflineClearUri(t *testing.T) {
	down := teardown(t)
	defer down(t)

	err := client.ClearOfflineTasks(1)
	assert.Nil(t, err)
}

func TestMkdir(t *testing.T) {
	down := teardown(t)
	defer down(t)

	dirName := NowMilli().String()
	_, err := client.Mkdir("0", dirName)
	assert.Nil(t, err)
	_, err = client.Mkdir("0", dirName)
	assert.ErrorIs(t, err, ErrExist)
}

func TestDelete(t *testing.T) {
	down := teardown(t)
	defer down(t)

	dirName := NowMilli().String()
	assert.Error(t, client.Delete(dirName))
}

func TestRename(t *testing.T) {
	down := teardown(t)
	defer down(t)

	dirName := NowMilli().String()
	assert.Nil(t, client.Rename(dirName, "not Exist"))
}

func TestCopy(t *testing.T) {
	down := teardown(t)
	defer down(t)

	dirName := NowMilli().String()
	assert.Error(t, client.Copy("0", dirName))
}

func TestMove(t *testing.T) {
	down := teardown(t)
	defer down(t)

	dirName := NowMilli().String()
	assert.Error(t, client.Move("0", dirName))
}

func TestList(t *testing.T) {
	down := teardown(t)
	defer down(t)

	f1, err := client.List("0", WithApiURLs(ApiFileList))
	assert.NotEmpty(t, *f1)
	assert.Nil(t, err)
	f2, err := client.List("0", WithApiURLs(ApiFileList1))
	assert.NotEmpty(t, *f2)
	assert.Nil(t, err)
	// f3, err := client.List("0", WithApiURLs(ApiFileList2))
	// assert.NotEmpty(t, *f3)
	// assert.Nil(t, err)
	// f4, err := client.List("0", WithApiURLs(ApiFileList3))
	// assert.NotEmpty(t, *f4)
	// assert.Nil(t, err)

	assert.Equal(t, *f1, *f2)
	// assert.Equal(t, *f1, *f3)
	// assert.Equal(t, *f1, *f4)
	dirName := NowMilli().String()
	f, err := client.List(dirName)
	assert.Nil(t, err)
	assert.Empty(t, *f)
}

func TestDirName2CID(t *testing.T) {
	down := teardown(t)
	defer down(t)

	cid, err := client.DirName2CID("Pay")
	assert.NotEmpty(t, cid)
	assert.Nil(t, err)
}

func TestListPage(t *testing.T) {
	down := teardown(t)
	defer down(t)

	f, err := client.ListPage("0", 0, 5)
	assert.NotEmpty(t, *f)
	assert.Nil(t, err)
}

func TestDownload(t *testing.T) {
	down := teardown(t)
	defer down(t)

	pickCode := NowMilli().String()
	_, err := client.Download(pickCode)
	assert.ErrorIs(t, err, ErrPickCodeNotExist)
	_, err = client.Download("")
	assert.ErrorIs(t, err, ErrPickCodeIsEmpty)
}

func TestDownloadByShareCode(t *testing.T) {
	down := teardown(t)
	defer down(t)

	_, err := client.DownloadByShareCode("sw6pw793wfp", "w816", "2628478209787264315")
	assert.ErrorIs(t, err, ErrSharedNotFound)
}

func TestGetUploadInfo(t *testing.T) {
	down := teardown(t)
	defer down(t)

	assert.Nil(t, client.GetUploadInfo())
}

func TestGetUPloadEndpoint(t *testing.T) {
	result := UploadEndpointResp{}
	assert.NoError(t, New().GetUploadEndpoint(&result))
	assert.NotEmpty(t, result)
}

func TestUploadSHA1(t *testing.T) {
	down := teardown(t)
	defer down(t)

	r := strings.NewReader(NowMilli().String())
	d, err := client.GetDigestResult(r)
	assert.Nil(t, err)
	_, err = client.UploadSHA1(d.Size, "xxxa.txt", "0", d.PreID, d.QuickID, r)
	assert.Nil(t, err)
}

func TestGetOSSToken(t *testing.T) {
	down := teardown(t)
	defer down(t)

	token, err := client.GetOSSToken()
	assert.Nil(t, err)
	_ = token
}

func TestUploadByOSS(t *testing.T) {
	down := teardown(t)
	defer down(t)

	randStr := NowMilli().String()
	r := strings.NewReader(randStr)
	d, err := client.GetDigestResult(r)
	assert.Nil(t, err)
	_, err = r.Seek(0, io.SeekStart)
	assert.Nil(t, err)
	resp, err := client.UploadSHA1(d.Size, randStr+".txt", "0", d.PreID, d.QuickID, r)
	assert.Nil(t, err)
	ok, err := resp.Ok()
	assert.Nil(t, err)
	assert.False(t, ok)
	assert.Nil(t, client.UploadByOSS(&resp.UploadOSSParams, r, "0"))
}

func TestUpload(t *testing.T) {
	down := teardown(t)
	defer down(t)

	randStr := NowMilli().String()
	r := strings.NewReader(randStr)
	assert.Nil(t, client.UploadFastOrByOSS("0", randStr+".txt", r.Size(), r))
}

func TestUploadMultipart(t *testing.T) {
	start := time.Now()

	down := teardown(t)
	defer down(t)

	f, err := os.CreateTemp("./", "test-temp-*")
	assert.Nil(t, err)

	randStr := NowMilli().String()
	targetSize := 2048 // 目标文件大小，单位字节，这里设置为2KB
	currentSize := 0
	for currentSize < targetSize {
		// 获取当前时间的毫秒时间戳字符串作为随机内容的一部分
		n, err := f.WriteString(randStr)
		assert.Nil(t, err)
		currentSize += n
	}

	_, err = f.Seek(0, io.SeekStart)
	assert.Nil(t, err)

	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()

	fs, _ := f.Stat()
	err = client.UploadFastOrByMultipart("0", randStr+".txt", fs.Size(), f)
	assert.Nil(t, err)
	elapsedTime := time.Since(start) / time.Millisecond // duration in ms
	t.Logf("Segment finished in %dms", elapsedTime)
}

func TestGetUser(t *testing.T) {
	down := teardown(t)
	defer down(t)

	_, err := client.GetUser()
	assert.Nil(t, err)
}

func TestStat(t *testing.T) {
	down := teardown(t)
	defer down(t)

	_, err := client.Stat("fileID")
	assert.Error(t, err)
}

func TestGet(t *testing.T) {
	down := teardown(t)
	defer down(t)

	_, err := client.GetFile("")
	assert.Error(t, err)
}

func TestQRCodeStart(t *testing.T) {
	t.Skip()
	c := New(WithTrace(), WithDebug())
	s, err := c.QRCodeStart()
	assert.Nil(t, err)

	f, _ := os.CreateTemp("./", "tmp-qrcode-*.png")
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()
	b, err := s.QRCode()
	assert.Nil(t, err)

	_, err = f.Write(b)
	assert.Nil(t, err)

	status, err := c.QRCodeStatus(s)
	assert.Nil(t, err)

	if status.IsAllowed() {
		_, err = c.QRCodeLogin(s)
		assert.Nil(t, err)
	} else {
		_, err = c.QRCodeLogin(s)
		assert.Error(t, err)
	}
}

func TestQRCodeStartByOtherApp(t *testing.T) {
	t.Skip()
	c := New(WithTrace(), WithDebug())
	s, err := c.QRCodeStart()
	assert.Nil(t, err)

	f, _ := os.CreateTemp("./", "tmp-qrcode-*.png")
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()
	b, err := s.QRCodeByApi()
	assert.Nil(t, err)

	_, err = f.Write(b)
	assert.Nil(t, err)

	timer := time.NewTimer(50 * time.Second)
	defer timer.Stop()
	ch := make(chan error)
	go func() {
		for {
			status, err := c.QRCodeStatus(s)
			if err != nil {
				ch <- err
			}

			switch {
			case status.IsAllowed():
				_, err = c.QRCodeLoginWithApp(s, LoginAppIOS)
				ch <- err
				return
			case status.IsCanceled(), status.IsExpired():
				ch <- nil
				return
			case status.IsWaiting(), status.IsScanned():
				time.Sleep(1 * time.Second)
			default:
				_, err = c.QRCodeLoginWithApp(s, LoginAppIOS)
				ch <- err
				return
			}
		}
	}()

LOOP:
	for {
		select {
		case <-timer.C:
			assert.True(t, false, "time out")
			break LOOP
		case err := <-ch:
			assert.NoError(t, err)
			break LOOP
		}
	}
}

func TestShareSnap(t *testing.T) {
	down := teardown(t)
	defer down(t)

	_, err := client.GetShareSnap("sw6pw793wfp", "w816", "")
	assert.ErrorIs(t, err, ErrSharedNotFound)
}

func TestGetVersion(t *testing.T) {
	down := teardown(t)
	defer down(t)

	vers, err := client.GetAppVersion()
	assert.NoError(t, err)
	assert.NotEmpty(t, vers)
}

func TestGetInfo(t *testing.T) {
	down := teardown(t)
	defer down(t)

	info, err := client.GetInfo()
	assert.NoError(t, err)
	assert.NotEmpty(t, info.SpaceInfo)
}
