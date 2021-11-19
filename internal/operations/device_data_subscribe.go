package operations

import (
	bus "github.com/thingio/edge-device-sdk/internal/message_bus"
)

func (c *deviceServiceDeviceDataOperationClient) Subscribe(handler bus.MessageHandler, topic string) error {
	return c.mb.Subscribe(handler, topic)
}
