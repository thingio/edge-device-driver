package config

import (
	"fmt"
	"github.com/jinzhu/configor"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func init() {
	if _, err := LoadConfiguration(); err != nil {
		panic("fail to load configuration")
	}
}

const (
	ConfigurationPath = "./resources/configuration.yaml" // xxx/resource/configuration.yaml
	ProtocolsPath     = "./resources/protocols"          // xxx-device-conn/cmd/resources/protocols
	ProductsPath      = "./resources/products"           // edge-device-manager/resources/products
	DevicesPath       = "./resources/devices"            // edge-device-manager/resources/devices

	EnvSep = ","
)

var C *Configuration

type Configuration struct {
	MessageBus MessageBusOptions `json:"message_bus" yaml:"message_bus"`
}

func LoadConfiguration() (*Configuration, error) {
	C = new(Configuration)

	path := os.Getenv("CONFIG_FILE")
	if path == "" {
		path = ConfigurationPath
	}
	if err := configor.Load(C, path); err != nil {
		return nil, fmt.Errorf("failed to load configuration file, got %s", err.Error())
	}

	if value := reflect.ValueOf(C).Elem().FieldByName("MessageBus"); value.IsValid() {
		m := value.Interface().(MessageBusOptions)
		LoadMessageBusOptions(&m)
		value.Set(reflect.ValueOf(m))
	}
	return C, nil
}

func LoadEnv(target *string, env string) {
	value := os.Getenv(env)
	if value != "" {
		*target = value
	}
}

// LoadEnvs will read the values of given environments,
// then overwrite the pointer if value is not empty.
func LoadEnvs(envs map[string]interface{}) error {
	var err error
	for env, target := range envs {
		value := os.Getenv(env)
		if value == "" {
			continue
		}
		switch target.(type) {
		case *string:
			*(target.(*string)) = value
		case *[]string:
			values := strings.Split(value, EnvSep)
			result := make([]string, 0)
			for _, v := range values {
				if v != "" {
					result = append(result, v)
				}
			}
			if len(result) != 0 {
				*(target.(*[]string)) = result
			}
		case *int:
			*(target.(*int)), err = strconv.Atoi(value)
		case *bool:
			*(target.(*bool)), err = strconv.ParseBool(value)
		case *int64:
			*(target.(*int64)), err = strconv.ParseInt(value, 10, 64)
		case *float32:
			if v, err := strconv.ParseFloat(value, 32); err == nil {
				*(target.(*float32)) = float32(v)
			}
		case *float64:
			*(target.(*float64)), err = strconv.ParseFloat(value, 64)
		case *time.Duration:
			*(target.(*time.Duration)), err = time.ParseDuration(value)
		default:
			return fmt.Errorf("unsupported env type : %T", target)
		}
		if err != nil {
			return fmt.Errorf("fail to load environments, got %s", err.Error())
		}
	}
	return nil
}

func LoadEnvList(target *[]string, env string) {
	value := os.Getenv(env)
	values := strings.Split(value, EnvSep)
	result := make([]string, 0)
	for _, v := range values {
		if v != "" {
			result = append(result, v)
		}
	}
	if len(result) != 0 {
		*target = result
	}
}

func LoadEnvBool(target *bool, env string) {
	value := os.Getenv(env)
	if value != "" {
		*target, _ = strconv.ParseBool(value)
	}
}

func LoadEnvInt(target *int, env string) {
	value := os.Getenv(env)
	if value != "" {
		*target, _ = strconv.Atoi(value)
	}
}

func LoadEnvInt64(target *int64, env string) {
	value := os.Getenv(env)
	if value != "" {
		*target, _ = strconv.ParseInt(value, 10, 64)
	}
}
