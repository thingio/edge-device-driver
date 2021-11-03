package driver

import "github.com/thingio/edge-device-sdk/models"

// DeviceConn 定义了接入平台设备应实现的方法集
type DeviceConn interface {
	// Start will try to create connections to those devices who need to activate.
	// It will always return nil if protocol need not connect.
	Start() error

	// Stop will try to stop the connections to devices.
	// It will always return nil if protocol need not connect.
	Stop() error

	// Watch will read device properties regularly with the policy
	// specified in Product.
	// The key and value of properties will be put into the device data channel.
	Watch(bus chan models.DeviceData) error                    // 从设备定时读取数据
	Write(propID models.ProductPropID, data interface{}) error // 向设备的单个字段写入数据
	Subscribe(eventID models.ProductEvtID, bus chan models.DeviceData) error
	Call(methodID models.ProductMethodID, ins models.DeviceData) (models.DeviceData, error)

	Ping() bool
}

type Builder func(product *models.Product, device *models.Device) (DeviceConn, error)
