package helper

import (
	"github.com/thingio/edge-device-sdk/models"
	"strconv"
)

var strConverters = map[models.DeviceDataFieldType]strConverter{
	models.DeviceDataFieldTypeInt:    str2Int,
	models.DeviceDataFieldTypeUint:   str2Uint,
	models.DeviceDataFieldTypeFloat:  str2Float,
	models.DeviceDataFieldTypeBool:   str2Bool,
	models.DeviceDataFieldTypeString: str2String,
}

type strConverter func(s string) (interface{}, error)

func str2Int(s string) (interface{}, error) {
	return strconv.ParseInt(s, 10, 64)
}

func str2Uint(s string) (interface{}, error) {
	return strconv.ParseUint(s, 10, 64)

}

func str2Float(s string) (interface{}, error) {
	return strconv.ParseFloat(s, 64)
}

func str2Bool(s string) (interface{}, error) {
	return strconv.ParseBool(s)
}

func str2String(s string) (interface{}, error) {
	return s, nil
}
