package conf

import "time"
var DeviceServiceConfig Config
type Config struct {
	ConnectTimeout time.Duration
}
