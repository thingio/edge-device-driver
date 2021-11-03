package models

type Device struct {
	ID           string
	Name         string
	Desc         string
	ProductId    string            `json:"product_id"`    // 设备所属产品ID, 不可更新
	ProductName  string            `json:"product_name"`  // 设备所属产品名称
	Category     string            `json:"category"`      // 设备类型(多媒体, 时序), 不可更新
	Recording    bool              `json:"recording"`     // 是否正在录制
	DeviceStatus string            `json:"device_status"` // 设备状态
	DeviceProps  map[string]string `json:"device_props"`  // 设备动态属性, 取决于具体的设备协议
	DeviceLabels map[string]string `json:"device_labels"` // 设备标签
	DeviceMeta   map[string]string `json:"device_meta"`   // 视频流元信息
}

func (s *Device) GetProp(k string) string {
	return s.DeviceProps[k]
}
