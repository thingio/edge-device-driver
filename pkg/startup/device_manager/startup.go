package startup

import (
	"context"
	"github.com/thingio/edge-device-sdk/pkg/models"
)

func Startup(metaStore models.MetaStore) {
	ctx, cancel := context.WithCancel(context.Background())

	dm := &DeviceManager{}
	dm.Initialize(ctx, cancel, metaStore)
	dm.Serve()
}
