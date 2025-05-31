package driver

const (
	ApiGetVersion = "https://appversion.115.com/1/web/1.0/api/chrome"

	// login
	ApiLoginCheck  = "https://passportapi.115.com/app/1.0/web/1.0/check/sso"
	ApiUserInfo    = "https://my.115.com/?ct=ajax&ac=nav"
	ApiStatusCheck = "https://my.115.com/?ct=guide&ac=status"
	// dir
	ApiDirAdd = "https://webapi.115.com/files/add"
	ApiDirName2CID = "https://webapi.115.com/files/getid"

	// file
	ApiFileDelete    = "https://webapi.115.com/rb/delete"
	ApiFileMove      = "https://webapi.115.com/files/move"
	ApiFileCopy      = "https://webapi.115.com/files/copy"
	ApiFileRename    = "https://webapi.115.com/files/batch_rename"
	ApiFileIndexInfo = "https://webapi.115.com/files/index_info"

	ApiFileList       = "https://webapi.115.com/files"
	ApiFileList1       = "http://web.api.115.com/files"
	// ApiFileList2       = "http://anxia.com/webapi/files"
	// ApiFileList3       = "http://v.anxia.com/webapi/files"
	ApiFileListByName = "https://aps.115.com/natsort/files.php"

	ApiFileStat = "https://webapi.115.com/category/get"
	ApiFileInfo = "https://webapi.115.com/files/get_info"

	// share
	ApiShareSnap = "https://webapi.115.com/share/snap"

	// download
	ApiDownloadGetUrl        = "https://proapi.115.com/app/chrome/downurl"
	ApiDownloadGetShareUrl   = "https://proapi.115.com/app/share/downurl"
	AndroidApiDownloadGetUrl = "https://proapi.115.com/android/2.0/ufile/download"

	// offline download
	ApiAddOfflineUrl   = "https://lixian.115.com/lixianssp/?ac=add_task_urls"
	ApiDelOfflineUrl   = "https://lixian.115.com/lixian/?ct=lixian&ac=task_del"
	ApiListOfflineUrl  = "https://lixian.115.com/lixian/?ct=lixian&ac=task_lists"
	ApiClearOfflineUrl = "https://lixian.115.com/lixian/?ct=lixian&ac=task_clear"

	// upload
	ApiUploadInfo        = "https://proapi.115.com/app/uploadinfo"
	ApiGetUploadEndpoint = "https://uplb.115.com/3.0/getuploadinfo.php"
	ApiUploadInit        = "https://uplb.115.com/4.0/initupload.php"

	// oss
	ApiUploadOSSToken = "https://uplb.115.com/3.0/gettoken.php"

	// qrcode
	ApiQrcodeToken        = "https://qrcodeapi.115.com/api/1.0/web/1.0/token"
	ApiQrcodeStatus       = "https://qrcodeapi.115.com/get/status/"
	ApiQrcodeLogin        = "https://passportapi.115.com/app/1.0/web/1.0/login/qrcode"
	ApiQrcodeLoginWithApp = "https://passportapi.115.com/app/1.0/%s/1.0/login/qrcode"
	ApiQrcodeImage        = "https://qrcodeapi.115.com/api/1.0/mac/1.0/qrcode?uid=%s"

	// recycle
	ApiRecycleList   = "https://webapi.115.com/rb"
	ApiRecycleClean  = "https://webapi.115.com/rb/clean"
	ApiRecycleRevert = "https://webapi.115.com/rb/revert"
)
