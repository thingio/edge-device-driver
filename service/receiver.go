package service

import (
	"encoding/json"
	"fmt"

	"github.com/thingio/edge-device-sdk/clients"
	"github.com/thingio/edge-device-sdk/logger"
	"github.com/thingio/edge-device-sdk/models"
	"github.com/thingio/edge-device-sdk/mqtt"
)

type Receiver interface {
	Start() error
	Stop()
}
type mqttReceiver struct {
	client mqtt.Client
	queue  chan *mqtt.Msg
}

// subTopics 为Debus组件需要处理的MQTT消息类型列表
var subTopics = map[mqtt.TopicType]mqtt.MsgHandler{
	//mqtt.Midmsg:     CommonMsgHandler,
	//mqtt.Display:    CommonMsgHandler,
	mqtt.DeviceData: MQTTDeviceDataHandler,
}

func NewMQTTReceiver(client mqtt.Client) Receiver {
	s := mqttReceiver{
		client: client,
		queue:  make(chan *mqtt.Msg, 100),
	}
	return s
}

func (r mqttReceiver) Start() error {
	r.client = clients.MqttCli
	if !r.client.IsConnected() {

		if err := r.client.Start(); err != nil {
			return err
		}
	}
	for tpc, handler := range subTopics {
		logger.Infof("subscribe mqtt topic:%s", tpc)
		if err := r.client.Sub(handler, tpc.Topic()); err != nil {
			return fmt.Errorf("%e, failed to subscribe mqtt topic: %s", err, tpc)
		}
	}

	return nil
}

func (r mqttReceiver) Stop() {
	close(r.queue)
	r.client.Stop()
}

// ParseMsg 负责反序列化MQTT消息
func ParseMsg(msg *mqtt.Msg) (topic mqtt.Topic, fields map[string]interface{}, err error) {
	logger.Infof("receive msg, topic: %s", topic)

	// parse point tags from mqtt message's topic
	if topic, err = mqtt.NewTopic(msg.Topic); err != nil {
		err = fmt.Errorf("%e, error while parsing mqtt msg topic", err)
		return
	}

	// parse point fields from mqtt message's payload
	fields = make(map[string]interface{}, 0)
	if err = json.Unmarshal(msg.Payload, &fields); err != nil {
		err = fmt.Errorf("%e, error while unmarshal mqtt payload to a map", err)
		return
	}
	return
}

// MQTTDeviceDataHandler 负责处理接收到的MQTT Message形式的设备数据
func MQTTDeviceDataHandler(msg *mqtt.Msg) {
	tpc, fields, err := ParseMsg(msg)
	if err != nil {
		fmt.Printf("failed to preprocess common mqtt message")
		return
	}

	// handle device data before sending it
	// TODO 限制协程数量
	go func() {
		data := models.NewDeviceDataFromMQTTMsg(tpc, fields)
		if err := HandleMQTTDeviceData(&data); err != nil {
			logger.WithError(err).Errorf("error while handling device data mqtt msg")
			// publish a response including error info of device request
			if data.OptType == models.DeviceRequest {
				handlerDeviceReqErr(tpc, err)
			}
		}

	}()
}

// handlerDeviceReqErr 将设备服务调用的错误封装为MQTT Message并进行发送
func handlerDeviceReqErr(tpc mqtt.Topic, err error) {
	err = fmt.Errorf("%e, error while handling device method req: %s", err, tpc.String())
	payload, _ := json.Marshal(map[string]string{models.DeviceError: err.Error()})
	msg := &mqtt.Msg{
		Topic: mqtt.NewDeviceDataTopic(
			tpc.GetValue(mqtt.ProductID),
			tpc.GetValue(mqtt.DeviceID),
			models.DeviceError,
			tpc.GetValue(mqtt.DataID),
		).String(),
		Payload: payload,
	}
	if err := clients.MqttCli.Pub(msg); err != nil {
		logger.WithError(err).Errorf("error while publishing device data handling error")
	}
}
