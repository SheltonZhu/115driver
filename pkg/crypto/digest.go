package hash

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"io"
	"strings"
)

const (
	hashPreSize = 128 * 1024
)

type DigestResult struct {
	Size    int64
	PreId   string
	QuickId string
	MD5     string
}

func Digest(r io.Reader, result *DigestResult) (err error) {
	hs, hm := sha1.New(), md5.New()
	w := io.MultiWriter(hs, hm)
	// Calculate SHA1 hash of first 128K, which is used as PreId
	result.Size, err = io.CopyN(w, r, hashPreSize)
	if err != nil && err != io.EOF {
		return
	}
	result.PreId = strings.ToUpper(hex.EncodeToString(hs.Sum(nil)))
	// Write remain data.
	if err == nil {
		var n int64
		if n, err = io.Copy(w, r); err != nil {
			return
		}
		result.Size += n
		result.QuickId = strings.ToUpper(hex.EncodeToString(hs.Sum(nil)))
	} else {
		result.QuickId = result.PreId
	}
	result.MD5 = base64.StdEncoding.EncodeToString(hm.Sum(nil))
	return nil
}
