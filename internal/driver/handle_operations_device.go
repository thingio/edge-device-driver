package driver

import (
	"fmt"
	"github.com/thingio/edge-device-std/models"
	"github.com/thingio/edge-device-std/msgbus/message"
)

// publishingDeviceOperation tries to publish data in the bus into the MQTTMessageBus.
func (d *DeviceDriver) publishingDeviceOperation() {
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

func (d *DeviceDriver) handlingDeviceOperation() {
	d.wg.Add(1)

	go func() {
		defer d.wg.Done()

		topic := models.TopicTypeDataOperation.Topic()
		if err := d.doc.Subscribe(d.handleDeviceOperation, topic); err != nil {
			d.logger.WithError(err).Errorf("fail to subscribe the topic: %s", topic)
			return
		}
	}()
}

func (d *DeviceDriver) handleDeviceOperation(msg *message.Message) {
	o, err := models.ParseDataOperation(msg)
	if err != nil {
		d.logger.WithError(err).Errorf("fail to parse the message[%s]", msg.ToString())
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
		d.logger.Errorf("unsupported operation type: %s", optType)
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
//    (a.) Directly (mock):   mosquitto_pub -h 172.16.251.163 -p 1883 -t "v1/DATA/randnum_test01/randnum_test01/health-check-ping/randnum_test01" -m "{}"
//    (b.) Indirectly:        invoke the DataOperationManagerClient.Call("randnum_test01", "randnum_test01")
// 2. Observe the log of device driver and subscribe the specified topic:
//	  mosquitto_sub -h 172.16.251.163 -p 1883 -t "v1/DATA/randnum_test01/randnum_test01/health-check-pong/randnum_test01".
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

// handleSoftRead is responsible for handling the soft read request forwarded by the device manager.
// It will read fields from cache in the device twin.
//
// This handler could be tested as follows:
// 1. Send the specified format data to the message bus:
//    (a.) Directly (mock):   mosquitto_pub -h 172.16.251.163 -p 1883 -t "v1/DATA/randnum_test01/randnum_test01/read-req/float" -m "{\"float\": 100}"
//    (b.) Indirectly:        invoke the DataOperationManagerClient.Read("randnum_test01", "randnum_test01", "float", 100)
// 2. Observe the log of device driver and subscribe the specified topic:
//	  mosquitto_sub -h 172.16.251.163 -p 1883 -t "v1/DATA/randnum_test01/randnum_test01/read-rsp/float".
func (d *DeviceDriver) handleSoftRead(product *models.Product, device *models.Device, twin models.DeviceTwin,
	propertyID models.ProductPropertyID, _ map[string]interface{}) error {
	o := models.NewDataOperation(product.ID, device.ID, models.DataOperationTypeSoftReadRsp, propertyID)
	values, err := twin.Read(propertyID)
	if err != nil {
		if values == nil {
			values = make(map[string]interface{})
		}
		values[models.DeviceTwinError] = err.Error()
	}
	o.SetFields(values)
	return d.doc.Publish(*o)
}

// handleHardRead is responsible for handling the hard read request forwarded by the device manager.
// It will read fields from the real device.
//
// This handler could be tested as follows:
// 1. Send the specified format data to the message bus:
//    (a.) Directly (mock):   mosquitto_pub -h 172.16.251.163 -p 1883 -t "v1/DATA/randnum_test01/randnum_test01/hard-read-req/float" -m "{\"float\": 100}"
//    (b.) Indirectly:        invoke the DataOperationManagerClient.Read("randnum_test01", "randnum_test01", "float", 100)
// 2. Observe the log of device driver and subscribe the specified topic:
//	  mosquitto_sub -h 172.16.251.163 -p 1883 -t "v1/DATA/randnum_test01/randnum_test01/hard-read-rsp/float".
func (d *DeviceDriver) handleHardRead(product *models.Product, device *models.Device, twin models.DeviceTwin,
	propertyID models.ProductPropertyID, _ map[string]interface{}) error {
	o := models.NewDataOperation(product.ID, device.ID, models.DataOperationTypeHardReadRsp, propertyID)
	values, err := twin.HardRead(propertyID)
	if err != nil {
		if values == nil {
			values = make(map[string]interface{})
		}
		values[models.DeviceTwinError] = err.Error()
	}
	o.SetFields(values)
	return d.doc.Publish(*o)
}

// handleWrite is responsible for handling the write request forwarded by the device manager.
// The fields will be written into the real device finally.
//
// This handler could be tested as follows:
// 1. Send the specified format data to the message bus:
//    (a.) Directly (mock):   mosquitto_pub -h 172.16.251.163 -p 1883 -t "v1/DATA/randnum_test01/randnum_test01/write-req/float" -m "{\"float\": 100}"
//    (b.) Indirectly:        invoke the DataOperationManagerClient.Write("randnum_test01", "randnum_test01", "float", 100)
// 2. Observe the log of device driver and subscribe the specified topic:
//	  mosquitto_sub -h 172.16.251.163 -p 1883 -t "v1/DATA/randnum_test01/randnum_test01/write-rsp/float".
func (d *DeviceDriver) handleWrite(product *models.Product, device *models.Device, twin models.DeviceTwin,
	propertyID models.ProductPropertyID, fields map[models.ProductPropertyID]interface{}) error {
	var err error
	o := models.NewDataOperation(product.ID, device.ID, models.DataOperationTypeWriteRsp, propertyID)
	results := map[string]interface{}{}
	if fields, err = checkWriteValidity(product, propertyID, fields); err != nil {
		results[models.DeviceTwinError] = err.Error()
	} else if err = twin.Write(propertyID, fields); err != nil {
		results[models.DeviceTwinError] = err.Error()
	}
	o.SetFields(results)
	return d.doc.Publish(*o)
}

func checkWriteValidity(product *models.Product, propertyID string,
	fields map[models.ProductPropertyID]interface{}) (map[models.ProductPropertyID]interface{}, error) {
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
				return nil, fmt.Errorf("undefined property: %s", propertyID)
			}
			if !property.Writeable {
				return nil, fmt.Errorf("the property[%s] is read only", propertyID)
			}
			tmp[k] = v
		}
		values = tmp
	} else {
		property, ok := properties[propertyID]
		if !ok {
			return nil, fmt.Errorf("undefined property: %s", propertyID)
		}
		if !property.Writeable {
			return nil, fmt.Errorf("the property[%s] is read only", propertyID)
		}

		v, ok := fields[propertyID]
		if !ok {
			return nil, fmt.Errorf("the property[%s]'d value is missed", propertyID)
		}
		values = map[models.ProductPropertyID]interface{}{propertyID: v}
	}
	return values, nil
}

// handleRequest is responsible for handling the method's request forwarded by the device manager.
// The fields will be expanded a request, and
//
// This handler could be tested as follows:
// 1. Send the specified format data to the message bus:
//    (a.) Directly (mock):   mosquitto_pub -h 172.16.251.163 -p 1883 -t "v1/DATA/randnum_test01/randnum_test01/request/Intn" -m "{\"n\": 100}"
//    (b.) Indirectly:        invoke the DataOperationManagerClient.Call("randnum_test01", "randnum_test01", "Intn", map[string]interface{}{"n": 100})
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
