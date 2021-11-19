package bus

import (
	"encoding/json"
	"fmt"
)

type MessageHandler func(msg *Message)

// Message is an intermediate data format between MQTT and Data.
type Message struct {
	Topic   string
	Payload []byte
}

func (m *Message) Parse() ([]string, map[string]interface{}, error) {
	// parse topic
	topic, err := NewTopic(m.Topic)
	if err != nil {
		return nil, nil, err
	}
	tagValues := topic.TagValues()

	// parse payload
	fields := make(map[string]interface{})
	if err := json.Unmarshal(m.Payload, &fields); err != nil {
		return nil, nil, err
	}
	return tagValues, fields, nil
}

func (m *Message) ToString() string {
	return fmt.Sprintf("%+v", *m)
}
