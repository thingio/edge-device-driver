package operations

import "encoding/json"

type OperationMessage interface {
	Unmarshal(fields map[string]interface{}) error
	Marshal() (map[string]interface{}, error)
}

func Struct2Map(o interface{}) (map[string]interface{}, error) {
	data, err := json.Marshal(o)
	if err != nil {
		return nil, err
	}
	fields := make(map[string]interface{})
	if err = json.Unmarshal(data, &fields); err != nil {
		return nil, err
	}
	return fields, nil
}

func Map2Struct(m map[string]interface{}, s interface{}) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(data, s); err != nil {
		return err
	}
	return nil
}
