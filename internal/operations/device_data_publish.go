package operations

import "github.com/thingio/edge-device-sdk/pkg/models"

func (c *deviceServiceDeviceDataOperationClient) Publish(data models.DeviceData) error {
	return c.mb.Publish(&data)
}
