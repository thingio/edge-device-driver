package startup

import (
	"context"
)

func Startup() {
	ctx, cancel := context.WithCancel(context.Background())

	dm := &DeviceManager{}
	dm.Initialize(ctx, cancel)
	dm.Serve()
}
