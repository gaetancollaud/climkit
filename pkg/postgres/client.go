package postgres

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Client interface {
	// Connect to Postgres
	Connect() error
	// Disconnect from Postgres
	Disconnect() error

	Execute(query string) error
	Select(query string) interface{}
}

type client struct {
	//mqttClient mqtt.Client
	options ClientOptions
	log     zerolog.Logger
}

func NewClient(options *ClientOptions) Client {
	logger := log.With().Str("Component", "MQTT").Logger()
	//mqttOptions := mqtt.NewClientOptions().
	//	AddBroker(options.MqttUrl).
	//	SetClientID("climkit-to-mqtt-" + uuid.New().String()).
	//	SetOrderMatters(false).
	//	SetUsername(options.Username).
	//	SetPassword(options.Password)

	return &client{
		//mqttClient: mqtt.NewClient(mqttOptions),
		options: *options,
		log:     logger,
	}
}

func (c client) Connect() error {
	return nil
}

func (c client) Disconnect() error {
	return nil
}

func (c client) Execute(query string) error {
	//TODO implement me
	panic("implement me")
}

func (c client) Select(query string) interface{} {
	//TODO implement me
	panic("implement me")
}
