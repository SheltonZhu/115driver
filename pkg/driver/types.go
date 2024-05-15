package driver

import (
	"encoding/json"
	"strconv"
	"time"
)

// StringInt uses for json field which maybe a string or an int.
type StringInt int64

func (v *StringInt) UnmarshalJSON(b []byte) (err error) {
	var i int
	if b[0] == '"' {
		var s string
		if err = json.Unmarshal(b, &s); err == nil {
			i, _ = strconv.Atoi(s)
		}
	} else {
		err = json.Unmarshal(b, &i)
	}
	if err == nil {
		*v = StringInt(i)
	}
	return
}

// StringInt64 uses for json field which maybe a string or an int64.
type StringInt64 int64

func (v *StringInt64) UnmarshalJSON(b []byte) (err error) {
	var i int64
	if b[0] == '"' {
		var s string
		if err = json.Unmarshal(b, &s); err == nil {
			i, err = strconv.ParseInt(s, 10, 64)
		}
	} else {
		err = json.Unmarshal(b, &i)
	}
	if err == nil {
		*v = StringInt64(i)
	}
	return
}

// StringFloat64 uses for json field which maybe a string or a float64.
type StringFloat64 float64

func (v *StringFloat64) UnmarshalJSON(b []byte) (err error) {
	var f float64
	if b[0] == '"' {
		var s string
		if err = json.Unmarshal(b, &s); err == nil {
			f, err = strconv.ParseFloat(s, 64)
		}
	} else {
		err = json.Unmarshal(b, &f)
	}
	if err == nil {
		*v = StringFloat64(f)
	}
	return
}

type IntString string

func (v *IntString) UnmarshalJSON(b []byte) (err error) {
	var s string
	if b[0] == '"' {
		err = json.Unmarshal(b, &s)
	} else {
		var i int64
		if err = json.Unmarshal(b, &i); err == nil {
			s = strconv.FormatInt(i, 10)
		}
	}
	if err == nil {
		*v = IntString(s)
	}
	return
}

type BoolInt int

func (v *BoolInt) UnmarshalJSON(b []byte) (err error) {
	if b[0] == 'f' || b[0] == 'F' {
		*v = -1
	} else {
		var i int
		if err = json.Unmarshal(b, &i); err == nil {
			*v = BoolInt(i)
		}
	}
	return
}

func BoolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

type StringTime int64

func (v *StringTime) UnmarshalJSON(b []byte) (err error) {
	var t time.Time
	if b[0] == '"' {
		var s string
		if err = json.Unmarshal(b, &s); err == nil {
			t, err = time.Parse("2006-01-02 15:04", s)
		}
	} else {
		err = json.Unmarshal(b, &t)
	}
	if err == nil {
		*v = StringTime(t.Unix())
	}
	return
}
