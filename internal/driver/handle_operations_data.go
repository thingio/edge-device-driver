package driver

import (
	"fmt"
	"github.com/thingio/edge-device-std/errors"
	"github.com/thingio/edge-device-std/models"
	"time"
)

func (d *DeviceDriver) handleDataOperation() error {
	if err := d.ds.ReadHandler(d.protocol.ID, d.handleRead); err != nil {
		return err
	}
	if err := d.ds.HardReadHandler(d.protocol.ID, d.handleHardRead); err != nil {
		return err
	}
	if err := d.ds.WriteHandler(d.protocol.ID, d.handleWrite); err != nil {
		return err
	}
	if err := d.ds.CallHandler(d.protocol.ID, d.handleCall); err != nil {
		return err
	}
	return nil
}

// handleRead is responsible for handling the soft read request forwarded by the device manager.
// It will read fields from cache in the device twin.
//
// This handler could be tested as follows:
// 1. Send the specified format data to the message bus:
//    (a.) Directly (mock):   mosquitto_pub -h 172.16.251.163 -p 1883 -t "DATA/v1/DOWN/randnum/randnum_test01/randnum_test01/float/READ/{ReqID}" -m "{\"float\": 100}"
//    (b.) Indirectly:        invoke the DataManagerClient.Read("randnum_test01", "randnum_test01", "float", 100)
// 2. Observe the log of device driver and subscribe the specified topic:
//	  mosquitto_sub -h 172.16.251.163 -p 1883 -t "DATA/v1/UP/randnum/randnum_test01/randnum_test01/float/READ/{ReqID}".
func (d *DeviceDriver) handleRead(productID, deviceID string, propertyID models.ProductPropertyID) (
	props map[models.ProductPropertyID]*models.DeviceData, err error) {
	twin, err := d.getDeviceTwin(deviceID)
	if err != nil {
		return nil, errors.Internal.Cause(err, "fail to get the device twin[%s]", deviceID)
	}
	props, err = twin.Read(propertyID)
	if err != nil {
		d.logger.WithError(err).Errorf("fail to read softly the property[%s] "+
			"from the device[%s]", propertyID, deviceID)
		return nil, err
	}
	return
}

// handleHardRead is responsible for handling the hard read request forwarded by the device manager.
// It will read fields from the real device.
//
// This handler could be tested as follows:
// 1. Send the specified format data to the message bus:
//    (a.) Directly (mock):   mosquitto_pub -h 172.16.251.163 -p 1883 -t "DATA/v1/DOWN/randnum/randnum_test01/randnum_test01/float/HEAD-READ/{ReqID}" -m "{\"float\": 100}"
//    (b.) Indirectly:        invoke the DataOperationManagerClient.HardRead("randnum_test01", "randnum_test01", "float", 100)
// 2. Observe the log of device driver and subscribe the specified topic:
//	  mosquitto_sub -h 172.16.251.163 -p 1883 -t "DATA/v1/UP/randnum/randnum_test01/randnum_test01/float/HARD-READ/{ReqID}".
func (d *DeviceDriver) handleHardRead(productID, deviceID string, propertyID models.ProductPropertyID) (
	props map[models.ProductPropertyID]*models.DeviceData, err error) {
	twin, err := d.getDeviceTwin(deviceID)
	if err != nil {
		return nil, errors.Internal.Cause(err, "fail to get the device twin[%s]", deviceID)
	}
	return twin.HardRead(propertyID)
}

// handleWrite is responsible for handling the write request forwarded by the device manager.
// The fields will be written into the real device finally.
//
// This handler could be tested as follows:
// 1. Send the specified format data to the message bus:
//    (a.) Directly (mock):   mosquitto_pub -h 172.16.251.163 -p 1883 -t "DATA/v1/DOWN/randnum/randnum_test01/randnum_test01/float/WRITE/{ReqID}" -m "{\"float\": 100}"
//    (b.) Indirectly:        invoke the DataOperationManagerClient.Write("randnum_test01", "randnum_test01", "float", 100)
// 2. Observe the log of device driver and subscribe the specified topic:
//	  mosquitto_sub -h 172.16.251.163 -p 1883 -t "DATA/v1/UP/randnum/randnum_test01/randnum_test01/float/WRITE/{ReqID}".
func (d *DeviceDriver) handleWrite(productID, deviceID string, propertyID models.ProductPropertyID,
	props map[models.ProductPropertyID]*models.DeviceData) (err error) {
	twin, err := d.getDeviceTwin(deviceID)
	if err != nil {
		return errors.Internal.Cause(err, "fail to get the device twin[%s]", deviceID)
	}

	validProps, err := d.filterValidProps(productID, propertyID, props)
	if err != nil {
		return errors.BadRequest.Error(err.Error())
	}
	if err = twin.Write(propertyID, validProps); err != nil {
		d.logger.WithError(err).Errorf("fail to read hardly the property[%s] "+
			"from the device[%s]", propertyID, deviceID)
		return err
	}
	return nil
}

