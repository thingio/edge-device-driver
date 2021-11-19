package models

import (
	"encoding/json"
	"fmt"
	"github.com/thingio/edge-device-sdk/internal/message_bus"
)

type (
	DeviceDataOperation          = string // the type of device data's operation
	DeviceDataPropertyReportMode = string // the mode of device data's reporting

	ProductFuncID     = string        // product functionality ID
	ProductFuncType   = string        // product functionality type
	ProductPropertyID = ProductFuncID // product property's functionality ID
	ProductEventID    = ProductFuncID // product event's functionality ID
	ProductMethodID   = ProductFuncID // product method's functionality ID
)

const (
	DeviceDataOperationRead     DeviceDataOperation = "read"     // Device Property Read
	DeviceDataOperationWrite    DeviceDataOperation = "write"    // Device Property Write
	DeviceDataOperationEvent    DeviceDataOperation = "event"    // Device Event
	DeviceDataOperationRequest  DeviceDataOperation = "request"  // Device Method Request
	DeviceDataOperationResponse DeviceDataOperation = "response" // Device Method Response
	DeviceDataOperationError    DeviceDataOperation = "error"    // Device Method Error

	DeviceDataReportModePeriodical DeviceDataPropertyReportMode = "periodical" // report device data at intervals, e.g. 5s, 1m, 0.5h
	DeviceDataReportModeMutated    DeviceDataPropertyReportMode = "mutated"    // report device data while mutated

	PropertyFunc ProductFuncType = "props"   // product property's functionality
	EventFunc    ProductFuncType = "events"  // product event's functionality
	MethodFunc   ProductFuncType = "methods" // product method's functionality

	DeviceDataMultiPropsID   ProductFuncID = "*"
	DeviceDataMultiPropsName               = "多属性"
)

// Opts2FuncType maps operation upon device data as product's functionality.
var opts2FuncType = map[DeviceDataOperation]ProductFuncType{
	DeviceDataOperationEvent:    EventFunc,
	DeviceDataOperationRead:     PropertyFunc,
	DeviceDataOperationWrite:    PropertyFunc,
	DeviceDataOperationRequest:  MethodFunc,
	DeviceDataOperationResponse: MethodFunc,
	DeviceDataOperationError:    MethodFunc,
}

type DeviceData struct {
	bus.MessageData

	ProductID string              `json:"product_id"`
	DeviceID  string              `json:"device_id"`
	OptType   DeviceDataOperation `json:"opt_type"`
	FuncID    ProductFuncID       `json:"func_id"`
	FuncType  ProductFuncType     `json:"func_type"`
}

func (d *DeviceData) ToMessage() (*bus.Message, error) {
	topic := bus.NewDeviceDataTopic(d.ProductID, d.DeviceID, d.OptType, d.FuncID)
	payload, err := json.Marshal(d.Fields)
	if err != nil {
		return nil, err
	}

	return &bus.Message{
		Topic:   topic.String(),
		Payload: payload,
	}, nil
}

func (d *DeviceData) isRequest() bool {
	return d.OptType == DeviceDataOperationRequest
}

func (d *DeviceData) Response() (response bus.Data, err error) {
	if !d.isRequest() {
		return nil, fmt.Errorf("the device data is not a request: %+v", *d)
	}
	return NewDeviceData(d.ProductID, d.DeviceID, DeviceDataOperationResponse, d.FuncID), nil
}

func NewDeviceData(productID, deviceID string, optType DeviceDataOperation, dataID ProductFuncID) *DeviceData {
	return &DeviceData{
		ProductID: productID,
		DeviceID:  deviceID,
		OptType:   optType,
		FuncID:    dataID,
		FuncType:  opts2FuncType[optType],
	}
}

func ParseDeviceData(msg *bus.Message) (*DeviceData, error) {
	tags, fields, err := msg.Parse()
	if err != nil {
		return nil, err
	}
	deviceData := NewDeviceData(tags[0], tags[1], tags[2], tags[3])
	deviceData.SetFields(fields)
	return deviceData, nil
}
