package operations

import (
	bus "github.com/thingio/edge-device-sdk/internal/message_bus"
	"github.com/thingio/edge-device-sdk/pkg/models"
)

type ListProductsRequest struct {
	ProtocolID string `json:"protocol_id"`
}

func (r *ListProductsRequest) Unmarshal(fields map[string]interface{}) error {
	return Map2Struct(fields, r)
}
func (r *ListProductsRequest) Marshal() (map[string]interface{}, error) {
	return Struct2Map(*r)
}

type ListProductsResponse struct {
	Products []*models.Product `json:"products"`
}

func (r *ListProductsResponse) Unmarshal(fields map[string]interface{}) error {
	return Map2Struct(fields, r)
}
func (r *ListProductsResponse) Marshal() (map[string]interface{}, error) {
	return Struct2Map(*r)
}

// ListProducts for the device manager puts products into the message bus.
func (c *deviceManagerProductOperationClient) ListProducts(list func(protocolID string) ([]*models.Product, error)) error {
	schema := bus.NewMetaData(bus.MetaDataTypeProduct, bus.MetaDataOperationList,
		bus.MetaDataOperationModeRequest, bus.TopicWildcard)
	message, err := schema.ToMessage()
	if err != nil {
		return err
	}
	if err = c.mb.Subscribe(func(msg *bus.Message) {
		// parse request from the message
		_, fields, err := msg.Parse()
		if err != nil {
			c.logger.WithError(err).Error("fail to parse the message for listing products")
			return
		}
		req := &ListProductsRequest{}
		if err := req.Unmarshal(fields); err != nil {
			c.logger.WithError(err).Error("fail to unmarshal the request for listing products")
			return
		}

		protocolID := req.ProtocolID
		products, err := list(protocolID)
		if err != nil {
			c.logger.WithError(err).Error("fail to list products")
			return
		}

		// publish response
		response := bus.NewMetaData(bus.MetaDataTypeProduct, bus.MetaDataOperationList,
			bus.MetaDataOperationModeResponse, protocolID)
		if err != nil {
			c.logger.WithError(err).Error("fail to construct response for the request")
			return
		}
		rsp := &ListProductsResponse{Products: products}
		fields, err = rsp.Marshal()
		if err != nil {
			c.logger.WithError(err).Error("fail to marshal the response for listing products")
			return
		}
		response.SetFields(fields)
		if err := c.mb.Publish(response); err != nil {
			c.logger.WithError(err).Error("fail to publish the response for listing products")
			return
		}
	}, message.Topic); err != nil {
		return err
	}
	return nil
}

// ListProducts for the device service takes products from the message bus.
func (c *deviceServiceProductOperationClient) ListProducts(protocolID string) ([]*models.Product, error) {
	request := bus.NewMetaData(bus.MetaDataTypeProduct, bus.MetaDataOperationList,
		bus.MetaDataOperationModeRequest, protocolID)
	fields, err := (&ListProductsRequest{ProtocolID: protocolID}).Marshal()
	if err != nil {
		return nil, err
	}
	request.SetFields(fields)
	response, err := c.mb.Call(request)
	if err != nil {
		return nil, err
	}

	rsp := &ListProductsResponse{}
	if err := rsp.Unmarshal(response.GetFields()); err != nil {
		return nil, err
	}
	return rsp.Products, nil
}
