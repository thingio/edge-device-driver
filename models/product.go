package models

import (
	"fmt"
	"github.com/thingio/edge-device-sdk/mqtt"
)

// PredefineProductID 为系统中已经预置的设备产品ID
type PredefineProductID string

const (
	// 预定义时序设备产品
	ProductLed               PredefineProductID = "led"
	ProductLightSensor       PredefineProductID = "light_sensor"
	ProductPump              PredefineProductID = "pump"
	ProductServo             PredefineProductID = "servo"
	ProductTemperatureSensor PredefineProductID = "temperature_sensor"
	ProductKM18B90           PredefineProductID = "KM18B90"

	// 预定义多媒体设备产品
	ProductMMGatewaySource PredefineProductID = "mgw_source" // 多媒体网关产品, 无具体产品对应
	ProductFileSource      PredefineProductID = "file_source"
	ProductGB28181Source   PredefineProductID = "gb28181_source"
	ProductHlsSource       PredefineProductID = "hls_source"
	ProductOnvifCamera     PredefineProductID = "onvif_camera"
	ProductRTMPSource      PredefineProductID = "rtmp_source"
	ProductRTSPSource      PredefineProductID = "rtsp_source"
	ProductUsbCamera       PredefineProductID = "usb_camera"
	MMGatewaySourceName    string             = "多媒体网关"
)

type Product struct {
	ID         string             `json:"id"`
	Name       string             `json:"name"`
	Desc       string             `json:"desc"`
	Protocol   string             `json:"protocol"`
	DataFormat string             `json:"data_format"`
	Properties []*ProductProperty `json:"properties"` // 属性功能列表
	Events     []*ProductEvent    `json:"events"`     // 事件功能列表
	Methods    []*ProductMethod   `json:"methods"`    // 服务功能列表
	Topics     []*ProductTopic    `json:"topics"`     // 各功能对应的消息主题
}

const (
	TopicTypePub       = "发布"
	TopicTypeSub       = "订阅"
	TopicOptRead       = "read"
	TopicOptWrite      = "write"
	TopicOptEvent      = "event"
	TopicOptRequest    = "request"
	TopicOptResponse   = "response"
	TopicGeneralDevice = "${DeviceID}"
	TopicDataMultiProp = "*"
)

// SetTopics 将设置产品中所有功能的TOPIC : /${ProductID}/${DeviceID}/${OptType}/${DataID}
func (p *Product) SetTopics() {
	topics := make([]*ProductTopic, 0)
	newTopic := func(optType, dataID string) string {
		return mqtt.NewDeviceDataTopic(p.ID, TopicGeneralDevice, optType, dataID).String()
	}
	if len(p.Properties) > 0 {
		if p.Protocol == "mqtt" {
			topics = append(topics, &ProductTopic{
				Topic:   newTopic(TopicOptRead, TopicDataMultiProp),
				OptType: TopicTypePub,
				Desc:    "设备批量上报数据",
			})
			topics = append(topics, &ProductTopic{
				Topic:   newTopic(TopicOptWrite, TopicDataMultiProp),
				OptType: TopicTypeSub,
				Desc:    "批量修改设备数据",
			})
		}
		for _, p := range p.Properties {
			topics = append(topics, &ProductTopic{
				Topic:   newTopic(TopicOptRead, p.Id),
				OptType: TopicTypePub,
				Desc:    fmt.Sprintf("设备上传属性(%s)数据", p.Id),
			})
			if p.Writeable {
				topics = append(topics, &ProductTopic{
					Topic:   newTopic(TopicOptWrite, p.Id),
					OptType: TopicTypeSub,
					Desc:    fmt.Sprintf("修改设备属性(%s)数据", p.Id),
				})
			}
		}
	}

	for _, e := range p.Events {
		topics = append(topics, &ProductTopic{
			Topic:   newTopic(TopicOptEvent, e.Id),
			OptType: TopicTypePub,
			Desc:    fmt.Sprintf("设备上报事件(%s)数据", e.Id),
		})
	}

	for _, m := range p.Methods {
		topics = append(topics, &ProductTopic{
			Topic:   newTopic(TopicOptRequest, m.Id),
			OptType: TopicTypeSub,
			Desc:    fmt.Sprintf("发送设备服务(%s)调用请求", m.Id),
		})
		topics = append(topics, &ProductTopic{
			Topic:   newTopic(TopicOptResponse, m.Id),
			OptType: TopicTypePub,
			Desc:    fmt.Sprintf("设备返回服务(%s)请求结果", m.Id),
		})
	}
	p.Topics = topics
}

type ProductProperty struct {
	Id         string            `json:"id"`
	Name       string            `json:"name"`
	Desc       string            `json:"desc"`
	Interval   string            `json:"interval"`
	Unit       string            `json:"unit"`
	FieldType  string            `json:"field_type"`
	ReportMode string            `json:"report_mode"`
	Writeable  bool              `json:"writeable"`
	AuxProps   map[string]string `json:"aux_props"`
}

type ProductEvent struct {
	Id       string            `json:"id"`
	Name     string            `json:"name"`
	Desc     string            `json:"desc"`
	Outs     []*ProductField   `json:"outs"`
	AuxProps map[string]string `json:"aux_props"`
}

type ProductMethod struct {
	Id       string            `json:"id"`
	Name     string            `json:"name"`
	Desc     string            `json:"desc"`
	Ins      []*ProductField   `json:"ins"`
	Outs     []*ProductField   `json:"outs"`
	AuxProps map[string]string `json:"aux_props"`
}

type ProductTopic struct {
	Topic   string `json:"topic"`
	OptType string `json:"opt_type"`
	Desc    string `json:"desc"`
}

type ProductField struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	FieldType string `json:"field_type"`
	Desc      string `json:"desc"`
}
