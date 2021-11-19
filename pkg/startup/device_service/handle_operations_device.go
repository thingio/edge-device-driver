package startup

import (
	"fmt"
	bus "github.com/thingio/edge-device-sdk/internal/message_bus"
	"github.com/thingio/edge-device-sdk/pkg/models"
)

// publishingDeviceData tries to publish data in the bus into the MessageBus.
func (s *DeviceService) publishingDeviceData() {
	s.wg.Add(1)

	go func() {
		defer s.wg.Done()

		for {
			select {
			case data := <-s.bus:
				if err := s.doc.Publish(data); err != nil {
					s.logger.WithError(err).Errorf("fail to publish the device data %+v", data)
				} else {
					s.logger.Debugf("success to publish the device data %+v", data)
				}
			case <-s.ctx.Done():
				break
			}
		}
	}()
}

// TODO 是否可以将 DeviceDataHandler 转移到 DeviceServiceDeviceDataOperationClient 中？

type DeviceDataHandler func(product *models.Product, device *models.Device, conn models.DeviceConnector,
	dataID string, fields map[string]interface{}) error

func (s *DeviceService) handlingDeviceData() {
	s.wg.Add(1)

	s.unsupportedDeviceDataTypes = map[models.DeviceDataOperation]struct{}{
		models.DeviceDataOperationRead:     {}, // property read
		models.DeviceDataOperationEvent:    {}, // event
		models.DeviceDataOperationResponse: {}, // method response
		models.DeviceDataOperationError:    {}, // method error
	}
	s.deviceDataHandlers = map[models.DeviceDataOperation]DeviceDataHandler{
		models.DeviceDataOperationWrite:   s.handleWriteData,   // property write
		models.DeviceDataOperationRequest: s.handleRequestData, // method request
	}

	go func() {
		defer s.wg.Done()

		topic := bus.TopicTypeDeviceData.Topic()
		if err := s.doc.Subscribe(s.handleDeviceData, topic); err != nil {
			s.logger.WithError(err).Errorf("fail to subscribe the topic: %s", topic)
			return
		}
	}()
}

func (s *DeviceService) handleDeviceData(msg *bus.Message) {
	tags, fields, err := msg.Parse()
	if err != nil {
		s.logger.WithError(err).Errorf("fail to parse the message[%s]", msg.ToString())
		return
	}

	productID, deviceID, optType, dataID := tags[0], tags[1], tags[2], tags[3]
	product, err := s.getProduct(productID)
	if err != nil {
		s.logger.Error(err.Error())
		return
	}
	device, err := s.getDevice(deviceID)
	if err != nil {
		s.logger.Error(err.Error())
		return
	}
	connector, err := s.getDeviceConnector(deviceID)
	if err != nil {
		s.logger.Error(err.Error())
		return
	}

	if _, ok := s.unsupportedDeviceDataTypes[optType]; ok {
		return
	}
	handler, ok := s.deviceDataHandlers[optType]
	if !ok {
		s.logger.Errorf("unsupported operation type: %s", optType)
		return
	}
	if err = handler(product, device, connector, dataID, fields); err != nil {
		s.logger.WithError(err).Errorf("fail to handle the message: %+v", msg)
		return
	}
}

// handleWriteData is responsible for handling the write request forwarded by the device manager.
// The fields will be written into the real device finally.
//
// This handler could be tested as follows:
// 1. Send the specified format data to the message bus:
//    (a.) Directly (mock):   mosquitto_pub -h 172.16.251.163 -p 1883 -t "v1/DATA/randnum_test01/randnum_test01/write/float" -m "{\"intf\": 100}"
//    (b.) Indirectly:        invoke the DeviceManagerDeviceDataOperationClient.Write("randnum_test01", "randnum_test01", "float", 100)
// 2. Observe the log of device service.
func (s *DeviceService) handleWriteData(product *models.Product, device *models.Device, conn models.DeviceConnector,
	propertyID string, fields map[string]interface{}) error {
	properties := map[models.ProductPropertyID]*models.ProductProperty{}
	for _, property := range product.Properties {
		properties[property.Id] = property
	}

	var written interface{}
	if propertyID == models.DeviceDataMultiPropsID {
		tmp := make(map[models.ProductPropertyID]interface{}, len(fields))
		for k, v := range fields {
			property, ok := properties[k]
			if !ok {
				s.logger.Errorf("undefined property: %s", propertyID)
				continue
			}
			if !property.Writeable {
				s.logger.Errorf("the property[%s] is read only", propertyID)
				continue
			}
			tmp[k] = v
		}
		written = tmp
	} else {
		property, ok := properties[propertyID]
		if !ok {
			return fmt.Errorf("undefined property: %s", propertyID)
		}
		if !property.Writeable {
			return fmt.Errorf("the property[%s] is read only", propertyID)
		}

		v, ok := fields[propertyID]
		if !ok {
			return fmt.Errorf("the property[%s]'s value is missed", propertyID)
		}
		written = map[models.ProductPropertyID]interface{}{propertyID: v}
	}
	return conn.Write(propertyID, written)
}

// handleRequestData is responsible for handling the method's request forwarded by the device manager.
// The fields will be expanded a request, and
//
// This handler could be tested as follows:
// 1. Send the specified format data to the message bus:
//    (a.) Directly (mock):   mosquitto_pub -h 172.16.251.163 -p 1883 -t "v1/DATA/randnum_test01/randnum_test01/request/Intn" -m "{\"n\": 100}"
//    (b.) Indirectly:        invoke the DeviceManagerDeviceDataOperationClient.Call("randnum_test01", "randnum_test01", "Intn", map[string]interface{}{"n": 100})
// 2. Observe the log of device service and subscribe the specified topic:
//	  mosquitto_sub -h 172.16.251.163 -p 1883 -t "v1/DATA/randnum_test01/randnum_test01/response/Intn".
func (s *DeviceService) handleRequestData(product *models.Product, device *models.Device, conn models.DeviceConnector,
	methodID string, fields map[string]interface{}) error {
	request := models.NewDeviceData(product.ID, device.ID, models.DeviceDataOperationRequest, methodID)
	request.SetFields(fields)
	response, err := conn.Call(methodID, *request)
	if err != nil {
		return err
	}
	if err = s.doc.Publish(response); err != nil {
		return err
	}
	return nil
}
