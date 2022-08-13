package modules

import (
	"encoding/json"
	"fmt"
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
		mm.publishInstallation(info)

		meters, err := mm.climkit.GetMetersInfos(installations[i])
		metersStr, _ := json.Marshal(meters)
		mm.log.Info().RawJSON("meters", metersStr).Msg("got installation meters")

		for j := range meters {
			meterInfo := meters[j]
			mm.publishMeterInfo(info, meterInfo)
		}

		timeSeries, err := mm.climkit.GetMeterData(installations[i], meters, climkit.Electricity, time.Now().Add(-time.Minute*30))
		timeSeriesStr, _ := json.Marshal(timeSeries)
		mm.log.Info().RawJSON("timeSeries", timeSeriesStr).Msg("got data")

		last := timeSeries[len(timeSeries)-1]
		mm.publishMetersLiveValue(info, last)
	}
	return nil

}

func (mm *MeterModule) Stop() error {
	return nil
}

func init() {
	Register("meter", NewMeterModule)
}

func (mm *MeterModule) publishInstallation(installation climkit.InstallationInfo) {
	mm.mqttClient.PublishAndLogError("installation/"+installation.SiteRef+"/name", installation.Name)
	mm.mqttClient.PublishAndLogError("installation/"+installation.SiteRef+"/timezone", installation.Timezone)
	mm.mqttClient.PublishAndLogError("installation/"+installation.SiteRef+"/creationDate", installation.CreationDate)
	mm.mqttClient.PublishAndLogError("installation/"+installation.SiteRef+"/latitude", fmt.Sprintf("%f", installation.Latitude))
	mm.mqttClient.PublishAndLogError("installation/"+installation.SiteRef+"/longitude", fmt.Sprintf("%f", installation.Longitude))
}

func (mm *MeterModule) publishMeterInfo(installation climkit.InstallationInfo, meter climkit.MeterInfo) {
	mm.mqttClient.PublishAndLogError("installation/"+installation.SiteRef+"/meters/"+meter.Id+"/type", meter.Type)
	mm.mqttClient.PublishAndLogError("installation/"+installation.SiteRef+"/meters/"+meter.Id+"/prim_ad", fmt.Sprintf("%d", meter.PrimAd))
	mm.mqttClient.PublishAndLogError("installation/"+installation.SiteRef+"/meters/"+meter.Id+"/virtual", fmt.Sprintf("%d", meter.PrimAd))
}

func (mm *MeterModule) publishMetersLiveValue(installation climkit.InstallationInfo, lastValues climkit.MeterData) {
	mm.mqttClient.PublishAndLogError("installation/"+installation.SiteRef+"/prod_total", fmt.Sprintf("%f", lastValues.ProdTotal*4))
	mm.mqttClient.PublishAndLogError("installation/"+installation.SiteRef+"/self", fmt.Sprintf("%f", lastValues.Self*4))
	mm.mqttClient.PublishAndLogError("installation/"+installation.SiteRef+"/to_ext", fmt.Sprintf("%f", lastValues.ToExt*4))

	for i := range lastValues.Meters {
		meterValue := lastValues.Meters[i]

		mm.mqttClient.PublishAndLogError("installation/"+installation.SiteRef+"/meters/"+meterValue.MeterId+"/ext", fmt.Sprintf("%f", meterValue.Ext*4))
		mm.mqttClient.PublishAndLogError("installation/"+installation.SiteRef+"/meters/"+meterValue.MeterId+"/ext", fmt.Sprintf("%f", meterValue.Self*4))
		mm.mqttClient.PublishAndLogError("installation/"+installation.SiteRef+"/meters/"+meterValue.MeterId+"/ext", fmt.Sprintf("%f", meterValue.Total*4))
	}
}
