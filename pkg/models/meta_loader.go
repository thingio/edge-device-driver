package models

import (
	"encoding/json"
	"fmt"
	"github.com/thingio/edge-device-sdk/config"
	"gopkg.in/yaml.v2"
	"io/fs"
	"io/ioutil"
	"path/filepath"
)

func LoadProtocol(path string) (*Protocol, error) {
	protocol := new(Protocol)
	if err := load(path, protocol); err != nil {
		return nil, err
	}
	return protocol, nil
}

func LoadProducts(protocolID string) []*Product {
	products := make([]*Product, 0)
	_ = filepath.Walk(config.ProductsPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		var product *Product
		product, err = loadProduct(path)
		if err != nil {
			return err
		}
		if product.Protocol != protocolID {
			return nil
		}
		products = append(products, product)
		return nil
	})
	return products
}

func LoadDevices(productID string) []*Device {
	devices := make([]*Device, 0)
	_ = filepath.Walk(config.DevicesPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		var device *Device
		device, err = loadDevice(path)
		if err != nil {
			return err
		}
		if device.ProductID != productID {
			return nil
		}
		devices = append(devices, device)
		return nil
	})
	return devices
}

func loadProduct(path string) (*Product, error) {
	product := new(Product)
	if err := load(path, product); err != nil {
		return nil, err
	}
	return product, nil
}

func loadDevice(path string) (*Device, error) {
	device := new(Device)
	if err := load(path, device); err != nil {
		return nil, err
	}
	return device, nil
}

func load(path string, meta interface{}) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("fail to load the meta configurtion stored in %s, got %s",
			path, err.Error())
	}

	var unmarshaller func([]byte, interface{}) error
	switch ext := filepath.Ext(path); ext {
	case ".json":
		unmarshaller = json.Unmarshal
	case ".yaml", ".yml":
		unmarshaller = yaml.Unmarshal
	default:
		return fmt.Errorf("invalid meta configuration extension %s, only supporing json/yaml/yml", ext)
	}

	if err := unmarshaller(data, meta); err != nil {
		return fmt.Errorf("fail to unmarshal the device configuration, got %s", err.Error())
	}
	return nil
}
