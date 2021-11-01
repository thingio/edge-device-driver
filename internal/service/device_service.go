package service

import (
	"context"
	"github.com/thingio/edge-device-sdk/internal/logger"
	"github.com/thingio/edge-device-sdk/pkg/models"
)

func Run(ctx context.Context, cancel context.CancelFunc,
	serviceName string, serviceVersion string, driver models.ProtocolDriver) {

	ds := &DeviceService{
		Name:   serviceName,
		driver: driver,
		logger: logger.NewLogger(),
	}

	ds.Stop(false)
}

type DeviceService struct {
	ID      string
	Name    string
	Version string
	driver  models.ProtocolDriver

	logger *logger.Logger
}

// Stop shuts down the service.
func (s *DeviceService) Stop(force bool) {
	if err := s.driver.Stop(force); err != nil {
		s.logger.Error(err.Error())
	}
}
