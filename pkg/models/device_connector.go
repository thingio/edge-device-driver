package models

import (
	"github.com/thingio/edge-device-sdk/logger"
)

type DeviceConnector interface {
	// Initialize will try to initialize a device connector to
	// create the connection with device which needs to activate.
	// It must always return nil if the device needn't be initialized.
	Initialize(lg *logger.Logger) error

	// Start will to try to create connection with the real device.
	// It must always return nil if the device needn't be initialized.
	Start() error
	// Stop will to try to destroy connection with the real device.
	// It must always return nil if the device needn't be initialized.
	Stop(force bool) error
	// Ping is used to test the connectivity of the real device.
	// If the device is connected, it will return true, else return false.
	Ping() bool

	// Watch will read device's properties periodically with the specified policy.
	Watch(bus chan<- DeviceData) error
	// Write will write the specified property to the real device.
	Write(propertyID ProductPropertyID, data interface{}) error
	// Subscribe will subscribe the specified event,
	// and put data belonging to this event into the bus.
	Subscribe(eventID ProductEventID, bus chan<- DeviceData) error
	// Call is used to call the specified method defined in product,
	// then waiting for a while to receive its response.
	// If the call is timeout, it will return a timeout error.
	Call(methodID ProductMethodID, request DeviceData) (response DeviceData, err error) // method call
}

// ConnectorBuilder is used to create a new device connector using the specified product and device.
type ConnectorBuilder func(product *Product, device *Device) (DeviceConnector, error)
