package operations

import (
	"errors"
	bus "github.com/thingio/edge-device-sdk/internal/message_bus"
)

type UnregisterProtocolRequest struct {
	ProtocolID string `json:"protocol_id"`
}

func (r *UnregisterProtocolRequest) Unmarshal(fields map[string]interface{}) error {
	return Map2Struct(fields, r)
}
func (r *UnregisterProtocolRequest) Marshal() (map[string]interface{}, error) {
	return Struct2Map(*r)
}

type UnregisterProtocolResponse struct {
	Success bool `json:"success"`
}

func (r *UnregisterProtocolResponse) Unmarshal(fields map[string]interface{}) error {
	return Map2Struct(fields, r)
}
func (r *UnregisterProtocolResponse) Marshal() (map[string]interface{}, error) {
	return Struct2Map(*r)
}

// UnregisterProtocols for the device manager takes the protocols from the message bus.
func (c *deviceManagerProtocolOperationClient) UnregisterProtocols(unregister func(protocolID string) error) error {
	schema := bus.NewMetaData(bus.MetaDataTypeProtocol, bus.MetaDataOperationDelete,
		bus.MetaDataOperationModeRequest, bus.TopicWildcard)
	message, err := schema.ToMessage()
	if err != nil {
		return err
	}
	if err = c.mb.Subscribe(func(msg *bus.Message) {
		// parse request from the message
		_, fields, err := msg.Parse()
		if err != nil {
			c.logger.WithError(err).Error("fail to parse the message for unregistering protocol")
			return
		}
		req := &UnregisterProtocolRequest{}
		if err := req.Unmarshal(fields); err != nil {
			c.logger.WithError(err).Error("fail to unmarshal the request for unregistering protocol")
			return
		}
		protocolID := req.ProtocolID
		if err := unregister(protocolID); err != nil {
			c.logger.Error(err.Error())
			return
		}

		// publish response
		response := bus.NewMetaData(bus.MetaDataTypeProtocol, bus.MetaDataOperationDelete,
			bus.MetaDataOperationModeResponse, protocolID)
		rsp := &UnregisterProtocolResponse{Success: true}
		fields, err = rsp.Marshal()
		if err != nil {
			c.logger.WithError(err).Error("fail to marshal the response for unregistering protocol")
			return
		}
		response.SetFields(fields)
		if err := c.mb.Publish(response); err != nil {
			c.logger.WithError(err).Error("fail to publish the response for unregistering protocol")
			return
		}
	}, message.Topic); err != nil {
		return err
	}
	return nil
}

// UnregisterProtocol for the device service puts the protocol into the message bus.
func (c *deviceServiceProtocolOperationClient) UnregisterProtocol(protocolID string) error {
	request := bus.NewMetaData(bus.MetaDataTypeProtocol, bus.MetaDataOperationDelete,
		bus.MetaDataOperationModeRequest, protocolID)
	fields, err := (&UnregisterProtocolRequest{ProtocolID: protocolID}).Marshal()
	if err != nil {
		return err
	}
	request.SetFields(fields)
	response, err := c.mb.Call(request)
	if err != nil {
		return err
	}

	rsp := &UnregisterProtocolResponse{}
	if err := rsp.Unmarshal(response.GetFields()); err != nil {
		return err
	}
	if !rsp.Success {
		return errors.New("fail to unregister protocol")
	}
	return nil
}
