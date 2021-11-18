package operations

import (
	"errors"
	bus "github.com/thingio/edge-device-sdk/internal/message_bus"
	"github.com/thingio/edge-device-sdk/pkg/models"
)

type RegisterProtocolRequest struct {
	Protocol models.Protocol `json:"protocol"`
}

func (r *RegisterProtocolRequest) Unmarshal(fields map[string]interface{}) error {
	return Map2Struct(fields, r)
}
func (r *RegisterProtocolRequest) Marshal() (map[string]interface{}, error) {
	return Struct2Map(*r)
}

type RegisterProtocolResponse struct {
	Success bool `json:"success"`
}

func (r *RegisterProtocolResponse) Unmarshal(fields map[string]interface{}) error {
	return Map2Struct(fields, r)
}
func (r *RegisterProtocolResponse) Marshal() (map[string]interface{}, error) {
	return Struct2Map(*r)
}

// RegisterProtocols for the device manager takes the protocols from the message bus.
func (c *deviceManagerProtocolOperationClient) RegisterProtocols(register func(protocol *models.Protocol) error) error {
	schema := bus.NewMetaData(bus.MetaDataTypeProtocol, bus.MetaDataOperationCreate,
		bus.MetaDataOperationModeRequest, bus.TopicWildcard)
	message, err := schema.ToMessage()
	if err != nil {
		return err
	}
	if err = c.mb.Subscribe(func(msg *bus.Message) {
		// parse request from the message
		_, fields, err := msg.Parse()
		if err != nil {
			c.logger.WithError(err).Error("fail to parse the message for registering protocol")
			return
		}
		req := &RegisterProtocolRequest{}
		if err := req.Unmarshal(fields); err != nil {
			c.logger.WithError(err).Error("fail to unmarshal the request for registering protocol")
			return
		}
		protocol := &req.Protocol
		if err := register(protocol); err != nil {
			c.logger.Error(err.Error())
			return
		}

		// publish response
		response := bus.NewMetaData(bus.MetaDataTypeProtocol, bus.MetaDataOperationCreate,
			bus.MetaDataOperationModeResponse, protocol.ID)
		rsp := &RegisterProtocolResponse{Success: true}
		fields, err = rsp.Marshal()
		if err != nil {
			c.logger.WithError(err).Error("fail to marshal the response for registering protocol")
			return
		}
		response.SetFields(fields)
		if err := c.mb.Publish(response); err != nil {
			c.logger.WithError(err).Error("fail to publish the response for registering protocol")
			return
		}
	}, message.Topic); err != nil {
		return err
	}
	return nil
}

// RegisterProtocol for the device service puts the protocol into the message bus.
func (c *deviceServiceProtocolOperationClient) RegisterProtocol(protocol *models.Protocol) error {
	request := bus.NewMetaData(bus.MetaDataTypeProtocol, bus.MetaDataOperationCreate,
		bus.MetaDataOperationModeRequest, protocol.ID)
	fields, err := (&RegisterProtocolRequest{Protocol: *protocol}).Marshal()
	if err != nil {
		return err
	}
	request.SetFields(fields)
	response, err := c.mb.Call(request)
	if err != nil {
		return err
	}

	rsp := &RegisterProtocolResponse{}
	if err := rsp.Unmarshal(response.GetFields()); err != nil {
		return err
	}
	if !rsp.Success {
		return errors.New("fail to register protocol")
	}
	return nil
}
