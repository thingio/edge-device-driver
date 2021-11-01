package bootstrap

import (
	"context"
	"github.com/thingio/edge-device-sdk/internal/service"
	"github.com/thingio/edge-device-sdk/pkg/models"
)

func Bootstrap(serviceName string, serviceVersion string, driver models.ProtocolDriver) {
	ctx, cancel := context.WithCancel(context.Background())
	service.Run(ctx, cancel, serviceName, serviceVersion, driver)
}
