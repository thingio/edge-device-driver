package models

type MetaStore interface {
	ListProducts(protocolID string) ([]*Product, error)
	ListDevices(productID string) ([]*Device, error)
}
