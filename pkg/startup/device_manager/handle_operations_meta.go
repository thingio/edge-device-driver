package startup

import (
	"github.com/thingio/edge-device-sdk/pkg/models"
	"time"
)

func (m *DeviceManager) watchingProtocols() {
	defer func() {
		m.wg.Done()
	}()

	if err := m.moc.OnRegisterProtocols(m.registerProtocol); err != nil {
		m.logger.WithError(err).Error("fail to wait for registering protocols")
		return
	}
	if err := m.moc.OnUnregisterProtocols(m.unregisterProtocol); err != nil {
		m.logger.WithError(err).Error("fail to wait for unregistering protocols")
		return
	}

	ticker := time.NewTicker(1 * time.Minute)
	for {
		select {
		case <-ticker.C: // check the connection of protocol
			m.protocols.Range(func(key, value interface{}) bool {
				// TODO PING
				return true
			})
		case <-m.ctx.Done():
			break
		}
	}
}

func (m *DeviceManager) registerProtocol(protocol *models.Protocol) error {
	m.protocols.Store(protocol.ID, protocol)
	m.logger.Infof("the protocol[%s] has registered successfully", protocol.ID)
	return nil
}

func (m *DeviceManager) unregisterProtocol(protocolID string) error {
	m.protocols.Delete(protocolID)
	m.logger.Infof("the protocol[%s] has unregistered successfully", protocolID)
	return nil
}

func (m *DeviceManager) watchingProductOperations() {
	defer func() {
		m.wg.Done()
	}()

	if err := m.moc.OnListProducts(m.metaStore.ListProducts); err != nil {
		m.logger.WithError(err).Error("fail to wait for listing products")
		return
	} else {
		m.logger.Infof("start to watch the changes for product...")
	}
}

func (m *DeviceManager) watchingDeviceOperations() {
	defer func() {
		m.wg.Done()
	}()

	if err := m.moc.OnListDevices(m.metaStore.ListDevices); err != nil {
		m.logger.WithError(err).Error("fail to wait for listing devices")
		return
	} else {
		m.logger.Infof("start to watch the changes for device...")
	}
}
