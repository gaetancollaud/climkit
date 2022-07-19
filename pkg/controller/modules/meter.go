package modules

import (
	"github.com/gaetancollaud/climkit-to-mqtt/pkg/climkit"
	"github.com/gaetancollaud/climkit-to-mqtt/pkg/config"
	"github.com/gaetancollaud/climkit-to-mqtt/pkg/mqtt"
)

type MeterModule struct {
	mqttClient mqtt.Client
	climkit    climkit.Climkit
}

func (c *MeterModule) Start() error {
	return nil
}

func (c *MeterModule) Stop() error {
	return nil
}

func NewMeterModule(mqttClient mqtt.Client, climkit climkit.Climkit, config *config.Config) Module {
	return &MeterModule{
		mqttClient: mqttClient,
		climkit:    climkit,
		// TODO other init stuff
	}
}

func init() {
	Register("meter", NewMeterModule)
}
