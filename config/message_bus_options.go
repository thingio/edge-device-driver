package config

type MessageBusOptions struct {
	// Host is the hostname or IP address of the MQTT broker.
	Host string `json:"host" yaml:"host"`
	// Port is the port of the MQTT broker.
	Port int `json:"port" yaml:"port"`
	// Protocol is the protocol to use when communicating with the MQTT broker, such as "tcp".
	Protocol string `json:"protocol" yaml:"protocol"`

	// ConnectTimoutMillisecond indicates the timeout of connecting to the MQTT broker.
	ConnectTimoutMillisecond int `json:"connect_timout_millisecond" yaml:"connect_timout_millisecond"`
	// TimeoutMillisecond indicates the timeout of manipulations.
	TimeoutMillisecond int `json:"timeout_millisecond" yaml:"timeout_millisecond"`
	// QoS is the abbreviation of MQTT Quality of Service.
	QoS int `json:"qos" yaml:"qos"`
	// CleanSession indicates whether retain messages after reconnecting for QoS1 and QoS2.
	CleanSession bool `json:"clean_session" yaml:"clean_session"`
}
