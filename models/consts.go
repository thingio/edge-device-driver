package models

import (
	"time"
)

type (
	DeviceOperation   = string        // 设备的操作类型
	DeviceReportModel = string        // 设备的数据上报类型
	ProductFuncID     = string        // 产品功能的ID
	ProductFuncType   = string        // 产品功能的类型
	ProductPropID     = ProductFuncID // 设备属性的功能ID
	ProductEvtID      = ProductFuncID // 设备时间的功能ID
	ProductMethodID   = ProductFuncID // 设备服务的功能ID

	ProtocolDevPropKey  = string // 协议定义中设备共有的属性字段
	DeviceDataFieldType = string
)

const (
	RegularMode                          = "regular" // using interval : 5s, 1m, 0.5h
	OnChangedModel                       = "changed" // upload device data while it change
	DefaultMode                          = RegularMode
	DeviceDataRegularModeDefaultInterval = "5s"

	DeviceRead     DeviceOperation = "read"     // 读取设备属性
	DeviceWrite    DeviceOperation = "write"    // 写入设备属性
	DeviceEvent    DeviceOperation = "event"    // 接收到设备事件
	DeviceRequest  DeviceOperation = "request"  // 请求设备服务
	DeviceResponse DeviceOperation = "response" // 设备服务返回的结果
	DeviceError    DeviceOperation = "error"    // 请求或写入时，若发生错误，则以此作为opt_type

	PropFunc   ProductFuncType = "props"   // 产品-属性功能-读写设备属性
	EventFunc  ProductFuncType = "events"  // 产品-事件功能-接收设备事件
	MethodFunc ProductFuncType = "methods" // 产品-服务功能-设备服务调用

	DeviceDataMultiPropsID   = "*"
	DeviceDataMultiPropsName = "多属性"

	DeviceDataFieldTypeInt    DeviceDataFieldType = "int"
	DeviceDataFieldTypeUint   DeviceDataFieldType = "uint"
	DeviceDataFieldTypeFloat  DeviceDataFieldType = "float"
	DeviceDataFieldTypeBool   DeviceDataFieldType = "bool"
	DeviceDataFieldTypeString DeviceDataFieldType = "string"

	DefaultConnectTimeout = 3 * time.Second
)

var (
	AvailableFieldTypes = map[string]struct{}{
		DeviceDataFieldTypeInt:    {},
		DeviceDataFieldTypeUint:   {},
		DeviceDataFieldTypeFloat:  {},
		DeviceDataFieldTypeBool:   {},
		DeviceDataFieldTypeString: {},
	}
	AvailableReportModes = map[string]struct{}{
		RegularMode:    {},
		OnChangedModel: {},
	}
)
