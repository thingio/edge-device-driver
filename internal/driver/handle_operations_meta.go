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
		return true
	})
}

// activateDevice is responsible for establishing the connection with the real device.
func (d *DeviceDriver) activateDevice(device *models.Device) error {
	if _, err := d.getRunner(device.ID); err == nil { // the device has been activated
		_ = d.deactivateDevice(device.ID)
	}

	// build, initialize and start twin
	runner, err := NewTwinRunner(d, device)
	if err != nil {
		return err
	}
	if err := runner.Initialize(d.ctx); err != nil {
		d.logger.WithError(err).Errorf("fail to initialize the device twin[%s]", device.ID)
		return err
	}

	go func() {
		d.putDeviceAndRunner(device.ID, device, runner)
		if err := runner.Start(); err != nil {
			d.logger.WithError(err).Errorf("fail to start the device twin[%s]", device.ID)
			if !d.cfg.DriverOptions.DeviceAutoReconnect {
				d.deleteDeviceAndRunner(device.ID)
			}
			return
		}
		d.logger.Infof("success to activate the device[%s]", device.ID)
	}()
	return nil
}

// deactivateDevices tries to deactivate all devices.
func (d *DeviceDriver) deactivateDevices() {
	d.runners.Range(func(key, value interface{}) bool {
		deviceID := key.(string)
		if err := d.deactivateDevice(deviceID); err != nil {
			d.logger.WithError(err).Errorf("fail to deactivate the device[%s]", deviceID)
		}
		return true
	})
}

// deactivateDevice is responsible for breaking up the connection with the real device.
func (d *DeviceDriver) deactivateDevice(deviceID string) error {
	if runner, _ := d.getRunner(deviceID); runner != nil {
		if err := runner.Stop(false); err != nil {
			return err
		}
	}

	d.deleteDeviceAndRunner(deviceID)
	d.logger.Infof("success to deactivate the device[%s]", deviceID)
	return nil
}

func (d *DeviceDriver) reportingDriverHealth() {
	hello := true
	reportDriverHealth := func() {
		status := &models.DriverStatus{
			Hello:                     hello,
			Protocol:                  d.protocol,
			State:                     models.DriverStateRunning,
			HealthCheckIntervalSecond: d.cfg.DriverOptions.DriverHealthCheckIntervalSecond,
		}
		if err := d.dc.PublishDriverStatus(status); err != nil {
			d.logger.WithError(err).Errorf("fail to publish the status of the driver")
		} else {
			d.logger.Debugf("success to publish the status of the driver: %+v", status)
		}
		hello = false
	}

	interval := time.Duration(d.cfg.DriverOptions.DriverHealthCheckIntervalSecond) * time.Second
	ticker := time.NewTicker(interval)

	reportDriverHealth()
	for {
		select {
		case <-ticker.C:
			reportDriverHealth()
		case <-d.ctx.Done():
			return
		}
	}
}

func (d *DeviceDriver) subscribeMetaMutation() error {
	if err := d.ds.InitializeDriverHandler(d.protocol.ID, d.initializeDriver); err != nil {
		return err
	}

	if err := d.ds.MutateProductHandler(d.protocol.ID, d.updateProduct, d.removeProduct); err != nil {
		return err
	}
	if err := d.ds.MutateDeviceHandler(d.protocol.ID, d.updateDevice, d.removeDevice); err != nil {
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
		return true
	})

	d.deleteProduct(productID)
	return nil
}

func (d *DeviceDriver) updateDevice(device *models.Device) error {
	if err := d.activateDevice(device); err != nil {
		return err
	}
	return nil
}

func (d *DeviceDriver) removeDevice(deviceID string) error {
	if err := d.deactivateDevice(deviceID); err != nil {
		return err
	}
	return nil
}
