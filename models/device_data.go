package models

import (
	"encoding/json"
	"fmt"
	"github.com/thingio/edge-device-sdk/clients"
	"github.com/thingio/edge-device-sdk/mqtt"
	"strconv"
)

type DeviceData struct {
	DeviceID  string
	ProductID string
	OptType   DeviceOperation // 设备数据类型,包括 read, write, event, request, response
	FuncID    ProductFuncID   // 设备功能ID, 属性,事件,服务的ID
	FuncType  ProductFuncType
	Fields    map[string]interface{}
}

// Opts2FuncType 操作类型与设备数据的类型的对应关系:
// * read     : props(属性数据)
// * event    : events(事件数据)
// * response : methods(调用数据)
// 其余情况返回空字符串
var Opts2FuncType = map[DeviceOperation]ProductFuncType{
	DeviceEvent:    EventFunc,
	DeviceRead:     PropFunc,
	DeviceResponse: MethodFunc,
}

func NewDeviceData(productID, deviceID, dataID, optType string) DeviceData {
	return DeviceData{
		DeviceID:  deviceID,
		ProductID: productID,
		OptType:   optType,
		FuncID:    dataID,
		FuncType:  Opts2FuncType[optType],
		Fields:    make(map[string]interface{}, 0),
	}
}

func NewDeviceDataFromMQTTMsg(tpc mqtt.Topic, fields map[string]interface{}) DeviceData {
	tags := tpc.Tags()
	data := NewDeviceData(tags[mqtt.ProductID], tags[mqtt.DeviceID], tags[mqtt.DataID], tags[mqtt.OptType])
	data.Fields = fields
	return data
}

func (d DeviceData) GetValue(key string) interface{} {
	return d.Fields[key]
}

func (d DeviceData) ToMQTTMsg() (*mqtt.Msg, error) {
	if d.ProductID == "" || d.DeviceID == "" || d.OptType == "" || d.FuncID == "" {
		return nil, fmt.Errorf("error while publishing mqtt msg, incomplete topic")
	}
	payload, err := json.Marshal(d.Fields)
	if err != nil {
		return nil, err
	}
	return &mqtt.Msg{
		Topic:   mqtt.NewDeviceDataTopic(d.ProductID, d.DeviceID, d.OptType, d.FuncID).String(),
		Payload: payload}, nil
}

func (d DeviceData) MustToMQTTMsg() *mqtt.Msg {
	msg, err := d.ToMQTTMsg()
	if err != nil {
		panic(err)
	}
	return msg
}

func (d DeviceData) Pub() error {
	msg, err := d.ToMQTTMsg()
	if err != nil {
		return err
	}
	return clients.MqttCli.Pub(msg)
}

func (d DeviceData) SetField(key string, value interface{}) {
	d.Fields[key] = value
}

func (d DeviceData) SetFields(kvs map[string]interface{}) {
	for k, v := range kvs {
		d.Fields[k] = v
	}
}

func (d DeviceData) parseDeviceParams(params map[string]string, fields []*ProductField) error {
	kvs := make(map[string]string, len(params))
	for k, v := range params {
		kvs[k] = v
	}
	for _, f := range fields {
		value, ok := kvs[f.Id]
		if !ok {
			continue
		}
		var v interface{}
		var err error
		switch f.FieldType {
		case DeviceDataFieldTypeBool:
			v, err = strconv.ParseBool(value)
		case DeviceDataFieldTypeFloat:
			v, err = strconv.ParseFloat(value, 64)
		case DeviceDataFieldTypeInt:
			v, err = strconv.ParseInt(value, 10, 64)
		case DeviceDataFieldTypeUint:
			v, err = strconv.ParseUint(value, 10, 64)
		case DeviceDataFieldTypeString:
			v = value
			err = nil
		default:
			return fmt.Errorf("error while add key-values, filed type %s is not supported", f.FieldType)
		}
		if err != nil {
			return fmt.Errorf("%e,error while parsing device data, value %s cannot be parsed to %s", err, v, f.FieldType)
		}
		delete(kvs, f.Id)
		d.SetField(f.Id, v)
	}

	if len(kvs) != 0 {
		return fmt.Errorf("error while parsing device data, could not find field %+v in product", kvs)
	}

	return nil
}

func unmarshalFields(payload []byte) (map[string]interface{}, error) {
	fields := make(map[string]interface{})
	err := json.Unmarshal(payload, &fields)
	if err != nil {
		return fields, fmt.Errorf("%e,error unmarshal mqtt payload %s", err, string(payload))
	}
	return fields, nil
}
