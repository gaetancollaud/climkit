package modules

import (
	"encoding/json"
	"github.com/gaetancollaud/climkit-to-mqtt/pkg/climkit"
	"github.com/gaetancollaud/climkit-to-mqtt/pkg/config"
	"github.com/gaetancollaud/climkit-to-mqtt/pkg/mqtt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"time"
)

type MeterModule struct {
	mqttClient mqtt.Client
	climkit    climkit.Client
	log        zerolog.Logger
}

func NewMeterModule(mqttClient mqtt.Client, climkit climkit.Client, config *config.Config) Module {
	logger := log.With().Str("Component", "MeterModule").Logger()
	return &MeterModule{
		mqttClient: mqttClient,
		climkit:    climkit,
		log:        logger,
		// TODO other init stuff
	}
}

func (mm *MeterModule) Start() error {

	installations, err := mm.climkit.GetInstallations()
	if err != nil {

		return err
	}
	mm.log.Info().Strs("installations", installations).Msg("installations retrieved")

	for i := range installations {
		info, err := mm.climkit.GetInstallationInfo(installations[i])
		if err != nil {
			return err
		}
		infoStr, _ := json.Marshal(info)
		mm.log.Info().RawJSON("info", infoStr).Msg("got installation info")

		meters, err := mm.climkit.GetMetersInfos(installations[i])
		metersStr, _ := json.Marshal(meters)
		mm.log.Info().RawJSON("meters", metersStr).Msg("got installation meters")

		timeSeries, err := mm.climkit.GetMeterData(installations[i], meters, climkit.Electricity, time.Now().Add(-time.Minute*30))
		timeSeriesStr, _ := json.Marshal(timeSeries)
		mm.log.Info().RawJSON("timeSeries", timeSeriesStr).Msg("got data")
	}
	return nil

}

func (mm *MeterModule) Stop() error {
	return nil
}

func init() {
	Register("meter", NewMeterModule)
}