// handleCall is responsible for handling the method's request forwarded by the device manager.
//
// This handler could be tested as follows:
// 1. Send the specified format data to the message bus:
//    (a.) Directly (mock):   mosquitto_pub -h 172.16.251.163 -p 1883 -t "v1/DATA/method/request/randnum_test01/randnum_test01/Intn/{ReqID}" -m "{\"n\": 100}"
//    (b.) Indirectly:        invoke the DataOperationManagerClient.Call("randnum_test01", "randnum_test01", "Intn", map[string]interface{}{"n": 100})
// 2. Observe the log of device driver and subscribe the specified topic:
//	  mosquitto_sub -h 172.16.251.163 -p 1883 -t "v1/DATA/method/response/randnum_test01/randnum_test01/Intn/{ReqID}".
func (d *DeviceDriver) handleCall(productID, deviceID string, methodID models.ProductMethodID,
	ins map[string]*models.DeviceData) (outs map[string]*models.DeviceData, err error) {
	twin, err := d.getDeviceTwin(deviceID)
	if err != nil {
		return nil, errors.Internal.Cause(err, "fail to get the device twin[%s]", deviceID)
	}
	outs, err = twin.Call(methodID, ins)
	if err != nil {
		d.logger.WithError(err).Errorf("fail to call the method[%s] "+
			"of the device[%s]", methodID, deviceID)
		return nil, err
	}
	return outs, nil
}

func (d *DeviceDriver) filterValidProps(productID string,
	propertyID models.ProductPropertyID, props map[models.ProductPropertyID]*models.DeviceData) (
	map[models.ProductPropertyID]*models.DeviceData, error) {
	product, err := d.getProduct(productID)
	if err != nil {
		return nil, err
	}

	properties := map[models.ProductPropertyID]*models.ProductProperty{}
	for _, property := range product.Properties {
		properties[property.Id] = property
	}

	var values map[models.ProductPropertyID]*models.DeviceData
	if propertyID == models.DeviceDataMultiPropsID {
		tmp := make(map[models.ProductPropertyID]*models.DeviceData, len(props))
		for k, v := range props {
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

		v, ok := props[propertyID]
		if !ok {
			return nil, fmt.Errorf("the property[%s]'s value is missed", propertyID)
		}
		values = map[models.ProductPropertyID]*models.DeviceData{propertyID: v}
	}
	return values, nil
}

func (d *DeviceDriver) reportingDevicesHealth() {
	ticker := time.NewTicker(time.Duration(d.cfg.CommonOptions.DeviceHealthCheckIntervalSecond) * time.Second)
	for {
		select {
		case <-ticker.C:
			d.devices.Range(func(_, value interface{}) bool {
				device := value.(*models.Device)
				deviceTwin, err := d.getDeviceTwin(device.ID)
				if err != nil {
					d.logger.WithError(err).Errorf("fail to get the device[%s]'s twin", device.ID)
					return true
				}

				status, err := deviceTwin.HealthCheck()
				if err != nil {
					d.logger.WithError(err).Errorf("fail to check the device[%s]'s health", device.ID)
					return true
				}
				if err = d.dc.PublishDeviceStatus(d.protocol.ID, device.ID, device.ProductID, status); err != nil {
					d.logger.WithError(err).Errorf("fail to publish the status of the driver[%s]", d.protocol.ID)
				} else {
					d.logger.Debugf("success to publish the status of the device[%s]: %+v", device.ID, status)
				}

				return true
			})
		case <-d.ctx.Done():
			ticker.Stop()
			return
		}
	}
}

func (d *DeviceDriver) reportingDevicesData() {
	for {
		select {
		case props := <-d.propsBus:
			if err := d.dc.PublishDeviceProps(d.protocol.ID, props.ProductID, props.DeviceID, props.FuncID, props.Properties); err != nil {
				d.logger.WithError(err).Errorf("fail to publish the device[%s]'s props", props.DeviceID)
			}
		case event := <-d.eventBus:
			if err := d.dc.PublishDeviceEvent(d.protocol.ID, event.ProductID, event.DeviceID, event.FuncID, event.Properties); err != nil {
				d.logger.WithError(err).Errorf("fail to publish the device[%s]'s event", event.DeviceID)
			}
		case <-d.ctx.Done():
			break
		}
	}
}
