package driver

import (
	"github.com/thingio/edge-device-std/models"
	"time"
)

// activateDevices tries to activate all devices.
func (d *DeviceDriver) activateDevices() {
	d.devices.Range(func(key, value interface{}) bool {
		device := value.(*models.Device)

		if err := d.activateDevice(device); err != nil {
			d.logger.WithError(err).Errorf("fail to activate the device[%s]", device.ID)
			return true
		}
		d.putDevice(device)

		return true
	})
}

// activateDevice is responsible for establishing the connection with the real device.
func (d *DeviceDriver) activateDevice(device *models.Device) error {
	if twin, _ := d.getDeviceTwin(device.ID); twin != nil { // the device has been activated
		_ = d.deactivateDevice(device.ID)
	}

	// build, initialize and start twin
	product, err := d.getProduct(device.ProductID)
	if err != nil {
		return err
	}
	twin, err := d.twinBuilder(product, device)
	if err != nil {
		return err
	}
	if err = twin.Initialize(d.logger); err != nil {
		d.logger.WithError(err).Error("fail to initialize the random device twin")
		return err
	}
	if err = twin.Start(); err != nil {
		d.logger.WithError(err).Error("fail to start the random device twin")
		return err
	}

	if len(product.Properties) != 0 {
		if err = twin.Watch(d.propsBus); err != nil {
			d.logger.WithError(err).Error("fail to watch properties for the device")
			return err
		}
	}
	for _, event := range product.Events {
		if err = twin.Subscribe(event.Id, d.eventBus); err != nil {
			d.logger.WithError(err).Errorf("fail to subscribe the product event[%s]", event.Id)
			continue
		}
	}

	d.putDeviceTwin(device.ID, twin)
	d.logger.Infof("success to activate the device[%s]", device.ID)
	return nil
}

// deactivateDevices tries to deactivate all devices.
func (d *DeviceDriver) deactivateDevices() {
	d.deviceTwins.Range(func(key, value interface{}) bool {
		deviceID := key.(string)
		if err := d.deactivateDevice(deviceID); err != nil {
			d.logger.WithError(err).Errorf("fail to deactivate the device[%s]", deviceID)
		}
		return true
	})
}

// deactivateDevice is responsible for breaking up the connection with the real device.
func (d *DeviceDriver) deactivateDevice(deviceID string) error {
	twin, _ := d.getDeviceTwin(deviceID)
	if twin == nil {
		return nil
	}
	if err := twin.Stop(false); err != nil {
		return err
	}

	d.deleteDeviceTwin(deviceID)
	d.logger.Infof("success to deactivate the device[%s]", deviceID)
	return nil
}

func (d *DeviceDriver) reportingDriverHealth() {
	hello := true
	ticker := time.NewTicker(time.Duration(d.cfg.CommonOptions.DriverHealthCheckIntervalSecond) * time.Second)
	for {
		select {
		case <-ticker.C:
			status := &models.DriverStatus{
				Hello:    hello,
				Protocol: d.protocol,
				State:    models.DriverStateRunning,
			}
			if err := d.dc.PublishDriverStatus(status); err != nil {
				d.logger.WithError(err).Errorf("fail to publish the status of the driver")
			} else {
				d.logger.Debugf("success to publish the status of the driver: %+v", status)
			}
			hello = false
		case <-d.ctx.Done():
			ticker.Stop()
			return
		}
	}
}

func (d *DeviceDriver) subscribeMetaMutation() error {
	if err := d.ds.InitializeDriverHandler(d.protocol.ID, d.initializeDriver); err != nil {
		return err
	}

	if err := d.ds.UpdateProductHandler(d.protocol.ID, d.updateProduct); err != nil {
		return err
	}
	if err := d.ds.DeleteProductHandler(d.protocol.ID, d.removeProduct); err != nil {
		return err
	}

	if err := d.ds.UpdateDeviceHandler(d.protocol.ID, d.updateDevice); err != nil {
		return err
	}
	if err := d.ds.DeleteDeviceHandler(d.protocol.ID, d.removeDevice); err != nil {
		return err
	}
	return nil
}

func (d *DeviceDriver) initializeDriver(products []*models.Product, devices []*models.Device) error {
	for _, product := range products {
		d.putProduct(product)
	}

	for _, device := range devices {
		if err := d.updateDevice(device); err != nil {
			return err
		}
	}
	return nil
}

func (d *DeviceDriver) updateProduct(product *models.Product) error {
	d.putProduct(product)

	d.devices.Range(func(key, value interface{}) bool {
		device := value.(*models.Device)
		if device.ProductID != product.ID {
			return true
		}
		if err := d.activateDevice(device); err != nil {
			d.logger.WithError(err).Errorf("fail to reactivate the device[%s] after updating the product[%s]",
				device.ID, product.ID)
			return true
		}
		d.putDevice(device)
		return true
	})
	return nil
}

func (d *DeviceDriver) removeProduct(productID string) error {
	d.devices.Range(func(key, value interface{}) bool {
		device := value.(*models.Device)
		if device.ProductID != productID {
			return true
		}
		if err := d.deactivateDevice(device.ID); err != nil {
			d.logger.WithError(err).Errorf("fail to deactivate the device[%s] after updating the product[%s]",
				device.ID, productID)
			return true
		}
		d.deleteDevice(device.ID)
		return true
	})

	d.deleteProduct(productID)
	return nil
}

func (d *DeviceDriver) updateDevice(device *models.Device) error {
	if err := d.activateDevice(device); err != nil {
		return err
	}
	d.putDevice(device)
	return nil
}

func (d *DeviceDriver) removeDevice(deviceID string) error {
	if err := d.deactivateDevice(deviceID); err != nil {
		return err
	}
	d.deleteDevice(deviceID)
	return nil
}
