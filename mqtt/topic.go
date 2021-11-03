package mqtt

import (
	"fmt"
	"strings"
)

const (
	DeviceData   TopicType = "DATA"

	TopicSep = "/"

	Time       TopicKey = "time"       // 落库时间
	NodeID     TopicKey = "node_id"    // 规则实例数据展示算子定义的NodeID
	DataID     TopicKey = "data_id"    // 设备数据(属性,时间,服务)的ID
	OptType    TopicKey = "opt_type"   // 设备数据类型
	MidmsgID   TopicKey = "msg_id"     // 中间消息的MsgID
	InstanceID TopicKey = "instance"   // 实例ID
	PipelineID TopicKey = "pipeline"   // 规则ID
	ProductID  TopicKey = "product_id" // 产品ID
	DeviceID   TopicKey = "device_id"  // 设备ID
	RefType    TopicKey = "ref_type"   // 相关资源类型
	RefID      TopicKey = "ref_id"     // 相关资源ID
	EdgeID     TopicKey = "edge_id"    // 资源所在边缘节点的ID
)

// TopicKey是MQTT消息的Topic中的特定字段, 依据不同的Topic其具备的TopicKey也不同
// 每个TopicKey都有其固定的位置, 不同的Key之间使用分隔符进行分割.
// 当存储mqtt消息至InfluxDB时, 每个TopicKey 对应一个 TagKey
type TopicKey = string

var TopicKeys = map[TopicType][]TopicKey{


	// 设备消息主题, 用以传输设备上报的数据以及用户下发的指令等
	// DATA/{product_id}/{device_id}/{opt_type}/{data_id}
	DeviceData: {ProductID, DeviceID, OptType, DataID},

}

type TopicType string

type DeviceOptType string

type Topic interface {
	String() string
	Tags() map[TopicKey]string
	GetValue(TopicKey) string
	Type() TopicType
}

type commonTopic struct {
	topicType TopicType
	tags      map[TopicKey]string
}

func NewTopic(topic string) (Topic, error) {
	ts := strings.Split(topic, TopicSep)
	if len(ts) == 0 {
		return nil, fmt.Errorf("could not handle mqtt topic '%s'", topic)
	}
	topicType := TopicType(ts[0])
	keys, ok := TopicKeys[topicType]
	if !ok {
		return nil, fmt.Errorf("undefined mqtt topic '%s'", topic)
	}
	if len(ts)-1 != len(keys) {
		return nil, fmt.Errorf("invalid mqtt topic '%s', keys [%+v] are necessary", topic, keys)
	}
	tags := make(map[TopicKey]string)
	for i, key := range keys {
		tags[key] = ts[i+1]
	}
	return &commonTopic{
		topicType: topicType,
		tags:      tags,
	}, nil
}


func NewDeviceDataTopic(productID, deviceID, optType, dataID string) Topic {
	return &commonTopic{
		topicType: DeviceData,
		tags: map[TopicKey]string{
			DataID:    dataID,
			OptType:   optType,
			DeviceID:  deviceID,
			ProductID: productID,
		},
	}
}

func (c *commonTopic) Type() TopicType {
	return c.topicType
}

func (c *commonTopic) String() string {
	keys := TopicKeys[c.topicType]
	values := make([]string, len(keys)+1)
	values[0] = string(c.topicType)
	for i, key := range keys {
		values[i+1] = c.tags[key]
	}
	return strings.Join(values, TopicSep)
}


func (c *commonTopic) Tags() map[TopicKey]string {
	return c.tags
}

func (c *commonTopic) GetValue(key TopicKey) string {
	return c.tags[key]
}

// Topic returns the topic that could be used to subscribe all messages of this type
func (t TopicType) Topic() string {
	return string(t) + "/#"
}
