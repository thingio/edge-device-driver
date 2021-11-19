package startup

import (
	"github.com/thingio/edge-device-sdk/pkg/models"
	"os"
)

// registerProtocol tries to register the protocol to the device manager.
func (s *DeviceService) registerProtocol() {
	if err := s.moc.RegisterProtocol(s.protocol); err != nil {
		s.logger.WithError(err).Errorf("fail to register the protocol[%s] "+
			"to the device manager", s.protocol.ID)
		os.Exit(1)
	}
	s.logger.Infof("success to register the protocol[%s] to the device manager", s.protocol.ID)
}

// unregisterProtocol tries to unregister the protocol from the device manager.
func (s *DeviceService) unregisterProtocol() {
	if err := s.moc.UnregisterProtocol(s.protocol.ID); err != nil {
		s.logger.WithError(err).Errorf("fail to unregister the protocol[%s] "+
			"from the device manager", s.protocol.ID)
	}
	s.logger.Infof("success to unregister the protocol[%s] from the device manager", s.protocol.ID)
}

// activateDevices tries to activate all devices.
func (s *DeviceService) activateDevices() {
	products, err := s.moc.ListProducts(s.protocol.ID)
	if err != nil {
		s.logger.WithError(err).Error("fail to fetch products from the device manager")
		os.Exit(1)
	}
	for _, product := range products {
		devices, err := s.moc.ListDevices(product.ID)
		if err != nil {
			s.logger.WithError(err).Error("fail to fetch devices from the device manager")
			os.Exit(1)
		}

		s.products.Store(product.ID, product)
		for _, device := range devices {
			if err := s.activateDevice(device); err != nil {
				s.logger.WithError(err).Errorf("fail to activate the device[%s]", device.ID)
				continue
			}
			s.devices.Store(device.ID, device)
		}
	}
}

// activateDevice is responsible for establishing the connection with the real device.
func (s *DeviceService) activateDevice(device *models.Device) error {
	if connector, _ := s.getDeviceConnector(device.ID); connector != nil { // the device has been activated
		_ = s.deactivateDevice(device.ID)
	}

	// build, initialize and start connector
	product, err := s.getProduct(device.ProductID)
	if err != nil {
		return err
	}
	connector, err := s.connectorBuilder(product, device)
	if err != nil {
		return err
	}
	if err := connector.Initialize(s.logger); err != nil {
		s.logger.WithError(err).Error("fail to initialize the random device connector")
		return err
	}
	if err := connector.Start(); err != nil {
		s.logger.WithError(err).Error("fail to start the random device connector")
		return err
	}

	if len(product.Properties) != 0 {
		if err := connector.Watch(s.bus); err != nil {
			s.logger.WithError(err).Error("fail to watch properties for the device")
			return err
		}
	}

	for _, event := range product.Events {
		if err := connector.Subscribe(event.Id, s.bus); err != nil {
			s.logger.WithError(err).Errorf("fail to subscribe the product event[%s]", event.Id)
			continue
		}
	}

	s.deviceConnectors.Store(device.ID, connector)
	s.logger.Infof("success to activate the device[%s]", device.ID)
	return nil
}

// deactivateDevices tries to deactivate all devices.
func (s *DeviceService) deactivateDevices() {
	s.deviceConnectors.Range(func(key, value interface{}) bool {
		deviceID := key.(string)
		if err := s.deactivateDevice(deviceID); err != nil {
			s.logger.WithError(err).Errorf("fail to deactivate the device[%s]", deviceID)
		}
		return true
	})
}

// deactivateDevice is responsible for breaking up the connection with the real device.
func (s *DeviceService) deactivateDevice(deviceID string) error {
	connector, _ := s.getDeviceConnector(deviceID)
	if connector == nil {
		return nil
	}
	if err := connector.Stop(false); err != nil {
		return err
	}

	s.deviceConnectors.Delete(deviceID)
	s.logger.Infof("success to deactivate the device[%s]", deviceID)
	return nil
}
