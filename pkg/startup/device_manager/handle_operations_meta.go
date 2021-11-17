package startup

import (
	"github.com/thingio/edge-device-sdk/pkg/models"
	"time"
)

func (m *DeviceManager) watchingProtocols() {
	defer func() {
		m.wg.Done()
	}()

	if err := m.moc.RegisterProtocols(m.registerProtocol); err != nil {
		m.logger.WithError(err).Error("fail to waiting for watching protocols")
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
