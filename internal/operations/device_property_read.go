package operations

import (
	bus "github.com/thingio/edge-device-sdk/internal/message_bus"
	"github.com/thingio/edge-device-sdk/pkg/models"
)

func (c *deviceManagerDeviceDataOperationClient) Read(productID, deviceID string,
	propertyID models.ProductPropertyID) (dd <-chan models.DeviceData, cc func(), err error) {
	data := models.NewDeviceData(productID, deviceID, models.DeviceDataOperationRead, propertyID)
	message, err := data.ToMessage()
	if err != nil {
		return nil, nil, err
	}
	topic := message.Topic
	if _, ok := c.reads[topic]; !ok {
		c.reads[topic] = make(chan models.DeviceData, 100)
	}
	if err = c.mb.Subscribe(func(msg *bus.Message) {
		_, fields, _ := msg.Parse()
		data.SetFields(fields)

		c.reads[propertyID] <- *data
	}, topic); err != nil {
		return nil, nil, err
	}

	return c.reads[propertyID], func() {
		if _, ok := c.reads[topic]; ok {
			close(c.reads[topic])
			delete(c.reads, topic)
		}
	}, nil
}
