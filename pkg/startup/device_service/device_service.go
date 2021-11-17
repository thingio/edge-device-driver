package startup

import (
	"context"
	"github.com/thingio/edge-device-sdk/config"
	bus "github.com/thingio/edge-device-sdk/internal/message_bus"
	"github.com/thingio/edge-device-sdk/internal/operations"
	"github.com/thingio/edge-device-sdk/logger"
	"github.com/thingio/edge-device-sdk/pkg/models"
	"github.com/thingio/edge-device-sdk/pkg/version"
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
	products         sync.Map
	devices          sync.Map
	deviceConnectors sync.Map

	// operation clients
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
	// TODO Read from the configuration file
	options := &config.MessageBusOptions{
		Host:                     "172.16.251.163",
		Port:                     1883,
		Protocol:                 "tcp",
		ConnectTimoutMillisecond: 30000,
		TimeoutMillisecond:       1000,
		QoS:                      0,
		CleanSession:             false,
	}
	mb, err := bus.NewMessageBus(options, s.logger)
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

	s.wg.Wait()
}

func (s *DeviceService) Stop(force bool) {}
