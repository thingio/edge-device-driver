package driver

import (
	"fmt"
	"github.com/thingio/edge-device-std/models"
	"github.com/thingio/edge-device-std/msgbus/message"
)

// publishingDeviceData tries to publish data in the bus into the MQTTMessageBus.
func (d *DeviceDriver) publishingDeviceData() {
	d.wg.Add(1)

	go func() {
		defer d.wg.Done()

		for {
			select {
			case data := <-d.bus:
				if err := d.doc.Publish(data); err != nil {
					d.logger.WithError(err).Errorf("fail to publish the device data %+v", data)
				} else {
					d.logger.Debugf("success to publish the device data %+v", data)
				}
			case <-d.ctx.Done():
				break
			}
		}
	}()
}

type DeviceOperationHandler func(product *models.Product, device *models.Device, conn models.DeviceTwin,
	dataID string, fields map[string]interface{}) error

func (d *DeviceDriver) handlingDeviceData() {
	d.wg.Add(1)

	go func() {
		defer d.wg.Done()

		topic := models.TopicTypeDeviceData.Topic()
		if err := d.doc.Subscribe(d.handleDeviceOperation, topic); err != nil {
			d.logger.WithError(err).Errorf("fail to subscribe the topic: %d", topic)
			return
		}
	}()
}

func (d *DeviceDriver) handleDeviceOperation(msg *message.Message) {
	o, err := models.ParseDataOperation(msg)
	if err != nil {
		d.logger.WithError(err).Errorf("fail to parse the message[%d]", msg.ToString())
		return
	}

	productID, deviceID, optType, dataID := o.ProductID, o.DeviceID, o.OptType, o.FuncID
	product, err := d.getProduct(productID)
	if err != nil {
		d.logger.Error(err.Error())
		return
	}
	device, err := d.getDevice(deviceID)
	if err != nil {
		d.logger.Error(err.Error())
		return
	}
	connector, err := d.getDeviceConnector(deviceID)
	if err != nil {
		d.logger.Error(err.Error())
		return
	}

	if _, ok := d.unsupportedDataOperations[optType]; ok {
		return
	}
	handler, ok := d.dataOperationHandlers[optType]
	if !ok {
		d.logger.Errorf("unsupported operation type: %d", optType)
		return
	}
	if err = handler(product, device, connector, dataID, o.GetFields()); err != nil {
		d.logger.WithError(err).Errorf("fail to handle the message: %+v", msg)
		return
	}
}

// handleHealthCheck is responsible for handling health check forwarded by the device manager.
//
// This handler could be tested as follows:
// 1. Send the specified format data to the message bus:
//    (a.) Directly (mock):   mosquitto_pub -h 172.16.251.163 -p 1883 -t "v1/DATA/randnum_test01/randnum_test01/health-check-ping/Intn" -m "{\"n\": 100}"
//    (b.) Indirectly:        invoke the DeviceManagerDeviceDataOperationClient.Call("randnum_test01", "randnum_test01", "Intn", map[string]interface{}{"n": 100})
// 2. Observe the log of device driver and subscribe the specified topic:
//	  mosquitto_sub -h 172.16.251.163 -p 1883 -t "v1/DATA/randnum_test01/randnum_test01/response/Intn".
func (d *DeviceDriver) handleHealthCheck(product *models.Product, device *models.Device, twin models.DeviceTwin,
	_ models.ProductMethodID, _ map[string]interface{}) error {
	fields := make(map[string]interface{})
	status, err := twin.HealthCheck()
	if err != nil {
		fields[models.ProductPropertyStatusDetail] = err.Error()
	}
	fields[models.ProductPropertyStatus] = status
	response := models.NewDataOperation(product.ID, device.ID, models.DataOperationTypeHealthCheckPong, device.ID)
	response.SetFields(fields)
	return d.doc.Publish(*response)
}

