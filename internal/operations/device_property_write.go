package operations

import (
	"fmt"
	"github.com/thingio/edge-device-sdk/pkg/models"
)

func (c *deviceManagerDeviceDataOperationClient) Write(productID, deviceID string,
	propertyID models.ProductPropertyID, value interface{}) error {
	data := models.NewDeviceData(productID, deviceID, models.DeviceDataOperationWrite, propertyID)
	if propertyID == models.DeviceDataMultiPropsID {
		fields, ok := value.(map[models.ProductPropertyID]interface{})
		if !ok {
			return fmt.Errorf("%+v must be type map[models.ProductPropertyID]interface{}", value)
		}
		data.SetFields(fields)
	} else {
		data.SetField(propertyID, value)
	}
	return c.mb.Publish(data)
}
