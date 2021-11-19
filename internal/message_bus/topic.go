package bus

import (
	"fmt"
	"github.com/thingio/edge-device-sdk/version"
	"strings"
)

const (
	TagsOffset = 2
)

type TopicTagKey string

type TopicTags map[TopicTagKey]string

type TopicType string

// Topic returns the topic that could subscribe all messages belong to this type.
func (t TopicType) Topic() string {
	return strings.Join([]string{version.Version, string(t), TopicWildcard}, TopicSep)
}

const (
	TopicTagKeyMetaDataType       TopicTagKey = "meta_type"
	TopicTagKeyMetaDataMethodType TopicTagKey = "method_type"
	TopicTagKeyMetaDataMethodMode TopicTagKey = "method_mode"
	TopicTagKeyProductID          TopicTagKey = "product_id"
	TopicTagKeyDeviceID           TopicTagKey = "device_id"
	TopicTagKeyOptType            TopicTagKey = "opt_type"
	TopicTagKeyDataID             TopicTagKey = "data_id"

	TopicTypeMetaData   TopicType = "META"
	TopicTypeDeviceData TopicType = "DATA"

	TopicSep      = "/"
	TopicWildcard = "#"
)

// Schemas describes all topics' forms. Every topic is formed by <TopicType>/<TopicTag1>/.../<TopicTagN>.
var Schemas = map[TopicType][]TopicTagKey{
	TopicTypeMetaData:   {TopicTagKeyMetaDataType, TopicTagKeyMetaDataMethodType, TopicTagKeyMetaDataMethodMode, TopicTagKeyDataID},
	TopicTypeDeviceData: {TopicTagKeyProductID, TopicTagKeyDeviceID, TopicTagKeyOptType, TopicTagKeyDataID},
}

type Topic interface {
	Type() TopicType

	String() string

	Tags() TopicTags
	TagKeys() []TopicTagKey
	TagValues() []string
	TagValue(key TopicTagKey) (value string, ok bool)
}

type commonTopic struct {
	topicTags TopicTags
	topicType TopicType
}

func (c *commonTopic) Type() TopicType {
	return c.topicType
}

func (c *commonTopic) String() string {
	topicType := string(c.topicType)
	tagValues := c.TagValues()
	return strings.Join(append([]string{version.Version, topicType}, tagValues...), TopicSep)
}

func (c *commonTopic) Tags() TopicTags {
	return c.topicTags
}

func (c *commonTopic) TagKeys() []TopicTagKey {
	return Schemas[c.topicType]
}

func (c *commonTopic) TagValues() []string {
	tagKeys := c.TagKeys()
	values := make([]string, len(tagKeys))
	for idx, topicTagKey := range tagKeys {
		values[idx] = c.topicTags[topicTagKey]
	}
	return values
}

func (c *commonTopic) TagValue(key TopicTagKey) (value string, ok bool) {
	tags := c.Tags()
	value, ok = tags[key]
	return
}

func NewTopic(topic string) (Topic, error) {
	parts := strings.Split(topic, TopicSep)
	if len(parts) < TagsOffset {
		return nil, fmt.Errorf("invalid topic: %s", topic)
	}
	topicType := TopicType(parts[1])
	keys, ok := Schemas[topicType]
	if !ok {
		return nil, fmt.Errorf("undefined topic type: %s", topicType)
	}
	if len(parts)-TagsOffset != len(keys) {
		return nil, fmt.Errorf("invalid topic: %s, keys [%+v] are necessary", topic, keys)
	}

	tags := make(map[TopicTagKey]string)
	for i, key := range keys {
		tags[key] = parts[i+TagsOffset]
	}
	return &commonTopic{
		topicType: topicType,
		topicTags: tags,
	}, nil
}

func NewMetaDataTopic(metaType, methodType, methodMode, dataID string) Topic {
	return &commonTopic{
		topicTags: map[TopicTagKey]string{
			TopicTagKeyMetaDataType:       metaType,
			TopicTagKeyMetaDataMethodType: methodType,
			TopicTagKeyMetaDataMethodMode: methodMode,
			TopicTagKeyDataID:             dataID,
		},
		topicType: TopicTypeMetaData,
	}
}

func NewDeviceDataTopic(productID, deviceID, optType, dataID string) Topic {
	return &commonTopic{
		topicTags: map[TopicTagKey]string{
			TopicTagKeyProductID: productID,
			TopicTagKeyDeviceID:  deviceID,
			TopicTagKeyOptType:   optType,
			TopicTagKeyDataID:    dataID,
		},
		topicType: TopicTypeDeviceData,
	}
}
