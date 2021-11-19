package startup

import (
	"context"
	"fmt"
	"github.com/thingio/edge-device-sdk/config"
	bus "github.com/thingio/edge-device-sdk/internal/message_bus"
	"github.com/thingio/edge-device-sdk/internal/operations"
	"github.com/thingio/edge-device-sdk/logger"
	"github.com/thingio/edge-device-sdk/pkg/models"
	"github.com/thingio/edge-device-sdk/version"
	"os"
	"sync"
)

type DeviceService struct {
	// service information
	ID      string
	Name    string
	Version string

	// driver
	protocol         *models.Protocol
	connectorBuilder models.ConnectorBuilder

	// caches
	products                   sync.Map
	devices                    sync.Map
	deviceConnectors           sync.Map
	unsupportedDeviceDataTypes map[models.DeviceDataOperation]struct{}
	deviceDataHandlers         map[models.DeviceDataOperation]DeviceDataHandler

	// operation clients
	bus chan models.DeviceData
	moc operations.DeviceServiceMetaOperationClient       // wrap the message bus to manipulate
	doc operations.DeviceServiceDeviceDataOperationClient // warp the message bus to manipulate device data

	// lifetime control variables for the device service
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	logger *logger.Logger
}

func (s *DeviceService) Initialize(ctx context.Context, cancel context.CancelFunc,
	protocol *models.Protocol, connectorBuilder models.ConnectorBuilder) {
	s.logger = logger.NewLogger()

	s.ID = protocol.ID
	s.Name = protocol.Name
	s.Version = version.Version

	s.protocol = protocol
	if connectorBuilder == nil {
		s.logger.Error("please implement and specify the connector builder")
		os.Exit(1)
	}
	s.connectorBuilder = connectorBuilder

	s.products = sync.Map{}
	s.devices = sync.Map{}
	s.deviceConnectors = sync.Map{}

	s.initializeOperationClients()

	s.ctx = ctx
	s.cancel = cancel
	s.wg = sync.WaitGroup{}
}

func (s *DeviceService) initializeOperationClients() {
	s.bus = make(chan models.DeviceData, 100)

	mb, err := bus.NewMessageBus(&config.C.MessageBus, s.logger)
	if err != nil {
		s.logger.WithError(err).Error("fail to initialize the message bus")
		os.Exit(1)
	}
	if err = mb.Connect(); err != nil {
		s.logger.WithError(err).Error("fail to connect to the message bus")
		os.Exit(1)
	}

	moc, err := operations.NewDeviceServiceMetaOperationClient(mb, s.logger)
	if err != nil {
		s.logger.WithError(err).Error("fail to initialize the meta operation client for the device service")
		os.Exit(1)
	}
	s.moc = moc
	doc, err := operations.NewDeviceServiceDeviceDataOperationClient(mb, s.logger)
	if err != nil {
		s.logger.WithError(err).Error("fail to initialize the device data operation client for the device service")
		os.Exit(1)
	}
	s.doc = doc
}

func (s *DeviceService) Serve() {
	defer s.Stop(false)

	s.registerProtocol()
	defer s.unregisterProtocol()

	s.activateDevices()
	defer s.deactivateDevices()

	s.publishingDeviceData()
	s.handlingDeviceData()

	s.wg.Wait()
}

func (s *DeviceService) Stop(force bool) {}

func (s *DeviceService) getProduct(productID string) (*models.Product, error) {
	v, ok := s.products.Load(productID)
	if ok {
		return v.(*models.Product), nil
	}
	return nil, fmt.Errorf("the product[%s] is not found in cache", productID)
}

func (s *DeviceService) getDevice(deviceID string) (*models.Device, error) {
	v, ok := s.devices.Load(deviceID)
	if ok {
		return v.(*models.Device), nil
	}
	return nil, fmt.Errorf("the device[%s] is not found in cache", deviceID)
}

func (s *DeviceService) getDeviceConnector(deviceID string) (models.DeviceConnector, error) {
	v, ok := s.deviceConnectors.Load(deviceID)
	if ok {
		return v.(models.DeviceConnector), nil
	}
	return nil, fmt.Errorf("the device[%s] is not activated", deviceID)
}
