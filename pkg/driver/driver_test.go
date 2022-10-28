package driver

import (
	"math/rand"
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

func TestDownload(t *testing.T) {
	down := teardown(t)
	defer down(t)

	pickCode := NowMilli().String()
	_, err := client.Download(pickCode)
	assert.ErrorIs(t, ErrPickCodeNotExist, err)
	_, err = client.Download("")
	assert.ErrorIs(t, ErrPickCodeisEmpty, err)
	f, err := client.Download("arod1twvavfexh9cv")
	assert.NotEmpty(t, f)
	assert.Nil(t, err)
}
