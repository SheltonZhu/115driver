package driver

import (
	"strconv"
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
