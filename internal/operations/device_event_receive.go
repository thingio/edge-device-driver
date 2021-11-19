package operations

import (
	bus "github.com/thingio/edge-device-sdk/internal/message_bus"
	"github.com/thingio/edge-device-sdk/pkg/models"
)

func (c *deviceManagerDeviceDataOperationClient) Receive(productID, deviceID string,
	eventID models.ProductEventID) (dd <-chan models.DeviceData, cc func(), err error) {
	data := models.NewDeviceData(productID, deviceID, models.DeviceDataOperationEvent, eventID)
	message, err := data.ToMessage()
	if err != nil {
		return nil, nil, err
	}
	topic := message.Topic
	if _, ok := c.events[eventID]; !ok {
		c.events[eventID] = make(chan models.DeviceData, 100)
	}
	if err = c.mb.Subscribe(func(msg *bus.Message) {
		_, fields, _ := msg.Parse()
		data.SetFields(fields)

		c.events[eventID] <- *data
	}, topic); err != nil {
		return nil, nil, err
	}

	return c.events[eventID], func() {
		if _, ok := c.events[topic]; ok {
			close(c.events[topic])
			delete(c.events, topic)
		}
	}, nil
}
