package bus

import (
	"errors"
)

type Data interface {
	SetFields(fields map[string]interface{})
	GetFields() map[string]interface{}
	SetField(key string, value interface{})
	GetField(key string) interface{}

	ToMessage() (*Message, error)

	Response() (response Data, err error)
}

type MessageData struct {
	Fields map[string]interface{}
}

func (d *MessageData) SetFields(fields map[string]interface{}) {
	for key, value := range fields {
		d.SetField(key, value)
	}
}

func (d *MessageData) GetFields() map[string]interface{} {
	return d.Fields
}

func (d *MessageData) SetField(key string, value interface{}) {
	if d.Fields == nil {
		d.Fields = make(map[string]interface{})
	}
	d.Fields[key] = value
}

func (d *MessageData) GetField(key string) interface{} {
	return d.Fields[key]
}

func (d *MessageData) ToMessage() (*Message, error) {
	return nil, errors.New("implement me")
}

func (d *MessageData) Response() (response Data, err error) {
	return nil, errors.New("implement me")
}
