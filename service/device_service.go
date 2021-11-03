package service

import (
	"encoding/json"
	"fmt"
	"github.com/thingio/edge-device-sdk/driver"
	"github.com/thingio/edge-device-sdk/internal/logger"
	"github.com/thingio/edge-device-sdk/models"
	"github.com/thingio/edge-device-sdk/mqtt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
)

func Run(protocol models.Protocol, driver driver.Builder) {
	ds := &DeviceService{
		ID:       protocol.ID,
		Name:     protocol.Name,
		Version:  "0",
		driver:   driver,
		protocol: protocol,
		Logger:   logger.NewLogger(),
	}
	if err := ds.Start(); err != nil {
		panic(fmt.Errorf("failed to start the device service, %e", err))
	}

	select {}
}

func LoadProtocol(path string) (*models.Protocol, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open the protocol config file, %e", err)
	}

	bs, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read the content of the protocol file '%s', %e", path, err)
	}
	var unmarshaler func([]byte, interface{}) error
	switch ext := filepath.Ext(path); ext {
	default:
		return nil, fmt.Errorf("invalid protocol file extension '%s', only supported [json,yml,yaml]", ext)
	case "json":
		unmarshaler = json.Unmarshal
	case "yml", "yaml":
		unmarshaler = yaml.Unmarshal
	}
	protocol := new(models.Protocol)
	if err := unmarshaler(bs, protocol); err != nil {
		return nil, fmt.Errorf("failed to unmarshal protocol, %e", err)
	}
	return protocol, nil
}

type DeviceService struct {
	ID       string
	Name     string
	Version  string
	driver   driver.Builder
	protocol models.Protocol

	mqttCli mqtt.Client // 接收控制请求，发送数据
	*logger.Logger
}

func (s *DeviceService) Start() error {
	return nil
}
