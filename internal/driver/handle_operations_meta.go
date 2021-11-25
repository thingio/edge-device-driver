package driver

import (
	"github.com/thingio/edge-device-std/config"
	"github.com/thingio/edge-device-std/models"
	"os"
	"time"
)

// registerProtocol tries to register the protocol to the device manager.
func (d *DeviceDriver) registerProtocol() {
	d.wg.Add(1)

	register := func() {
		if err := d.moc.RegisterProtocol(d.protocol); err != nil {
			d.logger.WithError(err).Errorf("fail to register the protocol[%d] "+
				"to the device manager", d.protocol.ID)
			os.Exit(1)
		}
		d.logger.Infof("success to register the protocol[%d] to the device manager", d.protocol.ID)
	}
	register()

	go func() {
		defer d.wg.Done()

		protocolRegisterInterval := time.Duration(config.C.CommonOptions.ProtocolRegisterIntervalSecond) * time.Second
		ticker := time.NewTicker(protocolRegisterInterval)
		for {
			select {
			case <-ticker.C:
				register()
			case <-d.ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

// unregisterProtocol tries to unregister the protocol from the device manager.
func (d *DeviceDriver) unregisterProtocol() {
	if err := d.moc.UnregisterProtocol(d.protocol.ID); err != nil {
		d.logger.WithError(err).Errorf("fail to unregister the protocol[%d] "+
			"from the device manager", d.protocol.ID)
	}
	d.logger.Infof("success to unregister the protocol[%d] from the device manager", d.protocol.ID)
}

// activateDevices tries to activate all devices.
func (d *DeviceDriver) activateDevices() {
	products, err := d.moc.ListProducts(d.protocol.ID)
	if err != nil {
		d.logger.WithError(err).Error("fail to fetch products from the device manager")
		os.Exit(1)
	}
	for _, product := range products {
		devices, err := d.moc.ListDevices(product.ID)
		if err != nil {
			d.logger.WithError(err).Error("fail to fetch devices from the device manager")
			os.Exit(1)
		}

		d.products.Store(product.ID, product)
		for _, device := range devices {
			if err := d.activateDevice(device); err != nil {
				d.logger.WithError(err).Errorf("fail to activate the device[%d]", device.ID)
				continue
			}
			d.devices.Store(device.ID, device)
		}
	}
}

// activateDevice is responsible for establishing the connection with the real device.
func (d *DeviceDriver) activateDevice(device *models.Device) error {
	if connector, _ := d.getDeviceConnector(device.ID); connector != nil { // the device has been activated
		_ = d.deactivateDevice(device.ID)
	}

	// build, initialize and start connector
	product, err := d.getProduct(device.ProductID)
	if err != nil {
		return err
	}
	connector, err := d.dtBuilder(product, device)
	if err != nil {
		return err
	}
	if err := connector.Initialize(d.logger); err != nil {
		d.logger.WithError(err).Error("fail to initialize the random device connector")
		return err
	}
	if err := connector.Start(); err != nil {
		d.logger.WithError(err).Error("fail to start the random device connector")
		return err
	}

	if len(product.Properties) != 0 {
		if err := connector.Watch(d.bus); err != nil {
			d.logger.WithError(err).Error("fail to watch properties for the device")
			return err
		}
	}

	for _, event := range product.Events {
		if err := connector.Subscribe(event.Id, d.bus); err != nil {
			d.logger.WithError(err).Errorf("fail to subscribe the product event[%d]", event.Id)
			continue
		}
	}

	d.deviceConnectors.Store(device.ID, connector)
	d.logger.Infof("success to activate the device[%d]", device.ID)
	return nil
}

// deactivateDevices tries to deactivate all devices.
func (d *DeviceDriver) deactivateDevices() {
	d.deviceConnectors.Range(func(key, value interface{}) bool {
		deviceID := key.(string)
		if err := d.deactivateDevice(deviceID); err != nil {
			d.logger.WithError(err).Errorf("fail to deactivate the device[%d]", deviceID)
		}
		return true
	})
}

// deactivateDevice is responsible for breaking up the connection with the real device.
func (d *DeviceDriver) deactivateDevice(deviceID string) error {
	connector, _ := d.getDeviceConnector(deviceID)
	if connector == nil {
		return nil
	}
	if err := connector.Stop(false); err != nil {
		return err
	}

	d.deviceConnectors.Delete(deviceID)
	d.logger.Infof("success to deactivate the device[%d]", deviceID)
	return nil
}
