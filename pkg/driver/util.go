package driver

import (
	"runtime"
	"strconv"
	"strings"
	"time"
)

type Time int64

func Now() Time {
	return Time(time.Now().Unix())
}

func NowMilli() Time {
	return Time(time.Now().UnixMilli())
}

func (t Time) String() string {
	return strconv.FormatInt(t.ToInt64(), 10)
}

func (t Time) ToInt64() int64 {
	return int64(t)
}

func Date() string {
	GMT, _ := time.LoadLocation("GMT")
	now := time.Now().In(GMT)
	return now.Format(time.RFC1123)
}

func isCalledByAlistV3() bool {
	pc, _, _, ok := runtime.Caller(3)
	if !ok {
		return false
	}

	funcName := runtime.FuncForPC(pc).Name()
	return strings.Contains(funcName, "alist")
}
