package driver

import (
	"context"
	"github.com/patrickmn/go-cache"
	"github.com/thingio/edge-device-std/errors"
	"github.com/thingio/edge-device-std/models"
	"github.com/thingio/edge-device-std/operations"
	"sync"
	"time"
)

const (
	PropertyCacheExpiration      = 30 * time.Second
	PropertyCacheCleanupInterval = 2 * PropertyCacheExpiration
)

func NewTwinRunner(driver *DeviceDriver, device *models.Device) (TwinRunner, error) {
	if driver == nil {
		return nil, errors.DeviceTwin.Error("the driver cannot be nil")
	}
	if device == nil {
		return nil, errors.DeviceTwin.Error("the device cannot be nil")
	}

	return &twinRunner{
		driver: driver,
		device: device,
	}, nil
}

type TwinRunner interface {
	Initialize(ctx context.Context) error
	Start() error
	Stop(force bool) error
	HealthCheck() (*models.DeviceStatus, error)

	// Read indicates soft read, it will read the specified property from the cache with TTL.
	// Specially, when propertyID is "*", it indicates read all properties.
	Read(propertyID models.ProductPropertyID) (map[models.ProductPropertyID]*models.DeviceData, error)
	// HardRead indicates head read, it will read the specified property from the real device.
	// Specially, when propertyID is "*", it indicates read all properties.
	HardRead(propertyID models.ProductPropertyID) (map[models.ProductPropertyID]*models.DeviceData, error)
	Write(propertyID models.ProductPropertyID, values map[models.ProductPropertyID]*models.DeviceData) error
	Call(methodID models.ProductMethodID, ins map[models.ProductPropertyID]*models.DeviceData) (outs map[models.ProductPropertyID]*models.DeviceData, err error)
}

type twinRunner struct {
	driver *DeviceDriver

	product *models.Product
	device  *models.Device
	twin    models.DeviceTwin

	properties     map[models.ProductPropertyID]*models.ProductProperty // for property's reading and writing
	watchScheduler map[time.Duration][]*models.ProductProperty          // for property's watching
	propertyCache  *cache.Cache                                         // for property's soft reading
	methods        map[models.ProductMethodID]*models.ProductMethod     // for method's calling

	once   sync.Once
	lock   sync.Mutex
	parent context.Context
	ctx    context.Context
	cancel context.CancelFunc
}

func (r *twinRunner) Initialize(ctx context.Context) error {
	r.parent = ctx
	if product, err := r.driver.getProduct(r.device.ProductID); err != nil {
		return err
	} else if twin, err := r.driver.twinBuilder(product, r.device); err != nil {
		return err
	} else {
		r.product = product
		r.twin = twin
	}
	if err := r.initProperties(); err != nil {
		return err
	}
	if err := r.initMethods(); err != nil {
		return err
	}
	return r.twin.Initialize(r.driver.logger)
}

func (r *twinRunner) Start() error {
	cfg := r.driver.cfg.DriverOptions
	if cfg.DeviceAutoReconnect {
		go r.once.Do(r.autoReconnect)
	}

	return r.start()
}
func (r *twinRunner) autoReconnect() {
	cfg := r.driver.cfg.DriverOptions
	interval := time.Duration(cfg.DeviceAutoReconnectIntervalSecond) * time.Second

	ticker := time.NewTicker(interval)
	defer func() {
		ticker.Stop()
	}()
	for {
		select {
		case <-ticker.C:
			status, _ := r.twin.HealthCheck()
			switch status.State {
			case models.DeviceStateConnected, models.DeviceStateReconnecting:
				continue
			case models.DeviceStateDisconnected:
				_ = r.Stop(false)
				return
			case models.DeviceStateException:
				if err := r.start(); err != nil {
					r.driver.logger.Errorf("fail to start the twin runner for the device[%s]", r.device.ID)
					continue
				}
			}
		case <-r.ctx.Done():
			return
		}
	}
}
func (r *twinRunner) start() error {
	if r.cancel != nil {
		r.cancel()
	}
	r.ctx, r.cancel = context.WithCancel(r.parent)
	if err := r.twin.Start(r.ctx); err != nil {
		r.device.DeviceStatus = models.DeviceStateException
		_ = r.driver.dc.PublishDeviceStatus(r.driver.protocol.ID, r.product.ID, r.device.ID, &models.DeviceStatus{
			Device:      r.device,
			State:       r.device.DeviceStatus,
			StateDetail: err.Error(),
		})
		return err
	}

	r.device.DeviceStatus = models.DeviceStateConnected
	_ = r.driver.dc.PublishDeviceStatus(r.driver.protocol.ID, r.product.ID, r.device.ID, &models.DeviceStatus{
		Device: r.device,
		State:  r.device.DeviceStatus,
	})
	if err := r.watch(); err != nil {
		return err
	}
	if err := r.subscribe(); err != nil {
		return err
	}
	return nil

}
func (r *twinRunner) Stop(force bool) error {
	defer func() {
		r.cancel()
	}()
	return r.twin.Stop(force)
}
func (r *twinRunner) HealthCheck() (*models.DeviceStatus, error) {
	return r.twin.HealthCheck()
}
func (r *twinRunner) Read(propertyID models.ProductPropertyID) (map[models.ProductPropertyID]*models.DeviceData, error) {
	values := make(map[models.ProductPropertyID]*models.DeviceData)
	if propertyID == models.DeviceDataMultiPropsID {
		for _, property := range r.properties {
			value, ok := r.propertyCache.Get(property.Id)
			if !ok {
				return nil, errors.NotFound.Error("the property[%s] hasn't been ready", property.Id)
			}
			values[property.Id] = value.(*models.DeviceData)
		}
	} else { // single property
		if _, ok := r.properties[propertyID]; !ok {
			return nil, errors.BadRequest.Error("undefined property: %s", propertyID)
		}

		if value, ok := r.propertyCache.Get(propertyID); !ok {
			return nil, errors.NotFound.Error("the property[%s] hasn't been ready", propertyID)
		} else {
			values[propertyID] = value.(*models.DeviceData)
		}
	}
	r.driver.logger.Debugf("success to softly read the property[%s] of the device[%s], returns %+v",
		propertyID, r.device.ID, values)
	return values, nil
}
func (r *twinRunner) HardRead(propertyID models.ProductPropertyID) (map[models.ProductPropertyID]*models.DeviceData, error) {
	values, err := r.twin.Read(propertyID)
	if err != nil {
		return nil, err
	}
	for key, value := range values {
		r.propertyCache.SetDefault(key, value)
	}
	r.driver.logger.Debugf("success to hardly read the property[%s] of the device[%s], returns %+v",
		propertyID, r.device.ID, values)
	return values, nil
}
func (r *twinRunner) Write(propertyID models.ProductPropertyID, values map[models.ProductPropertyID]*models.DeviceData) error {
	for _, value := range values {
		propertyID = value.Name
		property, ok := r.properties[propertyID]
		if !ok {
			return errors.NotFound.Error("undefined property: %s", propertyID)
		}
		if !property.Writeable {
			return errors.DeviceTwin.Error("the property[%s] is read-only", propertyID)
		}
	}
	if err := r.twin.Write(propertyID, values); err != nil {
		return err
	}

	r.driver.logger.Debugf("success to write the property[%s] of the device[%s] with values %+v",
		propertyID, r.device.ID, values)
	return nil
}
func (r *twinRunner) Call(methodID models.ProductMethodID, ins map[models.ProductPropertyID]*models.DeviceData) (
	outs map[models.ProductPropertyID]*models.DeviceData, err error) {
	method, ok := r.methods[methodID]
	if !ok {
		return nil, errors.NotFound.Error("undefined method: %s", methodID)
	}
	for _, in := range method.Ins {
		if _, ok := ins[in.Id]; !ok {
			return nil, errors.BadRequest.Error("missing method input: %+v", in)
		}
	}
	outs, err = r.twin.Call(methodID, ins)
	for _, out := range method.Outs {
		if _, ok := outs[out.Id]; !ok {
			return nil, errors.BadRequest.Error("missing method output: %+v", out)
		}
	}

	r.driver.logger.Debugf("success to call the method[%s] of the device[%s], input %+v, output %+v",
		methodID, r.device.ID, ins, outs)
	return outs, nil
}

