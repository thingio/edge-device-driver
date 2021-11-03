package client

import (
	"encoding/json"
	"fmt"
	"github.com/thingio/edge-device-sdk/models"
	"github.com/thingio/edge-device-sdk/mqtt"
	"time"
)

// FunctionClient 为设备产品中定义的功能相应的操作接口
type FunctionClient interface {
	// Read 读指定属性的值
	Read(productID, deviceID, propertyID string) (models.DeviceData, error)
	// Write 写指定属性的值
	Write(productID, deviceID, propertyID string, value interface{}) error
	// Receive 接收指定的事件数据
	Receive(productID, deviceID, eventID string) (chan models.DeviceData, error)
	// Call 调用指定的服务功能
	Call(productID, deviceID, methodID string, ins map[string]interface{}) (map[string]interface{}, error)
}

type ProductClient interface {
	// CreateProduct 新增设备产品(物模型)
	CreateProduct(product models.Product) error
	// UpdateProduct 新增设备产品(物模型)
	UpdateProduct(product models.Product) error
	// DeleteProduct 删除设备产品(物模型)及其下属的所有已激活设备
	DeleteProduct(productID string) error
}

type ProtocolClient interface {
	// GetProtocol 获取设备协议定义
	GetProtocol() (models.Protocol, error)
}

type DeviceClient interface {
	// ListDevices 返回所有已激活设备
	ListDevices() ([]*models.Device, error)
	// ActivateDevice 激活给定的设备
	ActivateDevice(device *models.Device) error
	// DeactivateDevice 下线指定的设备
	DeactivateDevice(deviceID string) error
}

type DriverClient interface {
	ProtocolClient
	ProductClient
	DeviceClient
	FunctionClient
}

type mqttDeviceCli struct {
	c           mqtt.Client
	callTimeout time.Duration
}

func (m *mqttDeviceCli) Write(productID, deviceID, propertyID string, value interface{}) error {
	d := models.NewDeviceData(productID, deviceID, propertyID, models.DeviceWrite)
	d.SetField(propertyID, value)
	msg, err := d.ToMQTTMsg()
	if err != nil {
		return err
	}
	return m.c.Pub(msg)
}

func (m *mqttDeviceCli) Call(productID, deviceID, methodID string, ins map[string]interface{}) (map[string]interface{}, error) {
	d := models.NewDeviceData(productID, deviceID, methodID, models.DeviceRequest)
	d.Fields = ins
	res, err := m.PubAndRecv(d)
	if err != nil {
		return nil, err
	}
	return res.Fields, nil
}

func (m *mqttDeviceCli) PubAndRecv(req models.DeviceData) (models.DeviceData, error) {
	rsp := models.NewDeviceData(req.ProductID, req.DeviceID, req.FuncID, models.DeviceResponse)
	if req.OptType != models.DeviceRequest {
		return rsp, fmt.Errorf("error while pub and recv device data, opt type : %s not supported", req.OptType)
	}

	// convert request to mqtt message
	reqMsg, err := req.ToMQTTMsg()
	if err != nil {
		return rsp, err
	}

	// subscribe response or error
	rspTpc := mqtt.NewDeviceDataTopic(req.ProductID, req.DeviceID, models.DeviceResponse, req.FuncID)
	errTpc := mqtt.NewDeviceDataTopic(req.ProductID, req.DeviceID, models.DeviceError, req.FuncID)
	tps := []string{rspTpc.String(), errTpc.String()}
	ch := make(chan *mqtt.Msg, 1)
	m.c.AddRoute(func(msg *mqtt.Msg) {
		ch <- msg
	}, tps...)
	defer func() {
		m.c.RemoveRoute(tps...)
		close(ch)
	}()

	// publish call method request
	if err = m.c.Pub(reqMsg); err != nil {
		return rsp, fmt.Errorf("%e, error while publishing device service reqMsg", err)
	}

	tk := time.NewTicker(m.callTimeout)
	defer tk.Stop()
	select {
	case rspMsg := <-ch:
		topic, err := mqtt.NewTopic(rspMsg.Topic)
		if err != nil {
			return rsp, err
		}
		fields := make(map[string]interface{})
		if err := json.Unmarshal(rspMsg.Payload, &fields); err != nil {
			return rsp, fmt.Errorf("%e,error unmarshal mqtt payload %s", err, string(rspMsg.Payload))
		}
		if topic.GetValue(mqtt.OptType) == models.DeviceError {
			return rsp, fmt.Errorf("error while call method %s of device %s of product %s, %s",
				req.FuncID, req.DeviceID, req.ProductID, fields[models.DeviceError].(string))
		}
		rsp.SetFields(fields)
	case <-tk.C:
		return rsp, fmt.Errorf("timeout while call method %s of device %s of product %s", req.FuncID, req.DeviceID, req.ProductID)
	}
	return rsp, nil
}
