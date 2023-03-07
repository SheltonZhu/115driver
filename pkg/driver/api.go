package driver

const (
	// login
	ApiLoginCheck = "https://passportapi.115.com/app/1.0/web/1.0/check/sso"
	ApiUserInfo   = "https://my.115.com/?ct=ajax&ac=nav"

	// dir
	ApiDirAdd = "https://webapi.115.com/files/add"

	// file
	ApiFileDelete = "https://webapi.115.com/rb/delete"
	ApiFileMove   = "https://webapi.115.com/files/move"
	ApiFileCopy   = "https://webapi.115.com/files/copy"
	ApiFileRename = "https://webapi.115.com/files/batch_rename"

	ApiFileList       = "https://webapi.115.com/files"
	ApiFileListByName = "https://aps.115.com/natsort/files.php"

	ApiFileStat = "https://webapi.115.com/category/get"
	ApiFileInfo = "https://webapi.115.com/files/get_info"

	// download
	ApiDownloadGetUrl = "https://proapi.115.com/app/chrome/downurl"

	// upload
	ApiUploadInfo = "https://proapi.115.com/app/uploadinfo"
	ApiUploadInit = "https://uplb.115.com/4.0/initupload.php"

	// oss
	ApiUploadOSSToken = "https://uplb.115.com/3.0/gettoken.php"

	// qrcode
	ApiQrcodeToken  = "https://qrcodeapi.115.com/api/1.0/web/1.0/token"
	ApiQrcodeStatus = "https://qrcodeapi.115.com/get/status/"
	ApiQrcodeLogin  = "https://passportapi.115.com/app/1.0/web/1.0/login/qrcode"
)