func (r *twinRunner) initProperties() error {
	r.properties = make(map[models.ProductPropertyID]*models.ProductProperty)
	for _, property := range r.product.Properties {
		r.properties[property.Id] = property
	}
	r.propertyCache = cache.New(PropertyCacheExpiration, PropertyCacheCleanupInterval)

	r.watchScheduler = make(map[time.Duration][]*models.ProductProperty)
	for _, property := range r.properties {
		if property.ReportMode != operations.DeviceDataReportModePeriodical {
			continue
		}
		duration, err := time.ParseDuration(property.Interval)
		if err != nil {
			return errors.DeviceTwin.Cause(err, "fail to parse the reporting interval: %s", property.Interval)
		} else if duration <= 0 {
			continue
		}

		_, ok := r.watchScheduler[duration]
		if !ok {
			r.watchScheduler[duration] = make([]*models.ProductProperty, 0)
		}
		r.watchScheduler[duration] = append(r.watchScheduler[duration], property)
	}

	return nil
}
func (r *twinRunner) initMethods() error {
	r.methods = make(map[models.ProductMethodID]*models.ProductMethod)
	for _, method := range r.product.Methods {
		r.methods[method.Id] = method
	}

	return nil
}

func (r *twinRunner) watch() error {
	multiRead := func(properties []*models.ProductProperty) map[models.ProductPropertyID]*models.DeviceData {
		result := map[models.ProductPropertyID]*models.DeviceData{}
		for _, property := range properties {
			pairs, err := r.HardRead(property.Id)
			if err != nil {
				r.driver.logger.WithError(err).Errorf("watch properiodly properties[%s]", property.Id)
				continue
			}
			for key, value := range pairs {
				result[key] = value
			}
		}
		return result
	}

	for duration, properties := range r.watchScheduler {
		go func(d time.Duration, pps []*models.ProductProperty) {
			ticker := time.NewTicker(d)
			defer func() {
				ticker.Stop()
			}()
			for {
				select {
				case <-ticker.C:
					propertyID := models.DeviceDataMultiPropsID
					if len(pps) == 1 {
						propertyID = pps[0].Id
					}
					r.driver.propsBus <- &models.DeviceDataWrapper{
						ProductID:  r.product.ID,
						DeviceID:   r.device.ID,
						FuncID:     propertyID,
						Properties: multiRead(pps),
					}
				case <-r.ctx.Done():
					return
				}
			}
		}(duration, properties)
	}
	r.driver.logger.Debugf("success to watch the device[%s]", r.device.ID)
	return nil
}
func (r *twinRunner) subscribe() error {
	for _, event := range r.product.Events {
		if err := r.twin.Subscribe(event.Id, r.driver.eventBus); err != nil {
			return errors.DeviceTwin.Cause(err, "fail to subscribe the event: %s", event.Id)
		}
		r.driver.logger.Debugf("success to subscribe the event[%s]", r.device.ID)
	}

	return nil
}
