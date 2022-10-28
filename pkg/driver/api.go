package driver

const (
	// login
	ApiLoginCheck = "https://passportapi.115.com/app/1.0/web/1.0/check/sso"

	// dir
	ApiDirAdd = "https://webapi.115.com/files/add"

	// file
	ApiFileDelete = "https://webapi.115.com/rb/delete"
	ApiFileMove   = "https://webapi.115.com/files/move"
	ApiFileCopy   = "https://webapi.115.com/files/copy"
	ApiFileRename = "https://webapi.115.com/files/batch_rename"

	ApiFileList       = "https://webapi.115.com/files"
	ApiFileListByName = "https://aps.115.com/natsort/files.php"

	// download
	ApiDownloadGetUrl = "https://proapi.115.com/app/chrome/downurl"
)
