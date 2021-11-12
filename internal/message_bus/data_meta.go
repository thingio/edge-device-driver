package bus

import (
	"encoding/json"
	"fmt"
)

type (
	MetaDataType = string

	MetaDataOperation     = string
	MetaDataOperationMode = string
)

const (
	MetaDataTypeProtocol MetaDataType = "protocol"
	MetaDataTypeProduct  MetaDataType = "product"
	MetaDataTypeDevice   MetaDataType = "device"

	MetaDataOperationCreate MetaDataOperation = "create"
	MetaDataOperationUpdate MetaDataOperation = "update"
	MetaDataOperationDelete MetaDataOperation = "delete"
	MetaDataOperationGet    MetaDataOperation = "get"
	MetaDataOperationList   MetaDataOperation = "list"

	MetaDataOperationModeRequest  MetaDataOperationMode = "request"
	MetaDataOperationModeResponse MetaDataOperationMode = "response"
)

type MetaData struct {
	data

	MetaType MetaDataType          `json:"meta_type"`
	OptType  MetaDataOperation     `json:"opt_type"`
	OptMode  MetaDataOperationMode `json:"opt_mode"`
	DataID   string                `json:"data_id"`
}

func (d *MetaData) ToMessage() (*Message, error) {
	topic := NewMetaDataTopic(d.MetaType, d.OptType, d.OptMode, d.DataID)
	payload, err := json.Marshal(d.fields)
	if err != nil {
		return nil, err
	}

	return &Message{
		Topic:   topic.String(),
		Payload: payload,
	}, nil
}

func (d *MetaData) isRequest() bool {
	return d.OptMode == MetaDataOperationModeRequest
}

func (d *MetaData) Response() (Data, error) {
	if !d.isRequest() {
		return nil, fmt.Errorf("the device data is not a request: %+v", *d)
	}
	return NewMetaData(d.MetaType, d.OptType, MetaDataOperationModeResponse, d.DataID), nil
}

func NewMetaData(metaType MetaDataType, methodType MetaDataOperation,
	optMode MetaDataOperationMode, dataID string) *MetaData {
	return &MetaData{
		MetaType: metaType,
		OptType:  methodType,
		OptMode:  optMode,
		DataID:   dataID,
	}
}

func ParseMetaData(msg *Message) (*MetaData, error) {
	tags, fields, err := msg.Parse()
	if err != nil {
		return nil, err
	}
	metaData := NewMetaData(tags[0], tags[1], tags[2], tags[3])
	metaData.SetFields(fields)
	return metaData, nil
}
