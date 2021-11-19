package operations

import (
	"github.com/thingio/edge-device-sdk/pkg/models"
)

func (c *deviceManagerDeviceDataOperationClient) Call(productID, deviceID string,
	methodID models.ProductMethodID, ins map[string]interface{}) (outs map[string]interface{}, err error) {
	request := models.NewDeviceData(productID, deviceID, models.DeviceDataOperationRequest, methodID)
	request.SetFields(ins)

	response, err := c.mb.Call(request)
	if err != nil {
		return nil, err
	}
	return response.GetFields(), nil
}
