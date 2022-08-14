package controller

import (
	"fmt"
	"github.com/gaetancollaud/climkit-to-mqtt/pkg/climkit"
	"github.com/gaetancollaud/climkit-to-mqtt/pkg/config"
	"github.com/gaetancollaud/climkit-to-mqtt/pkg/controller/modules"
	"github.com/gaetancollaud/climkit-to-mqtt/pkg/mqtt"
	"github.com/gaetancollaud/climkit-to-mqtt/pkg/postgres"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Controller struct {
	climkit        climkit.Client
	mqttClient     mqtt.Client
	postgresClient postgres.Client
	log            zerolog.Logger

	modules map[string]modules.Module
}

func NewController(cfg *config.Config) *Controller {
	logger := log.With().Str("Component", "Controller").Logger()
	// Create climkit client
	climkitOption := climkit.NewClientOptions().
		SetApiUrl(cfg.Climkit.ApiUrl).
		SetUsername(cfg.Climkit.Username).
		SetPassword(cfg.Climkit.Password)

	climkit := climkit.NewClient(climkitOption)

	var mqttClient mqtt.Client
	if cfg.Mode == config.Mqtt {
		mqttOptions := mqtt.NewClientOptions().
			SetMqttUrl(cfg.Mqtt.MqttUrl).
			SetUsername(cfg.Mqtt.Username).
			SetPassword(cfg.Mqtt.Password).
			SetTopicPrefix(cfg.Mqtt.TopicPrefix).
			SetRetain(cfg.Mqtt.Retain)
		mqttClient = mqtt.NewClient(mqttOptions)
	}

	var postgresClient postgres.Client
	if cfg.Mode == config.Postgres {
		postgresOptions := postgres.NewClientOptions().
			SetHost(cfg.Postgres.Host).
			SetPort(cfg.Postgres.Port).
			SetDatabase(cfg.Postgres.Database).
			SetUsername(cfg.Mqtt.Username).
			SetPassword(cfg.Mqtt.Password)
		postgresClient = postgres.NewClient(postgresOptions)
	}

	controller := Controller{
		climkit:        climkit,
		mqttClient:     mqttClient,
		postgresClient: postgresClient,
		log:            logger,
		modules:        map[string]modules.Module{},
	}

	for name, builder := range modules.Modules {
		module := builder(mqttClient, postgresClient, climkit, cfg)
		controller.modules[name] = module
	}

	return &controller
}

func (c *Controller) Start() error {
	c.log.Info().Msg("Starting.")
	if c.mqttClient != nil {
		if err := c.mqttClient.Connect(); err != nil {
			return fmt.Errorf("error connecting to MQTT client: %w", err)
		}
	}
	if c.postgresClient != nil {
		if err := c.postgresClient.Connect(); err != nil {
			return fmt.Errorf("error connecting to Postgres client: %w", err)
		}
	}

	for name, module := range c.modules {
		if module.Eligible() {
			c.log.Info().Str("module", name).Msg("Starting module.")
			if err := module.Start(); err != nil {
				return fmt.Errorf("error starting module '%s': %w", name, err)
			}
		} else {
			c.log.Debug().Str("module", name).Msg("Not eligible to start")
		}
	}

	return nil
}

func (c *Controller) Stop() error {
	c.log.Info().Msg("Stopping.")

	for name, module := range c.modules {
		if module.Eligible() {
			c.log.Info().Str("module", name).Msg("Stopping module.")
			if err := module.Stop(); err != nil {
				return fmt.Errorf("error stopping module '%s': %w", name, err)
			}
		}
	}

	if c.mqttClient != nil {
		if err := c.mqttClient.Disconnect(); err != nil {
			return fmt.Errorf("error disconnecting from MQTT client: %w", err)
		}
	}
	if c.postgresClient != nil {
		if err := c.postgresClient.Disconnect(); err != nil {
			return fmt.Errorf("error disconnecting from postgres client: %w", err)
		}
	}

	return nil
}
