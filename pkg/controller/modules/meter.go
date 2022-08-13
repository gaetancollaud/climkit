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
	log              zerolog.Logger
	mqttClient       mqtt.Client
	climkit          climkit.Client
	timerQuitChannel chan struct{}
	installations    map[string]([]climkit.MeterInfo)
}

func NewMeterModule(mqttClient mqtt.Client, climkitClient climkit.Client, config *config.Config) Module {
	logger := log.With().Str("Component", "MeterModule").Logger()
	return &MeterModule{
		mqttClient:    mqttClient,
		climkit:       climkitClient,
		log:           logger,
		installations: make(map[string]([]climkit.MeterInfo)),
		// TODO other init stuff
	}
}

func (mm *MeterModule) Start() error {

	mm.fetchAndPublishInstallationInformation()
	mm.fetchAndPublishMeterValue()

	ticker := time.NewTicker(15 * time.Minute)
	mm.timerQuitChannel = make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:
				mm.fetchAndPublishMeterValue()
			case <-mm.timerQuitChannel:
				mm.log.Info().Msg("Stopping interval requests")
				ticker.Stop()
				return
			}
		}
	}()

	return nil
}

func (mm *MeterModule) Stop() error {
	close(mm.timerQuitChannel)
	return nil
}

func init() {
	Register("meter", NewMeterModule)
}

func (mm *MeterModule) fetchAndPublishInstallationInformation() {
	installationIds, err := mm.climkit.GetInstallationIds()
	if err != nil {
		mm.log.Error().Err(err).Msg("Unable to get installations liost")
	}
	mm.log.Info().Strs("installationIds", installationIds).Msg("installation retrieved")

	for i := range installationIds {
		installationId := installationIds[i]
		info, err := mm.climkit.GetInstallationInfo(installationId)
		if err != nil {
			mm.log.Error().Err(err).Msg("Unable to get installation information")
		}
		infoStr, _ := json.Marshal(info)
		mm.log.Info().RawJSON("info", infoStr).Msg("got installation info")
		mm.publishInstallation(installationId, info)

		meters, err := mm.climkit.GetMetersInfos(installationIds[i])
		metersStr, _ := json.Marshal(meters)
		mm.log.Info().RawJSON("meters", metersStr).Msg("got installation meters")

		for j := range meters {
			meterInfo := meters[j]
			mm.publishMeterInfo(installationId, meterInfo)
		}

		mm.installations[installationId] = meters
	}
}

func (mm *MeterModule) fetchAndPublishMeterValue() {
	for installationId, meters := range mm.installations {
		timeSeries, err := mm.climkit.GetMeterData(installationId, meters, climkit.Electricity, time.Now().Add(-time.Minute*30))
		if err != nil {
			mm.log.Error().Err(err).Msg("Unable to get metric data")
		}
		timeSeriesStr, _ := json.Marshal(timeSeries)
		mm.log.Info().RawJSON("timeSeries", timeSeriesStr).Msg("got data")

		last := timeSeries[len(timeSeries)-1]
		mm.publishMetersLiveValue(installationId, last)
	}
}

func (mm *MeterModule) publishInstallation(installationId string, installation climkit.InstallationInfo) {
	mm.mqttClient.PublishAndLogError("installation/"+installationId+"/name", installation.Name)
	mm.mqttClient.PublishAndLogError("installation/"+installationId+"/site_ref", installation.SiteRef)
	mm.mqttClient.PublishAndLogError("installation/"+installationId+"/timezone", installation.Timezone)
	mm.mqttClient.PublishAndLogError("installation/"+installationId+"/creationDate", installation.CreationDate)
	mm.mqttClient.PublishAndLogError("installation/"+installationId+"/latitude", fmt.Sprintf("%f", installation.Latitude))
	mm.mqttClient.PublishAndLogError("installation/"+installationId+"/longitude", fmt.Sprintf("%f", installation.Longitude))
}

func (mm *MeterModule) publishMeterInfo(installationId string, meter climkit.MeterInfo) {
	mm.mqttClient.PublishAndLogError("installation/"+installationId+"/meters/"+meter.Id+"/type", meter.Type)
	mm.mqttClient.PublishAndLogError("installation/"+installationId+"/meters/"+meter.Id+"/prim_ad", fmt.Sprintf("%d", meter.PrimAd))
	mm.mqttClient.PublishAndLogError("installation/"+installationId+"/meters/"+meter.Id+"/virtual", fmt.Sprintf("%d", meter.PrimAd))
}

func (mm *MeterModule) publishMetersLiveValue(installationId string, lastValues climkit.MeterData) {
	timestamp := lastValues.Timestamp.Format(time.RFC3339)

	mm.mqttClient.PublishAndLogError("installation/"+installationId+"/prod_total", fmt.Sprintf("%f", lastValues.ProdTotal*4))
	mm.mqttClient.PublishAndLogError("installation/"+installationId+"/self", fmt.Sprintf("%f", lastValues.Self*4))
	mm.mqttClient.PublishAndLogError("installation/"+installationId+"/to_ext", fmt.Sprintf("%f", lastValues.ToExt*4))
	mm.mqttClient.PublishAndLogError("installation/"+installationId+"/timestamp", timestamp)

	for i := range lastValues.Meters {
		meterValue := lastValues.Meters[i]

		mm.mqttClient.PublishAndLogError("installation/"+installationId+"/meters/"+meterValue.MeterId+"/ext", fmt.Sprintf("%f", meterValue.Ext*4))
		mm.mqttClient.PublishAndLogError("installation/"+installationId+"/meters/"+meterValue.MeterId+"/self", fmt.Sprintf("%f", meterValue.Self*4))
		mm.mqttClient.PublishAndLogError("installation/"+installationId+"/meters/"+meterValue.MeterId+"/total", fmt.Sprintf("%f", meterValue.Total*4))
		mm.mqttClient.PublishAndLogError("installation/"+installationId+"/meters/"+meterValue.MeterId+"/timestamp", timestamp)
	}
}
