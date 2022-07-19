package controller

import (
	"fmt"
	"github.com/gaetancollaud/climkit-to-mqtt/pkg/climkit"
	"github.com/gaetancollaud/climkit-to-mqtt/pkg/config"
	"github.com/gaetancollaud/climkit-to-mqtt/pkg/controller/modules"
	"github.com/gaetancollaud/climkit-to-mqtt/pkg/mqtt"
	"github.com/rs/zerolog/log"
)

type Controller struct {
	climkit    climkit.Climkit
	mqttClient mqtt.Client

	modules map[string]modules.Module
}

func NewController(config *config.Config) *Controller {
	// Create climkit client
	climkitOption := climkit.NewClientOptions().
		SetApiUrl(config.Climkit.ApiUrl).
		SetUsername(config.Climkit.Username).
		SetPassword(config.Climkit.Password)

	climkit := climkit.New(climkitOption)

	mqttOptions := mqtt.NewClientOptions().
		SetMqttUrl(config.Mqtt.MqttUrl).
		SetUsername(config.Mqtt.Username).
		SetPassword(config.Mqtt.Password).
		SetTopicPrefix(config.Mqtt.TopicPrefix).
		SetRetain(config.Mqtt.Retain)
	mqttClient := mqtt.NewClient(mqttOptions)

	controller := Controller{
		climkit:    climkit,
		mqttClient: mqttClient,
		modules:    map[string]modules.Module{},
	}

	for name, builder := range modules.Modules {
		module := builder(mqttClient, climkit, config)
		controller.modules[name] = module
	}

	return &controller
}

func (c *Controller) Start() error {
	log.Info().Msg("Starting controller.")
	if err := c.mqttClient.Connect(); err != nil {
		return fmt.Errorf("error connecting to MQTT client: %w", err)
	}
	//if err := c.dsClient.Connect(); err != nil {
	//	return fmt.Errorf("error connecting to DigitalStrom client: %w", err)
	//}

	for name, module := range c.modules {
		log.Info().Str("module", name).Msg("Starting module.")
		if err := module.Start(); err != nil {
			return fmt.Errorf("error starting module '%s': %w", name, err)
		}
	}

	return nil
}

func (c *Controller) Stop() error {
	log.Info().Msg("Stopping controller.")

	for name, module := range c.modules {
		log.Info().Str("module", name).Msg("Stopping module.")
		if err := module.Stop(); err != nil {
			return fmt.Errorf("error stopping module '%s': %w", name, err)
		}
	}

	if err := c.mqttClient.Disconnect(); err != nil {
		return fmt.Errorf("error disconnecting to MQTT client: %w", err)
	}

	return nil
}
