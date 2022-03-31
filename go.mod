module github.com/thingio/edge-device-driver

replace github.com/thingio/edge-device-std v0.2.1 => ../edge-device-std

require (
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/pkg/errors v0.8.1
	github.com/thingio/edge-device-std v0.2.1
)

go 1.16
