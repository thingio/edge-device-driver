package startup

import (
	"context"
	"github.com/thingio/edge-device-sdk/pkg/models"
)

func Startup(protocol *models.Protocol, builder models.ConnectorBuilder) {
	ctx, cancel := context.WithCancel(context.Background())

	ds := &DeviceService{}
	ds.Initialize(ctx, cancel, protocol, builder)
	ds.Serve()
}
