package models

type Protocol struct {
	ID           string             `json:"id"`
	Name         string             `json:"name"`
	Desc         string             `json:"desc"`
	Category     string             `json:"category"`
	Language     string             `json:"language"`
	SupportFuncs []string           `json:"support_funcs"`
	AuxProps     []*GeneralProperty `json:"aux_props"`
	DeviceProps  []*GeneralProperty `json:"device_props"`
}
