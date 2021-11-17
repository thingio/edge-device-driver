package startup

import (
	"os"
)

func (s *DeviceService) registerProtocol() {
	if err := s.moc.RegisterProtocol(s.protocol); err != nil {
		s.logger.WithError(err).Errorf("fail to register the protocol[%s] "+
			"to the device manager", s.protocol.ID)
		os.Exit(1)
	}
	s.logger.Infof("success to register the protocol[%s] to the device manager", s.protocol.ID)
}
