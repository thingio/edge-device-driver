package bus

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/thingio/edge-device-sdk/config"
	"github.com/thingio/edge-device-sdk/logger"
	"strconv"
	"time"
)

func NewMessageBus(opts *config.MessageBusOptions, logger *logger.Logger) (MessageBus, error) {
	mb := messageBus{
		timeout: time.Millisecond * time.Duration(opts.TimeoutMillisecond),
		qos:     opts.QoS,
		routes:  make(map[string]MessageHandler),
		logger:  logger,
	}
	if err := mb.setClient(opts); err != nil {
		return nil, err
	}
	return &mb, nil
}

// MessageBus encapsulates all common manipulations based on MQTT.
type MessageBus interface {
	IsConnected() bool

	Connect() error

	Disconnect() error

	Publish(data Data) error

	Subscribe(handler MessageHandler, topics ...string) error

	Unsubscribe(topics ...string) error

	Call(request Data) (response Data, err error)
}

type messageBus struct {
	client  mqtt.Client
	timeout time.Duration
	qos     int

	routes map[string]MessageHandler // topic -> handler

	logger *logger.Logger
}

func (mb *messageBus) IsConnected() bool {
	return mb.client.IsConnected()
}

func (mb *messageBus) Connect() error {
	if mb.IsConnected() {
		return nil
	}

	token := mb.client.Connect()
	return mb.handleToken(token)
}

func (mb *messageBus) Disconnect() error {
	if mb.IsConnected() {
		mb.client.Disconnect(2000) // waiting 2s
	}
	return nil
}

func (mb *messageBus) Publish(data Data) error {
	msg, err := data.ToMessage()
	if err != nil {
		return err
	}

	token := mb.client.Publish(msg.Topic, byte(mb.qos), false, msg.Payload)
	return mb.handleToken(token)
}

func (mb *messageBus) Subscribe(handler MessageHandler, topics ...string) error {
	filters := make(map[string]byte)
	for _, topic := range topics {
		mb.routes[topic] = handler
		filters[topic] = byte(mb.qos)
	}
	callback := func(mc mqtt.Client, msg mqtt.Message) {
		go handler(&Message{
			Topic:   msg.Topic(),
			Payload: msg.Payload(),
		})
	}

	token := mb.client.SubscribeMultiple(filters, callback)
	return mb.handleToken(token)
}

func (mb *messageBus) Unsubscribe(topics ...string) error {
	for _, topic := range topics {
		delete(mb.routes, topic)
	}

	token := mb.client.Unsubscribe(topics...)
	return mb.handleToken(token)
}

func (mb *messageBus) Call(request Data) (response Data, err error) {
	response, err = request.Response()
	if err != nil {
		return nil, err
	}

	// subscribe response
	rspMsg, err := response.ToMessage()
	if err != nil {
		return nil, err
	}
	ch := make(chan *Message, 1)
	rspTpc := rspMsg.Topic
	if err = mb.Subscribe(func(msg *Message) {
		ch <- msg
	}, rspTpc); err != nil {
		return nil, err
	}
	defer func() {
		close(ch)
		_ = mb.Unsubscribe(rspTpc)
	}()

	// publish request
	if err = mb.Publish(request); err != nil {
		return nil, err
	}
	// waiting for the response
	ticker := time.NewTicker(mb.timeout)
	select {
	case msg := <-ch:
		_, fields, err := msg.Parse()
		if err != nil {
			return nil, fmt.Errorf("fail to parse the message: %s, got %s",
				msg.ToString(), err.Error())
		}
		response.SetFields(fields)
		return response, nil
	case <-ticker.C:
		ticker.Stop()
		return nil, fmt.Errorf("call timeout: %dms", mb.timeout/time.Millisecond)
	}
}

func (mb *messageBus) handleToken(token mqtt.Token) error {
	if mb.timeout > 0 {
		token.WaitTimeout(mb.timeout)
	} else {
		token.Wait()
	}
	return token.Error()
}

func (mb *messageBus) setClient(options *config.MessageBusOptions) error {
	opts := mqtt.NewClientOptions()
	clientID := "edge-device-sub-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	mb.logger.Infof("the ID of client for the message bus is %s", clientID)
	opts.SetClientID(clientID)
	opts.AddBroker(fmt.Sprintf("%s://%s:%d", options.Protocol, options.Host, options.Port))
	opts.SetConnectTimeout(time.Duration(options.ConnectTimoutMillisecond) * time.Millisecond)
	opts.SetKeepAlive(time.Minute)
	opts.SetAutoReconnect(true)
	opts.SetOnConnectHandler(mb.onConnect)
	opts.SetConnectionLostHandler(mb.onConnectLost)
	opts.SetCleanSession(options.CleanSession)

	mb.client = mqtt.NewClient(opts)
	return nil
}

func (mb *messageBus) onConnect(mc mqtt.Client) {
	reader := mc.OptionsReader()
	mb.logger.Infof("the connection with %s for the message bus has been established.", reader.Servers()[0].String())

	for tpc, hdl := range mb.routes {
		if err := mb.Subscribe(hdl, tpc); err != nil {
			mb.logger.WithError(err).Errorf("fail to resubscribe the topic: %s", tpc)
		}
	}
}

func (mb *messageBus) onConnectLost(mc mqtt.Client, err error) {
	reader := mc.OptionsReader()
	mb.logger.WithError(err).Errorf("the connection with %s for the message bus has lost, trying to reconnect.",
		reader.Servers()[0].String())
}
