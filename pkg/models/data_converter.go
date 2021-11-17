package models

import (
	"strconv"
)

var StrConverters = map[PropertyType]StrConverter{
	PropertyTypeInt:    Str2Int,
	PropertyTypeUint:   Str2Uint,
	PropertyTypeFloat:  Str2Float,
	PropertyTypeBool:   Str2Bool,
	PropertyTypeString: Str2String,
}

type StrConverter func(s string) (interface{}, error)

func Str2Int(s string) (interface{}, error) {
	return strconv.ParseInt(s, 10, 64)
}

func Str2Uint(s string) (interface{}, error) {
	return strconv.ParseUint(s, 10, 64)
}

func Str2Float(s string) (interface{}, error) {
	return strconv.ParseFloat(s, 64)
}

func Str2Bool(s string) (interface{}, error) {
	return strconv.ParseBool(s)
}

func Str2String(s string) (interface{}, error) {
	return s, nil
}
