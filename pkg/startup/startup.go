package startup

import (
	"context"
	"github.com/thingio/edge-device-driver/internal/driver"
	"github.com/thingio/edge-device-std/models"
)

func Startup(protocol *models.Protocol, builder models.DeviceTwinBuilder) {
	ctx, cancel := context.WithCancel(context.Background())

	ds, err := driver.NewDeviceDriver(ctx, cancel, protocol, builder)
	if err != nil {
		panic(err)
	}
	ds.Initialize()
	ds.Serve()
}
