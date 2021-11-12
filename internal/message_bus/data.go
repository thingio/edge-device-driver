package bus

import "errors"

type Data interface {
	SetFields(fields map[string]interface{})
	GetFields() map[string]interface{}
	SetField(key string, value interface{})
	GetField(key string) interface{}

	ToMessage() (*Message, error)

	Response() (response Data, err error)
}

type data struct {
	fields map[string]interface{}
}

func (d *data) SetFields(fields map[string]interface{}) {
	if d.fields == nil {
		d.fields = make(map[string]interface{})
	}

	for key, value := range fields {
		d.fields[key] = value
	}
}

func (d *data) GetFields() map[string]interface{} {
	return d.fields
}

func (d *data) SetField(key string, value interface{}) {
	if d.fields == nil {
		d.fields = make(map[string]interface{})
	}
	d.fields[key] = value
}

func (d *data) GetField(key string) interface{} {
	return d.fields[key]
}

func (d *data) ToMessage() (*Message, error) {
	return nil, errors.New("implement me")
}

func (d *data) Response() (response Data, err error) {
	return nil, errors.New("implement me")
}
