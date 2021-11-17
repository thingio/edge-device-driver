package models

import (
	"github.com/thingio/edge-device-sdk/logger"
)

type DeviceConnector interface {
	Initialize(lg *logger.Logger) error

	Start() error
	Stop(force bool) error
	Ping() bool

	Watch(bus chan<- DeviceData) error                                                  // property read periodically
	Write(propertyID ProductPropertyID, data interface{}) error                         // property write
	Subscribe(eventID ProductEventID, bus <-chan DeviceData) error                      // event subscribe
	Call(methodID ProductMethodID, request DeviceData) (response DeviceData, err error) // method call
}

type ConnectorBuilder func(product *Product, device *Device) (DeviceConnector, error)
