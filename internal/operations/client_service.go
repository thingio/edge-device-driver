package operations

import (
	bus "github.com/thingio/edge-device-sdk/internal/message_bus"
	"github.com/thingio/edge-device-sdk/logger"
	"github.com/thingio/edge-device-sdk/pkg/models"
)

func NewDeviceServiceMetaOperationClient(mb bus.MessageBus,
	logger *logger.Logger) (DeviceServiceMetaOperationClient, error) {
	protocolClient, err := newDeviceServiceProtocolOperationClient(mb, logger)
	if err != nil {
		return nil, err
	}
	productClient, err := newDeviceServiceProductOperationClient(mb, logger)
	if err != nil {
		return nil, err
	}
	deviceClient, err := newDeviceServiceDeviceOperationClient(mb, logger)
	if err != nil {
		return nil, err
	}
	return &deviceServiceMetaOperationClient{
		protocolClient,
		productClient,
		deviceClient,
	}, nil
}

type DeviceServiceMetaOperationClient interface {
	DeviceServiceProtocolOperationClient
	DeviceServiceProductOperationClient
	DeviceServiceDeviceOperationClient
}
type deviceServiceMetaOperationClient struct {
	DeviceServiceProtocolOperationClient
	DeviceServiceProductOperationClient
	DeviceServiceDeviceOperationClient
}

func newDeviceServiceProtocolOperationClient(mb bus.MessageBus,
	logger *logger.Logger) (DeviceServiceProtocolOperationClient, error) {
	return &deviceServiceProtocolOperationClient{mb: mb, logger: logger}, nil
}

type DeviceServiceProtocolOperationClient interface {
	RegisterProtocol(protocol *models.Protocol) error
}
type deviceServiceProtocolOperationClient struct {
	mb     bus.MessageBus
	logger *logger.Logger
}

func newDeviceServiceProductOperationClient(mb bus.MessageBus,
	logger *logger.Logger) (DeviceServiceProductOperationClient, error) {
	return &deviceServiceProductOperationClient{mb: mb, logger: logger}, nil
}

type DeviceServiceProductOperationClient interface {
	ListProducts(protocolID string) error
}
type deviceServiceProductOperationClient struct {
	mb     bus.MessageBus
	logger *logger.Logger
}

func newDeviceServiceDeviceOperationClient(mb bus.MessageBus,
	logger *logger.Logger) (DeviceServiceDeviceOperationClient, error) {
	return &deviceServiceDeviceOperationClient{mb: mb, logger: logger}, nil
}

type DeviceServiceDeviceOperationClient interface {
	ListDevices(productID string) ([]*models.Device, error)
}
type deviceServiceDeviceOperationClient struct {
	mb     bus.MessageBus
	logger *logger.Logger
}

func NewDeviceServiceDeviceDataOperationClient(mb bus.MessageBus,
	logger *logger.Logger) (DeviceServiceDeviceDataOperationClient, error) {
	return &deviceServiceDeviceDataOperationClient{mb: mb, logger: logger}, nil
}

type DeviceServiceDeviceDataOperationClient interface {
}

type deviceServiceDeviceDataOperationClient struct {
	mb     bus.MessageBus
	logger *logger.Logger
}
