package startup

import (
	"context"
	"github.com/thingio/edge-device-sdk/config"
	bus "github.com/thingio/edge-device-sdk/internal/message_bus"
	"github.com/thingio/edge-device-sdk/internal/operations"
	"github.com/thingio/edge-device-sdk/logger"
	"github.com/thingio/edge-device-sdk/pkg/models"
	"github.com/thingio/edge-device-sdk/version"
	"os"
	"sync"
)

type DeviceManager struct {
	// manager information
	Version string

	// caches
	protocols sync.Map

	// operation clients
	moc       operations.DeviceManagerMetaOperationClient       // wrap the message bus to manipulate meta
	doc       operations.DeviceManagerDeviceDataOperationClient // warp the message bus to manipulate device data
	metaStore models.MetaStore                                  // meta store

	// lifetime control variables for the device service
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	logger *logger.Logger
}

func (m *DeviceManager) Initialize(ctx context.Context, cancel context.CancelFunc, metaStore models.MetaStore) {
	m.logger = logger.NewLogger()

	m.Version = version.Version

	m.protocols = sync.Map{}

	m.initializeOperationClients(metaStore)

	m.ctx = ctx
	m.cancel = cancel
	m.wg = sync.WaitGroup{}
}

func (m *DeviceManager) initializeOperationClients(metaStore models.MetaStore) {
	mb, err := bus.NewMessageBus(&config.C.MessageBus, m.logger)
	if err != nil {
		m.logger.WithError(err).Error("fail to initialize the message bus")
		os.Exit(1)
	}
	if err = mb.Connect(); err != nil {
		m.logger.WithError(err).Error("fail to connect to the message bus")
		os.Exit(1)
	}

	moc, err := operations.NewDeviceManagerMetaOperationClient(mb, m.logger)
	if err != nil {
		m.logger.WithError(err).Error("fail to initialize the meta operation client for the device service")
		os.Exit(1)
	}
	m.moc = moc
	doc, err := operations.NewDeviceManagerDeviceDataOperationClient(mb, m.logger)
	if err != nil {
		m.logger.WithError(err).Error("fail to initialize the device data operation client for the device service")
		os.Exit(1)
	}
	m.doc = doc

	m.metaStore = metaStore
}

func (m *DeviceManager) Serve() {
	defer m.Stop(false)

	m.wg.Add(1)
	go m.watchingProtocols()
	m.wg.Add(1)
	go m.watchingProductOperations()
	m.wg.Add(1)
	go m.watchingDeviceOperations()

	m.wg.Wait()
}

func (m *DeviceManager) Stop(force bool) {}
