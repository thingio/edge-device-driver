package service

import (
	"fmt"
	"github.com/thingio/edge-device-sdk/driver"
	"github.com/thingio/edge-device-sdk/models"
	"sync"
)

var (
	bus           = make(chan models.DeviceData, 100)
	devicesCache  = sync.Map{}
	productsCache = sync.Map{}
	dvcConnCache  = sync.Map{}
)

// ActivateAllDevices 尝试激活所有设备
func (s *DeviceService) ActivateAllDevices() error {
	// activate every device
	wg := sync.WaitGroup{}
	wg.Add(1)
	devicesCache.Range(func(key, value interface{}) bool {
		device := value.(*models.Device)
		if _, ok := dvcConnCache.Load(device.ID); ok {
			// skip those devices which already has a connection
			return true
		}
		go func(dvc *models.Device) {
			wg.Add(1)
			_, _ = s.ActivateDevice(dvc)
			wg.Done()
			//if err != nil {
			//	stdlog.WithError(err).Errorf("error while activating device %s", dvc.ID)
			//}
		}(device)
		return true
	})
	wg.Wait()
	return nil
}

// DeactivateAllDevices 断开所有已激活设备的连接
func (s *DeviceService) DeactivateAllDevices() {
	dvcConnCache.Range(func(key, value interface{}) bool {
		s.DeactivateDevice(key.(string))
		return true
	})
}

// UpdateDevice 监控设备资源的改动事件, 并且将其及时应用到指定设备
func (s *DeviceService) UpdateDevice(device *models.Device) error {
	// 设备被删除
	s.DeactivateDevice(device.ID)

	// 设备更新
	if _, err := s.ActivateDevice(device); err != nil {
		return fmt.Errorf("%e,failed to reactivate device", err)
	}
	return nil
}

// DeleteProduct 删除产品，并下线对应的所有设备
func (s *DeviceService) DeleteProduct(pid string) error {
	devicesCache.Range(func(key, value interface{}) bool {
		d := value.(*models.Device)
		if d.ProductId != pid {
			return true
		}
		s.DeactivateDevice(d.ID)
		devicesCache.Delete(key)
		return true
	})
	return nil
}

// UpdateProduct 更新产品资源的改动, 并且将其及时应用到其下属所有相关设备
func (s *DeviceService) UpdateProduct(product *models.Product) error {
	var errs []error
	p := product
	productsCache.Store(p.ID, p)
	devicesCache.Range(func(key, value interface{}) bool {
		d := value.(*models.Device)
		if d.ProductId != p.ID {
			return true
		}
		// reactivate device
		if _, err := s.ActivateDevice(d); err != nil {
			errs = append(errs, fmt.Errorf("%e, failed to reactivate device %s", err, d.ID))
		}
		return true
	})
	return fmt.Errorf("%e, for the changing of product %s", errs, product.ID)
}

// publishingDeviceData 方法用于发布设备主动上传的设备数据
// 包括属性数据和事件数据, 通过共用bus管道实现
func (s *DeviceService) publishingDeviceData(bus <-chan models.DeviceData) {
	for {
		data := <-bus
		//stdlog.Debugf("received device data : %+v ", data)
		if err := s.mqttCli.Pub(data.MustToMQTTMsg()); err != nil {
			if err != nil {
				s.Errorf("error while publishing device data : %v", data)
			}
		}
	}
}

// ActivateDevice 激活或重启指定的设备
func (s *DeviceService) ActivateDevice(device *models.Device) (driver.DeviceConn, error) {
	// whether the device driver already exist
	if _, ok := dvcConnCache.Load(device.ID); ok {
		// deactivate device firstly for reactivating it
		s.DeactivateDevice(device.ID)
	}

	product, err := getProduct(device.ProductId)
	if err != nil {
		return nil, err
	}

	conn, err := s.driver(product, device)
	if err != nil {
		return nil, err
	}
	if err = conn.Start(); err != nil {
		return nil, err
	}

	//// watch all properties of this device
	//if product.Protocol == models.TickProtocolOPCUA ||
	//	product.Protocol == models.TickProtocolCSV ||
	//	len(product.Properties) != 0 {
	//	if err = conn.Watch(bus); err != nil {
	//		return nil, err
	//	}
	//}

	// watch all events of this device
	for _, evt := range product.Events {
		if err = conn.Subscribe(evt.Id, bus); err != nil {
			return nil, err
		}
	}

	s.Infof("activate device %s(%s) success", device.Name, device.ID)
	// put device driver into cache
	dvcConnCache.Store(device.ID, conn)
	return conn, nil
}

// DeactivateDevice 停止指定设备的数据驱动(数据采集,事件监听等)
func (s *DeviceService) DeactivateDevice(deviceID string) {
	conn, ok := dvcConnCache.Load(deviceID)
	if !ok {
		s.Warnf("device %s has already been deactivated", deviceID)
		return
	}
	if err := conn.(driver.DeviceConn).Stop(); err != nil {
		s.Errorf("%e,error while deactivating device %s, stopping driver failed", err, deviceID)
	}
	dvcConnCache.Delete(deviceID)
}

// getProduct 获取Product定义, 并对其进行缓存
func getProduct(pid string) (*models.Product, error) {
	// get product from cache
	v, ok := productsCache.Load(pid)
	if ok {
		return v.(*models.Product), nil
	}
	return nil, fmt.Errorf("product '%s' not found", pid)
}

// getDevice 获取Device定义, 并对其进行缓存
func getDevice(did string) (*models.Device, error) {
	// get device from cache
	v, ok := devicesCache.Load(did)
	if ok {
		return v.(*models.Device), nil
	}

	return nil, fmt.Errorf("device '%s' not found", did)
}

// getDvsConn 返回指定设备ID对应的数据访问连接, 未连接时返回 DeviceNotConnectedError
func getDvsConn(deviceID string) (driver.DeviceConn, error) {
	value, ok := dvcConnCache.Load(deviceID)
	if ok {
		return value.(driver.DeviceConn), nil
	}
	if _, err := getDevice(deviceID); err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("'%s' not connected", deviceID)
}
