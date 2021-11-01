package models

type DeviceOperation string

const (
	DeviceRead           DeviceOperation = "read"     // Device Property Read
	DeviceWrite          DeviceOperation = "write"    // Device Property Write
	DeviceEvent          DeviceOperation = "event"    // Device Event Receive
	DeviceRequest        DeviceOperation = "request"  // Device Service Request
	DeviceResponse       DeviceOperation = "response" // Device Service Response
	DeviceOperationError DeviceOperation = "error"    // Device Operation Error
)
