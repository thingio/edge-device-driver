package mqtt

import (
	"fmt"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/thingio/edge-device-sdk/internal/logger"
	"strconv"
	"sync"
	"time"
)

type Config struct {
	BrokerAddr        string        `json:"broker_addr" yaml:"broker_addr"`
	ClientId          string        `json:"client_id" yaml:"client_id"`
	Qos               int           `json:"qos" yaml:"qos"`
	ConnTimeOut       time.Duration `json:"conn_timeout" yaml:"conn_timeout"`
	PersistentSession bool          `json:"persistent_session" yaml:"persistent_session"`
	SubTopic          string        `json:"sub_topic" yaml:"sub_topic"`
}
type Client interface {
	Start() error
	Stop()
	IsConnected() bool
	Pub(msg *Msg) error
	Unsub(topics ...string) error
	Sub(handler MsgHandler, topics ...string) error
	SubWithChannel(channel chan<- *Msg, topics ...string) error
	GetConfig() Config

	// AddRoute allows you to add a handler for messages on specific topics
	// without making a subscription. For example having a different handler
	// for parts of a wildcard subscription
	AddRoute(handler MsgHandler, topics ...string)

	// RemoveRoute takes a topic string, looks for the correspond Route in the list of Routes. If
	// found it removes the Route from the list.
	RemoveRoute(topics ...string)
}
type mqttClient struct {
	sync.RWMutex
	logger.Logger
	client  mqtt.Client
	timeout time.Duration
	config  Config
	routes  map[string]MsgHandler // map[topic]msg handler

	// topics that this client subscribed, used to restore subscriptions after reconnect
	topics map[string]MsgHandler
}

type MsgHandler func(*Msg)

const MinConnTimeout = time.Second * 3

func NewMQTTClient(config Config) Client {
	if config.ConnTimeOut < MinConnTimeout {
		config.ConnTimeOut = MinConnTimeout
	}
	c := &mqttClient{
		timeout: config.ConnTimeOut,
		config:  config,
		routes:  make(map[string]MsgHandler, 0),
		topics:  make(map[string]MsgHandler, 0),
	}
	c.client = mqtt.NewClient(c.parseOpts(config))
	return c
}

func (c *mqttClient) Start() error {
	tk := c.client.Connect()
	if c.timeout != 0 {
		if !tk.WaitTimeout(c.timeout) {
			return fmt.Errorf("timeout while connecting mqtt client with config %+v", c.config)
		}
	} else {
		tk.Wait()
	}
	return tk.Error()
}

func (c *mqttClient) Stop() {
	if c.client.IsConnected() {
		c.client.Disconnect(200)
	}
}

func (c *mqttClient) IsConnected() bool {
	return c.client.IsConnected()
}

func (c *mqttClient) Pub(msg *Msg) error {
	token := c.client.Publish(msg.Topic, byte(c.config.Qos), false, msg.Payload)
	if c.timeout != 0 {
		token.WaitTimeout(c.timeout)
	} else {
		token.Wait()
	}
	return token.Error()
}

func (c *mqttClient) Unsub(topics ...string) error {
	token := c.client.Unsubscribe(topics...)
	if c.timeout != 0 {
		token.WaitTimeout(c.timeout)
	} else {
		token.Wait()
	}
	return token.Error()
}

func (c *mqttClient) AddRoute(handler MsgHandler, topics ...string) {
	c.Lock()
	for _, topic := range topics {
		c.routes[topic] = handler
	}
	c.Unlock()
}
func (c *mqttClient) RemoveRoute(topics ...string) {
	c.Lock()
	for _, topic := range topics {
		delete(c.routes, topic)
	}
	c.Unlock()
}
func (c *mqttClient) Sub(handler MsgHandler, topics ...string) error {
	filters := make(map[string]byte)
	for _, topic := range topics {
		c.topics[topic] = handler
		filters[topic] = byte(c.config.Qos)
	}
	callback := func(client mqtt.Client, message mqtt.Message) {
		msg := &Msg{message.Topic(), message.Payload()}
		c.Lock()
		defer c.Unlock()
		for topic, route := range c.routes {
			if topic != msg.Topic {
				continue
			}
			go route(msg)
		}
		handler(msg)
	}
	token := c.client.SubscribeMultiple(filters, callback)
	if c.timeout != 0 {
		token.WaitTimeout(c.timeout)
	} else {
		token.Wait()
	}
	return token.Error()
}

func (c *mqttClient) SubWithChannel(channel chan<- *Msg, topics ...string) error {
	return c.Sub(func(msg *Msg) { channel <- msg }, topics...)
}

func (c *mqttClient) GetConfig() Config {
	return c.config
}

func (c *mqttClient) onConnect(mc mqtt.Client) {
	opts := mc.OptionsReader()
	c.Infof("Client %s connected to %s", opts.ClientID(), opts.Servers()[0].String())
	for topic, handler := range c.topics {
		if err := c.Sub(handler, topic); err != nil {
			c.Errorf("%e, failed to re-subscribe topic '%s'", err, topic)
		}
	}
	return
}

func (c *mqttClient) onConnectionLost(mc mqtt.Client, err error) {
	opts := mc.OptionsReader()
	c.Errorf("%e, client %s connection to %s has lost, will try to reconnect",
		err, opts.ClientID(), opts.Servers()[0].String())
	return
}

func (c *mqttClient) parseOpts(config Config) *mqtt.ClientOptions {
	opts := mqtt.NewClientOptions()
	opts.SetConnectTimeout(config.ConnTimeOut)
	opts.SetClientID(config.ClientId + "-sub-" + strconv.FormatInt(time.Now().UnixNano(), 10))
	opts.AddBroker("tcp://" + config.BrokerAddr)
	opts.SetAutoReconnect(true)
	opts.SetKeepAlive(time.Second * 60)
	opts.SetCleanSession(!config.PersistentSession)
	opts.SetOnConnectHandler(c.onConnect)
	opts.SetConnectionLostHandler(c.onConnectionLost)
	return opts
}
