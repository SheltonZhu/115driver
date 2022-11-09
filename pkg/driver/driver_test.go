package driver

import (
	"io"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestImportFromCookie(t *testing.T) {
	cr := &Credential{}
	assert.Nil(t, cr.FromCookie("UID=1;CID=2;SEID=3;other=4"))
	assert.Error(t, ErrBadCookie, cr.FromCookie(""))
	assert.Error(t, ErrBadCookie, cr.FromCookie("k=a;;"))
	assert.Error(t, ErrBadCookie, cr.FromCookie("1=2;2=3;3=4"))
	assert.Error(t, ErrBadCookie, cr.FromCookie("1=2;2=3;3=4"))
}

func TestLogin(t *testing.T) {
	assert.Error(t, New().ImportCredential(&Credential{}).LoginCheck())
}

var cookieStr = ""
var client *Pan115Client

func teardown(t *testing.T) func(t *testing.T) {
	cr := &Credential{}
	assert.Nil(t, cr.FromCookie(cookieStr))
	client = New(UA(UA115Disk), WithDebug(), WithTrace()).ImportCredential(cr)
	assert.Nil(t, client.LoginCheck())
	rand.Seed(time.Now().Unix())
	return func(t *testing.T) {}
}

func TestMkdir(t *testing.T) {
	down := teardown(t)
	defer down(t)

	dirName := NowMilli().String()
	_, err := client.Mkdir("0", dirName)
	assert.Nil(t, err)
	_, err = client.Mkdir("0", dirName)
	assert.ErrorIs(t, ErrExist, err)
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

	f, err := client.List("0")
	assert.NotEmpty(t, *f)
	assert.Nil(t, err)
	dirName := NowMilli().String()
	f, err = client.List(dirName)
	assert.Nil(t, err)
	assert.Empty(t, *f)
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
	assert.ErrorIs(t, ErrPickCodeNotExist, err)
	_, err = client.Download("")
	assert.ErrorIs(t, ErrPickCodeisEmpty, err)
}

func TestGetUploadInfo(t *testing.T) {
	down := teardown(t)
	defer down(t)

	assert.Nil(t, client.GetUploadInfo())
}
func TestUploadSHA1(t *testing.T) {
	down := teardown(t)
	defer down(t)

	r := strings.NewReader(NowMilli().String())
	d, err := client.GetDigestResult(r)
	assert.Nil(t, err)
	_, err = client.UploadSHA1(d.Size, "xxx.txt", "0", d.PreID, d.QuickID)
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
	resp, err := client.UploadSHA1(d.Size, randStr+".txt", "0", d.PreID, d.QuickID)
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

func TestUploadMultpart(t *testing.T) {
	start := time.Now()

	down := teardown(t)
	defer down(t)

	f, err := os.CreateTemp("./", "test-temp-*")
	assert.Nil(t, err)

	randStr := NowMilli().String()
	_, err = f.WriteString(randStr)
	assert.Nil(t, err)

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

	_, err := client.GetFile("2491635336026092299")
	assert.Error(t, err)
}

func TestQRCodeStart(t *testing.T) {
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
