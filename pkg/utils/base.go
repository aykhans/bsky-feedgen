package utils

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/whyrusleeping/go-did"
)

// ParseString attempts to parse the input string `s` into a value of the specified type T.
// It supports parsing into the following types:
//   - int, int8, int16, int32, int64
//   - uint, uint8, uint16, uint32, uint64
//   - float64
//   - bool
//   - string
//   - time.Duration
//   - url.URL / *url.URL
//   - did.DID / *did.DID
//
// If T is not one of these supported types, it returns an error.
// If parsing the string `s` fails for a supported type, it returns the zero value of T
// and the parsing error.
func ParseString[T any](s string) (T, error) {
	var value T

	switch any(value).(type) {
	case int:
		i, err := strconv.Atoi(s)
		if err != nil {
			return value, err
		}
		value = any(i).(T)
	case int8:
		i, err := strconv.ParseInt(s, 10, 8)
		if err != nil {
			return value, err
		}
		value = any(int8(i)).(T)
	case int16:
		i, err := strconv.ParseInt(s, 10, 16)
		if err != nil {
			return value, err
		}
		value = any(int16(i)).(T)
	case int32:
		i, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return value, err
		}
		value = any(int32(i)).(T)
	case int64:
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return value, err
		}
		value = any(i).(T)
	case uint:
		u, err := strconv.ParseUint(s, 10, 0)
		if err != nil {
			return value, err
		}
		value = any(uint(u)).(T)
	case uint8:
		u, err := strconv.ParseUint(s, 10, 8)
		if err != nil {
			return value, err
		}
		value = any(uint8(u)).(T)
	case uint16:
		u, err := strconv.ParseUint(s, 10, 16)
		if err != nil {
			return value, err
		}
		value = any(uint16(u)).(T)
	case uint32:
		u, err := strconv.ParseUint(s, 10, 32)
		if err != nil {
			return value, err
		}
		value = any(uint32(u)).(T)
	case uint64:
		u, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return value, err
		}
		value = any(u).(T)
	case float64:
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return value, err
		}
		value = any(f).(T)
	case bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return value, err
		}
		value = any(b).(T)
	case string:
		value = any(s).(T)
	case []string:
		var items []string
		err := json.Unmarshal([]byte(s), &items)
		if err != nil {
			return value, err
		}
		value = any(items).(T)
	case time.Duration:
		d, err := time.ParseDuration(s)
		if err != nil {
			return value, err
		}
		value = any(d).(T)
	case url.URL:
		u, err := url.Parse(s)
		if err != nil {
			return value, err
		}
		value = any(*u).(T)
	case *url.URL:
		u, err := url.Parse(s)
		if err != nil {
			return value, err
		}
		value = any(u).(T)
	case did.DID:
		d, err := did.ParseDID(s)
		if err != nil {
			return value, err
		}
		value = any(d).(T)
	case *did.DID:
		d, err := did.ParseDID(s)
		if err != nil {
			return value, err
		}
		value = any(&d).(T)
	default:
		return value, fmt.Errorf("unsupported type: %T", value)
	}

	return value, nil
}

func ToPtr[T any](value T) *T {
	return &value
}
