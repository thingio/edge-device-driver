package operations

import (
	bus "github.com/thingio/edge-device-sdk/internal/message_bus"
	"github.com/thingio/edge-device-sdk/logger"
	"github.com/thingio/edge-device-sdk/pkg/models"
)

func NewDeviceManagerMetaOperationClient(mb bus.MessageBus,
	logger *logger.Logger) (DeviceManagerMetaOperationClient, error) {
	protocolClient, err := newDeviceManagerProtocolOperationClient(mb, logger)
	if err != nil {
		return nil, err
	}
	productClient, err := newDeviceManagerProductOperationClient(mb, logger)
	if err != nil {
		return nil, err
	}
	deviceClient, err := newDeviceManagerDeviceOperationClient(mb, logger)
	if err != nil {
		return nil, err
	}
	return &deviceManagerMetaOperationClient{
		protocolClient,
		productClient,
		deviceClient,
	}, nil
}

type DeviceManagerMetaOperationClient interface {
	DeviceManagerProtocolOperationClient
	DeviceManagerProductOperationClient
	DeviceManagerDeviceOperationClient
}
type deviceManagerMetaOperationClient struct {
	DeviceManagerProtocolOperationClient
	DeviceManagerProductOperationClient
	DeviceManagerDeviceOperationClient
}

func newDeviceManagerProtocolOperationClient(mb bus.MessageBus,
	logger *logger.Logger) (DeviceManagerProtocolOperationClient, error) {
	return &deviceManagerProtocolOperationClient{mb: mb, logger: logger}, nil
}

type DeviceManagerProtocolOperationClient interface {
	OnRegisterProtocols(register func(protocol *models.Protocol) error) error
	OnUnregisterProtocols(unregister func(protocolID string) error) error
}
type deviceManagerProtocolOperationClient struct {
	mb     bus.MessageBus
	logger *logger.Logger
}

func newDeviceManagerProductOperationClient(mb bus.MessageBus,
	logger *logger.Logger) (DeviceManagerProductOperationClient, error) {
	return &deviceManagerProductOperationClient{mb: mb, logger: logger}, nil
}

type DeviceManagerProductOperationClient interface {
	OnListProducts(list func(protocolID string) ([]*models.Product, error)) error
}
type deviceManagerProductOperationClient struct {
	mb     bus.MessageBus
	logger *logger.Logger
}

func newDeviceManagerDeviceOperationClient(mb bus.MessageBus,
	logger *logger.Logger) (DeviceManagerDeviceOperationClient, error) {
	return &deviceManagerDeviceOperationClient{mb: mb, logger: logger}, nil
}

type DeviceManagerDeviceOperationClient interface {
	OnListDevices(list func(productID string) ([]*models.Device, error)) error
}
type deviceManagerDeviceOperationClient struct {
	mb     bus.MessageBus
	logger *logger.Logger
}

func NewDeviceManagerDeviceDataOperationClient(mb bus.MessageBus, logger *logger.Logger) (DeviceManagerDeviceDataOperationClient, error) {
	reads := make(map[models.ProductPropertyID]chan models.DeviceData)
	events := make(map[models.ProductEventID]chan models.DeviceData)
	return &deviceManagerDeviceDataOperationClient{
		mb:     mb,
		logger: logger,
		reads:  reads,
		events: events,
	}, nil
}

type DeviceManagerDeviceDataOperationClient interface {
	Read(productID, deviceID string,
		propertyID models.ProductPropertyID) (dd <-chan models.DeviceData, cc func(), err error)

	Write(productID, deviceID string,
		propertyID models.ProductPropertyID, value interface{}) error

	Receive(productID, deviceID string,
		eventID models.ProductEventID) (dd <-chan models.DeviceData, cc func(), err error)

	Call(productID, deviceID string, methodID models.ProductMethodID,
		req map[string]interface{}) (rsp map[string]interface{}, err error)
}

type deviceManagerDeviceDataOperationClient struct {
	mb     bus.MessageBus
	logger *logger.Logger

	reads  map[string]chan models.DeviceData
	events map[string]chan models.DeviceData
}
