package mqtt

import (
	"fmt"
	"github.com/rs/zerolog"
	"path"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

const (
	Online  string = "online"
	Offline string = "offline"
)

// Topics.
const (
	serverStatus string = "server/status"
)

type Client interface {
	// Connect to the MQTT server.
	Connect() error
	// Disconnect from the MQTT server.
	Disconnect() error

	// Publishes a message under the prefix topic of DigitalStrom.
	Publish(topic string, message interface{}) error
	PublishAndLogError(topic string, message interface{})

	// Return the full topic for a given subpath.
	GetFullTopic(topic string) string
	// Returns the topic used to publish the server status.
	ServerStatusTopic() string

	RawClient() mqtt.Client
}

type client struct {
	mqttClient mqtt.Client
	options    ClientOptions
	log        zerolog.Logger
}

func NewClient(options *ClientOptions) Client {
	logger := log.With().Str("Component", "MQTT").Logger()
	mqttOptions := mqtt.NewClientOptions().
		AddBroker(options.MqttUrl).
		SetClientID("climkit-" + uuid.New().String()).
		SetOrderMatters(false).
		SetUsername(options.Username).
		SetPassword(options.Password)

	return &client{
		mqttClient: mqtt.NewClient(mqttOptions),
		options:    *options,
		log:        logger,
	}
}

func (c *client) Connect() error {
	t := c.mqttClient.Connect()
	<-t.Done()
	if t.Error() != nil {
		return fmt.Errorf("error connecting to MQTT broker '%s': %w", c.options.MqttUrl, t.Error())
	}

	if err := c.publishServerStatus(Online); err != nil {
		return err
	}
	return nil
}

func (c *client) Disconnect() error {
	c.log.Info().Msg("Publishing Offline status to MQTT server.")
	if err := c.publishServerStatus(Offline); err != nil {
		return err
	}
	c.mqttClient.Disconnect(uint(c.options.DisconnectTimeout.Milliseconds()))
	c.log.Info().Msg("Disconnected from MQTT server.")
	return nil
}

func (c *client) Publish(topic string, message interface{}) error {
	t := c.mqttClient.Publish(
		path.Join(c.options.TopicPrefix, topic),
		c.options.QoS,
		c.options.Retain,
		message)
	<-t.Done()
	return t.Error()
}

func (c *client) PublishAndLogError(topic string, message interface{}) {
	err := c.Publish(topic, message)
	if err != nil {
		c.log.Error().Str("topic", topic).Err(err).Msg("Cannot publish")
	}
}

// Publish the current binary status into the MQTT topic.
func (c *client) publishServerStatus(message string) error {
	c.log.Info().Str("status", message).Str("topic", serverStatus).Msg("Updating server status topic")
	return c.Publish(serverStatus, message)
}

func (c *client) ServerStatusTopic() string {
	return path.Join(c.options.TopicPrefix, serverStatus)
}

func (c *client) GetFullTopic(topic string) string {
	return path.Join(c.options.TopicPrefix, topic)
}

func (c *client) RawClient() mqtt.Client {
	return c.mqttClient
}

func normalizeForTopicName(item string) string {
	output := ""
	for i := 0; i < len(item); i++ {
		c := item[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-' {
			output += string(c)
		} else if c == ' ' || c == '/' {
			output += "_"
		}
	}
	return output
}