// handleWrite is responsible for handling the read request forwarded by the device manager.
// It will read fields from the real device.
//
// This handler could be tested as follows:
// 1. Send the specified format data to the message bus:
//    (a.) Directly (mock):   mosquitto_pub -h 172.16.251.163 -p 1883 -t "v1/DATA/randnum_test01/randnum_test01/read-req/float" -m "{\"float\": 100}"
//    (b.) Indirectly:        invoke the DeviceManagerDeviceDataOperationClient.Read("randnum_test01", "randnum_test01", "float", 100)
// 2. Observe the log of device driver and subscribe the specified topic:
//	  mosquitto_sub -h 172.16.251.163 -p 1883 -t "v1/DATA/randnum_test01/randnum_test01/read-rsp/float".
func (d *DeviceDriver) handleRead(product *models.Product, device *models.Device, twin models.DeviceTwin,
	propertyID string, _ map[string]interface{}) error {
	o := models.NewDataOperation(product.ID, device.ID, models.DataOperationTypeReadRsp, propertyID)
	values, err := twin.Read(propertyID)
	if err != nil {
		return err
	}
	o.SetFields(values)
	return d.doc.Publish(*o)
}

// handleWrite is responsible for handling the write request forwarded by the device manager.
// The fields will be written into the real device finally.
//
// This handler could be tested as follows:
// 1. Send the specified format data to the message bus:
//    (a.) Directly (mock):   mosquitto_pub -h 172.16.251.163 -p 1883 -t "v1/DATA/randnum_test01/randnum_test01/write/float" -m "{\"float\": 100}"
//    (b.) Indirectly:        invoke the DeviceManagerDeviceDataOperationClient.Write("randnum_test01", "randnum_test01", "float", 100)
// 2. Observe the log of device driver.
func (d *DeviceDriver) handleWrite(product *models.Product, device *models.Device, twin models.DeviceTwin,
	propertyID string, fields map[models.ProductPropertyID]interface{}) error {
	properties := map[models.ProductPropertyID]*models.ProductProperty{}
	for _, property := range product.Properties {
		properties[property.Id] = property
	}

	var values map[models.ProductPropertyID]interface{}
	if propertyID == models.DeviceDataMultiPropsID {
		tmp := make(map[models.ProductPropertyID]interface{}, len(fields))
		for k, v := range fields {
			property, ok := properties[k]
			if !ok {
				d.logger.Errorf("undefined property: %d", propertyID)
				continue
			}
			if !property.Writeable {
				d.logger.Errorf("the property[%d] is read only", propertyID)
				continue
			}
			tmp[k] = v
		}
		values = tmp
	} else {
		property, ok := properties[propertyID]
		if !ok {
			return fmt.Errorf("undefined property: %d", propertyID)
		}
		if !property.Writeable {
			return fmt.Errorf("the property[%d] is read only", propertyID)
		}

		v, ok := fields[propertyID]
		if !ok {
			return fmt.Errorf("the property[%d]'d value is missed", propertyID)
		}
		values = map[models.ProductPropertyID]interface{}{propertyID: v}
	}
	return twin.Write(propertyID, values)
}

// handleRequest is responsible for handling the method's request forwarded by the device manager.
// The fields will be expanded a request, and
//
// This handler could be tested as follows:
// 1. Send the specified format data to the message bus:
//    (a.) Directly (mock):   mosquitto_pub -h 172.16.251.163 -p 1883 -t "v1/DATA/randnum_test01/randnum_test01/request/Intn" -m "{\"n\": 100}"
//    (b.) Indirectly:        invoke the DeviceManagerDeviceDataOperationClient.Call("randnum_test01", "randnum_test01", "Intn", map[string]interface{}{"n": 100})
// 2. Observe the log of device driver and subscribe the specified topic:
//	  mosquitto_sub -h 172.16.251.163 -p 1883 -t "v1/DATA/randnum_test01/randnum_test01/response/Intn".
func (d *DeviceDriver) handleRequest(product *models.Product, device *models.Device, twin models.DeviceTwin,
	methodID models.ProductMethodID, ins map[string]interface{}) error {
	outs, err := twin.Call(methodID, ins)
	if err != nil {
		return err
	}
	response := models.NewDataOperation(product.ID, device.ID, models.DataOperationTypeResponse, methodID)
	response.SetFields(outs)
	return d.doc.Publish(*response)
}
