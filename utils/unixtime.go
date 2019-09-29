package utils

import (
	"strconv"
	"time"
)

// UnixTime wraps a time.Time to provide JSON unmarshalling from unix timestamps
type UnixTime struct {
	time.Time
}

// UnmarshalJSON parses a unix timestamp into a UnixTime
func (unixtime *UnixTime) UnmarshalJSON(data []byte) (err error) {
	if string(data) == "null" {
		return nil
	}
	i, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return err
	}
	tm := time.Unix(i, 0)
	unixtime.Time = tm
	return
}

// GetUnixTime parses the given string into a UnixTime.
// Before the actual conversion, the input value is divided by the given factor. This allows to convert timestamps
// which are based on milliseconds instead of seconds.
func GetUnixTime(val string, factor int64) (UnixTime, error) {
	i, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return UnixTime{}, err
	}
	tm := time.Unix(i/factor, 0)
	return UnixTime{Time: tm}, nil
}
