package operations

import (
	bus "github.com/thingio/edge-device-sdk/internal/message_bus"
	"github.com/thingio/edge-device-sdk/pkg/models"
)

type ListDevicesRequest struct {
	ProductID string `json:"product_id"`
}

func (r *ListDevicesRequest) Unmarshal(fields map[string]interface{}) error {
	return Map2Struct(fields, r)
}
func (r *ListDevicesRequest) Marshal() (map[string]interface{}, error) {
	return Struct2Map(*r)
}

type ListDevicesResponse struct {
	Devices []*models.Device `json:"devices"`
}

func (r *ListDevicesResponse) Unmarshal(fields map[string]interface{}) error {
	return Map2Struct(fields, r)
}
func (r *ListDevicesResponse) Marshal() (map[string]interface{}, error) {
	return Struct2Map(*r)
}

// OnListDevices for the device manager puts devices into the message bus.
func (c *deviceManagerDeviceOperationClient) OnListDevices(list func(productID string) ([]*models.Device, error)) error {
	schema := bus.NewMetaData(bus.MetaDataTypeDevice, bus.MetaDataOperationList,
		bus.MetaDataOperationModeRequest, bus.TopicWildcard)
	message, err := schema.ToMessage()
	if err != nil {
		return err
	}
	if err = c.mb.Subscribe(func(msg *bus.Message) {
		// parse request from the message
		_, fields, err := msg.Parse()
		if err != nil {
			c.logger.WithError(err).Errorf("fail to parse the message[%s] for listing devices", msg.ToString())
			return
		}
		req := &ListDevicesRequest{}
		if err := req.Unmarshal(fields); err != nil {
			c.logger.WithError(err).Error("fail to unmarshal the request for listing devices")
			return
		}

		productID := req.ProductID
		devices, err := list(productID)
		if err != nil {
			c.logger.WithError(err).Error("fail to list devices")
			return
		}

		// publish response
		response := bus.NewMetaData(bus.MetaDataTypeDevice, bus.MetaDataOperationList,
			bus.MetaDataOperationModeResponse, productID)
		rsp := &ListDevicesResponse{Devices: devices}
		fields, err = rsp.Marshal()
		if err != nil {
			c.logger.WithError(err).Error("fail to marshal the response for listing devices")
			return
		}
		response.SetFields(fields)
		if err := c.mb.Publish(response); err != nil {
			c.logger.WithError(err).Error("fail to publish the response for listing devices")
			return
		}
	}, message.Topic); err != nil {
		return err
	}
	return nil
}

// ListDevices for the device service takes devices from the message bus.
func (c *deviceServiceDeviceOperationClient) ListDevices(productID string) ([]*models.Device, error) {
	request := bus.NewMetaData(bus.MetaDataTypeDevice, bus.MetaDataOperationList,
		bus.MetaDataOperationModeRequest, productID)
	fields, err := (&ListDevicesRequest{ProductID: productID}).Marshal()
	if err != nil {
		return nil, err
	}
	request.SetFields(fields)
	response, err := c.mb.Call(request)
	if err != nil {
		return nil, err
	}

	rsp := &ListDevicesResponse{}
	if err := rsp.Unmarshal(response.GetFields()); err != nil {
		return nil, err
	}
	return rsp.Devices, nil
}
