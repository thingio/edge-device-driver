package service

import (
	"fmt"
	"github.com/thingio/edge-device-sdk/logger"
	"github.com/thingio/edge-device-sdk/models"
)


// DeviceDataHandler 为设备特定的产品功能的数据处理函数
type DeviceDataHandler func(product *models.Product, data *models.DeviceData) error

// deviceDataHandlers 为目前支持处理的设备数据类型
var deviceDataHandlers = map[models.DeviceOperation]DeviceDataHandler{
	models.DeviceRead:     handlePropData,     //属性(prop)读取结果
	models.DeviceEvent:    handleEventData,    //事件(event)到达
	models.DeviceWrite:    handleWriteData,    //属性(prop)写入请求
	models.DeviceRequest:  handleRequestData,  //服务(method)调用请求
	models.DeviceResponse: handleResponseData, //服务(method)调用响应
	models.DeviceError:    handleResponseData, //服务(method)调用异常返回
}

// HandleMQTTDeviceData 处理接收到的设备数据, 对相关的数据字段进行校验并负责后续的[属性写入/方法调用]的执行
func HandleMQTTDeviceData(data *models.DeviceData) error {
	// get the definition of device's product
	product, err := getProduct(data.ProductID)
	if err != nil {
		return err
	}

	handler, ok := deviceDataHandlers[data.OptType]
	if !ok {
		return fmt.Errorf("error while handling device data : opt_type %s not supported", data.OptType)
	}
	return handler(product, data)
}

// handlePropData 处理指定设备功能属性的值数据
func handlePropData(product *models.Product, data *models.DeviceData) error {
	// do not filter multi props
	if data.FuncID == models.DeviceDataMultiPropsID {
		return nil
	}

	for _, prop := range product.Properties {
		if data.FuncID != prop.Id {
			continue
		}
		// corresponding property found
		data.Fields = map[string]interface{}{
			data.FuncID: data.Fields[data.FuncID],
		}
		return nil
	}
	// corresponding property not found
	return fmt.Errorf("error while handling device prop data, property %s not found ", data.FuncID)
}

// handleWriteData 处理设置指定设备功能属性的请求
func handleWriteData(product *models.Product, data *models.DeviceData) error {
	// get device driver
	d, err := getDvsConn(data.DeviceID)
	if err != nil {
		return err
	}

	pIds := make(map[string]*models.ProductProperty, len(product.Properties))
	for _, prop := range product.Properties {
		pIds[prop.Id] = prop
	}

	// multi write
	var writeData interface{}
	if data.FuncID == models.DeviceDataMultiPropsID {
		props := make(map[models.ProductPropID]interface{}, len(data.Fields))
		for k, v := range data.Fields {
			p, ok := pIds[k]
			if !ok {
				logger.Warnf("property [%s] not found", k)
				continue
			}
			if !p.Writeable {
				logger.Warnf("property [%s] is read only", k)
				continue
			}
			props[k] = v
		}
		writeData = props
	} else {
		v, ok := data.Fields[data.FuncID]
		if !ok {
			return fmt.Errorf("error while handling device prop write opt, value of %s not found ", data.FuncID)
		}
		writeData = v
	}

	if err := d.Write(data.FuncID, writeData); err != nil {
		return err
	}
	return nil
}

// handleEventData 处理指定设备上报的事件数据
func handleEventData(product *models.Product, data *models.DeviceData) error {
	eIds := make(map[string]struct{}, len(product.Events))
	found := false
	for _, evt := range product.Events {
		if evt.Id != data.FuncID {
			continue
		}
		found = true
		for _, o := range evt.Outs {
			eIds[o.Id] = struct{}{}
		}
		break
	}
	if !found {
		return fmt.Errorf("error while handling device event data, event %s not found ", data.FuncID)
	}

	// delete invalid Fields
	for fid := range data.Fields {
		if _, ok := eIds[fid]; ok {
			continue
		}
		logger.Warnf("field %s is not expected,it will be deleted", fid)
		delete(data.Fields, fid)
	}
	return nil
}

// handleRequestData 处理指定设备的功能服务的调用请求
func handleRequestData(product *models.Product, data *models.DeviceData) (err error) {
	var rsp models.DeviceData
	defer func() {
		if err != nil {
			rsp = models.NewDeviceData(data.ProductID, data.DeviceID, data.FuncID, models.DeviceError)
			rsp.SetField(models.DeviceError, err)
		}
		err = rsp.Pub()
	}()

	conn, err := getDvsConn(data.DeviceID)
	if err != nil {
		return err
	}

	// find method's define
	inIds := make(map[string]struct{}, 0)
	found := false
	for _, m := range product.Methods {
		if m.Id != data.FuncID {
			continue
		}
		found = true
		for _, i := range m.Ins {
			inIds[i.Id] = struct{}{}
		}
		break
	}

	// method not found
	if !found {
		return fmt.Errorf("error while handling device method data, method %s not found ", data.FuncID)
	}

	// check params
	for in := range inIds {
		if _, ok := data.Fields[in]; !ok {
			return fmt.Errorf("error while call method %s, input params %v are necessary", data.FuncID, inIds)
		}
	}

	// delete invalid Fields
	for fid := range data.Fields {
		if _, ok := inIds[fid]; ok {
			continue
		}
		logger.Warnf("field %s is not expected, it will be deleted", fid)
		delete(data.Fields, fid)
	}

	rsp, err = conn.Call(data.FuncID, *data)
	if err != nil {
		return err
	}
	rsp.SetFields(data.Fields) // add request param to response
	return nil
}

// handleResponseData 处理指定设备的功能服务的响应结果
func handleResponseData(product *models.Product, data *models.DeviceData) error {
	ioIds := make(map[string]struct{}, 0)
	found := false
	for _, m := range product.Methods {
		if m.Id != data.FuncID {
			continue
		}
		found = true
		for _, i := range m.Ins {
			ioIds[i.Id] = struct{}{}
		}
		for _, o := range m.Outs {
			ioIds[o.Id] = struct{}{}
		}
	}
	if !found {
		return fmt.Errorf("error while handling device method data, method %s not found ", data.FuncID)
	}

	// delete invalid Fields
	for fid := range data.Fields {
		if _, ok := ioIds[fid]; ok || fid == models.DeviceError {
			continue
		}
		logger.Warnf("field %s is not expected, it will be deleted", fid)
		delete(data.Fields, fid)
	}
	return nil
}
