package operations

import (
	"errors"
	"github.com/thingio/edge-device-sdk/pkg/models"
)

// ListDevices for the device manager puts devices into the message bus.
func (c *deviceManagerDeviceOperationClient) ListDevices(productID string,
	listDevices func(productID string) ([]*models.Device, error)) error {
	return errors.New("implement me")
}

// ListDevices for the device service takes devices from the message bus.
func (c *deviceServiceDeviceOperationClient) ListDevices(productID string) ([]*models.Device, error) {
	return nil, errors.New("implement me")
}
