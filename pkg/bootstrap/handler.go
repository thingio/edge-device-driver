package bootstrap

import (
	"context"
	"github.com/thingio/edge-device-sdk/internal/di"
	"sync"
)

// Handler defines the contract that each bootstrap handler must fulfill.
// The 1st argument ctx is responsible for shutting down handler gracefully after receiving ctx.Done().
// The 2nd argument wg is used to let the main goroutine exit after all handlers have finished.
// The 3rd argument dic maintains all services, their constructors and their constructed instances.
// The 4th argument startupTimer is responsible for setting and monitoring the startup time of each handler.
// Implementation returns true if the handler completed successfully, false if it did not.
type Handler func(ctx context.Context, wg *sync.WaitGroup, dic *di.Container, startupTimer Timer) (success bool)
