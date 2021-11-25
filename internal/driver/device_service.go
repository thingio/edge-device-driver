package driver

import (
	"context"
	"errors"
	"fmt"
	"github.com/thingio/edge-device-std/config"
	"github.com/thingio/edge-device-std/logger"
	"github.com/thingio/edge-device-std/models"
	bus "github.com/thingio/edge-device-std/msgbus"
	"github.com/thingio/edge-device-std/operations"
	"github.com/thingio/edge-device-std/version"
	"os"
	"sync"
)

func NewDeviceDriver(ctx context.Context, cancel context.CancelFunc,
	protocol *models.Protocol, dtBuilder models.DeviceTwinBuilder) (*DeviceDriver, error) {
	if protocol == nil {
		return nil, errors.New("the product cannot be nil")
	}
	if dtBuilder == nil {
		return nil, errors.New("please implement and specify the connector builder")
	}
	dd := &DeviceDriver{
		ID:      protocol.ID,
		Name:    protocol.Name,
		Version: version.Version,

		protocol:  protocol,
		dtBuilder: dtBuilder,

		products:         sync.Map{},
		devices:          sync.Map{},
		deviceConnectors: sync.Map{},

		ctx:    ctx,
		cancel: cancel,
		wg:     sync.WaitGroup{},
		logger: logger.NewLogger(),
	}

	dd.unsupportedDataOperations = map[models.DataOperationType]struct{}{
		models.DataOperationTypeHealthCheckPong: {}, // device health check pong
		models.DataOperationTypeSoftReadRsp:     {}, // property soft read response
		models.DataOperationTypeHardReadRsp:     {}, // property hard read response
		models.DataOperationTypeWriteRsp:        {}, // property write response
		models.DataOperationTypeWatch:           {}, // property watch
		models.DataOperationTypeEvent:           {}, // event
		models.DataOperationTypeResponse:        {}, // method response
		models.DataOperationTypeError:           {}, // method error
	}
	dd.dataOperationHandlers = map[models.DataOperationType]DeviceOperationHandler{
		models.DataOperationTypeHealthCheckPing: dd.handleHealthCheck, // device health check ping
		models.DataOperationTypeSoftReadReq:     dd.handleSoftRead,    // property soft read
		models.DataOperationTypeHardReadReq:     dd.handleHardRead,    // property hard read
		models.DataOperationTypeWriteReq:        dd.handleWrite,       // property write
		models.DataOperationTypeRequest:         dd.handleRequest,     // method request
	}
	return dd, nil
}

type DeviceDriver struct {
	// driver information
	ID      string
	Name    string
	Version string

	// driver
	protocol  *models.Protocol
	dtBuilder models.DeviceTwinBuilder

	// caches
	products                  sync.Map
	devices                   sync.Map
	deviceConnectors          sync.Map
	unsupportedDataOperations map[models.DataOperationType]struct{}
	dataOperationHandlers     map[models.DataOperationType]DeviceOperationHandler

	// operation clients
	bus chan models.DataOperation
	moc operations.MetaOperationDriverClient // wrap the message bus to manipulate
	doc operations.DataOperationDriverClient // warp the message bus to manipulate device data

	// lifetime control variables for the device driver
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	logger *logger.Logger
}

func (d *DeviceDriver) Initialize() {
	d.initializeOperationClients()
}

func (d *DeviceDriver) initializeOperationClients() {
	d.bus = make(chan models.DataOperation, 100)

	mb, err := bus.NewMessageBus(&config.C.MessageBus, d.logger)
	if err != nil {
		d.logger.WithError(err).Error("fail to initialize the message bus")
		os.Exit(1)
	}
	if err = mb.Connect(); err != nil {
		d.logger.WithError(err).Error("fail to connect to the message bus")
		os.Exit(1)
	}

	moc, err := operations.NewMetaOperationDriverClient(mb, d.logger)
	if err != nil {
		d.logger.WithError(err).Error("fail to initialize the meta operation client for the device driver")
		os.Exit(1)
	}
	d.moc = moc
	doc, err := operations.NewDataOperationDriverClient(mb, d.logger)
	if err != nil {
		d.logger.WithError(err).Error("fail to initialize the device data operation client for the device driver")
		os.Exit(1)
	}
	d.doc = doc
}

func (d *DeviceDriver) Serve() {
	defer d.Stop(false)

	d.registerProtocol()
	defer d.unregisterProtocol()

	d.activateDevices()
	defer d.deactivateDevices()

	d.publishingDeviceOperation()
	d.handlingDeviceOperation()

	d.wg.Wait()
}

func (d *DeviceDriver) Stop(force bool) {}

func (d *DeviceDriver) getProduct(productID string) (*models.Product, error) {
	v, ok := d.products.Load(productID)
	if ok {
		return v.(*models.Product), nil
	}
	return nil, fmt.Errorf("the product[%s] is not found in cache", productID)
}

func (d *DeviceDriver) getDevice(deviceID string) (*models.Device, error) {
	v, ok := d.devices.Load(deviceID)
	if ok {
		return v.(*models.Device), nil
	}
	return nil, fmt.Errorf("the device[%s] is not found in cache", deviceID)
}

func (d *DeviceDriver) getDeviceConnector(deviceID string) (models.DeviceTwin, error) {
	v, ok := d.deviceConnectors.Load(deviceID)
	if ok {
		return v.(models.DeviceTwin), nil
	}
	return nil, fmt.Errorf("the device[%s] is not activated", deviceID)
}
