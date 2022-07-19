package modules

import (
	"encoding/json"
	"github.com/gaetancollaud/climkit-to-mqtt/pkg/climkit"
	"github.com/gaetancollaud/climkit-to-mqtt/pkg/config"
	"github.com/gaetancollaud/climkit-to-mqtt/pkg/mqtt"
	"github.com/rs/zerolog/log"
	"time"
)

type MeterModule struct {
	mqttClient mqtt.Client
	climkit    climkit.Client
}

func (c *MeterModule) Start() error {

	installations, err := c.climkit.GetInstallations()
	if err != nil {

		return err
	}
	log.Info().Strs("installations", installations).Msg("installations retrieved")

	for i := range installations {
		info, err := c.climkit.GetInstallationInfo(installations[i])
		if err != nil {
			return err
		}
		infoStr, _ := json.Marshal(info)
		log.Info().RawJSON("info", infoStr).Msg("got installation info")

		meters, err := c.climkit.GetMetersInfos(installations[i])
		metersStr, _ := json.Marshal(meters)
		log.Info().RawJSON("meters", metersStr).Msg("got installation meters")

		timeSeries, err := c.climkit.GetMeterData(installations[i], meters, climkit.Electricity, time.Now().Add(-time.Minute*30))
		timeSeriesStr, _ := json.Marshal(timeSeries)
		log.Info().RawJSON("timeSeries", timeSeriesStr).Msg("got data")
	}
	return nil

}

func (c *MeterModule) Stop() error {
	return nil
}

func NewMeterModule(mqttClient mqtt.Client, climkit climkit.Client, config *config.Config) Module {
	return &MeterModule{
		mqttClient: mqttClient,
		climkit:    climkit,
		// TODO other init stuff
	}
}

func init() {
	Register("meter", NewMeterModule)
}
