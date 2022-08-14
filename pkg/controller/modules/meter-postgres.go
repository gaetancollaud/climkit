package modules

import (
	"encoding/json"
	"github.com/gaetancollaud/climkit-to-mqtt/pkg/climkit"
	"github.com/gaetancollaud/climkit-to-mqtt/pkg/config"
	"github.com/gaetancollaud/climkit-to-mqtt/pkg/mqtt"
	"github.com/gaetancollaud/climkit-to-mqtt/pkg/postgres"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"time"
)

type MeterPostgresModule struct {
	log              zerolog.Logger
	postgresClient   postgres.Client
	climkit          climkit.Client
	timerQuitChannel chan struct{}
	installations    map[string]([]climkit.MeterInfo)
}

func NewMeterPostgresModule(_ mqtt.Client, postgresClient postgres.Client, climkitClient climkit.Client, config *config.Config) Module {
	logger := log.With().Str("Component", "MeterPostgresModule").Logger()
	return &MeterPostgresModule{
		postgresClient: postgresClient,
		climkit:        climkitClient,
		log:            logger,
		installations:  make(map[string]([]climkit.MeterInfo)),
	}
}

func (mm *MeterPostgresModule) Eligible() bool {
	return mm.postgresClient != nil
}

func (mm *MeterPostgresModule) Start() error {
	mm.fetchAndUpdateInstallationInformation()
	//mm.fetchAndPublishInstallationInformation()
	//mm.fetchAndPublishMeterValue()
	//
	//ticker := time.NewTicker(15 * time.Minute)
	//mm.timerQuitChannel = make(chan struct{})
	//
	//go func() {
	//	for {
	//		select {
	//		case <-ticker.C:
	//			mm.fetchAndPublishMeterValue()
	//		case <-mm.timerQuitChannel:
	//			mm.log.Info().Msg("Stopping interval requests")
	//			ticker.Stop()
	//			return
	//		}
	//	}
	//}()
	return nil
}

func (mm *MeterPostgresModule) Stop() error {
	close(mm.timerQuitChannel)
	return nil
}

func init() {
	Register("meter-postgres", NewMeterPostgresModule)
}

func (mm *MeterPostgresModule) fetchAndUpdateInstallationInformation() {
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
		mm.updateInstallation(installationId, info)

		//meters, err := mm.climkit.GetMetersInfos(installationIds[i])
		//metersStr, _ := json.Marshal(meters)
		//mm.log.Info().RawJSON("meters", metersStr).Msg("got installation meters")
		//
		//for j := range meters {
		//	meterInfo := meters[j]
		//	mm.updateMeterInfo(installationId, meterInfo)
		//}
		//
		//mm.installations[installationId] = meters
	}
}

func (mm *MeterPostgresModule) fetchAndPublishMeterValue() {
	for installationId, meters := range mm.installations {
		timeSeries, err := mm.climkit.GetMeterData(installationId, meters, climkit.Electricity, time.Now().Add(-time.Minute*30))
		if err != nil {
			mm.log.Error().Err(err).Msg("Unable to get metric data")
		}
		timeSeriesStr, _ := json.Marshal(timeSeries)
		mm.log.Info().RawJSON("timeSeries", timeSeriesStr).Msg("got data")

		last := timeSeries[len(timeSeries)-1]
		mm.updateMetersLiveValue(installationId, last)
	}
}

func (mm *MeterPostgresModule) updateInstallation(installationId string, installation climkit.InstallationInfo) {
	query := `INSERT INTO t_installations(installation_id, site_ref, name, timezone, creation_date, latitude, longitude)
		VALUES($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (installation_id) DO UPDATE set site_ref=$2, name=$3, timezone=$4, creation_date=$5, latitude=$6, longitude=$7`

	err := mm.postgresClient.Execute(query, installationId, installation.SiteRef, installation.Name, installation.Timezone, installation.CreationDate, installation.Latitude, installation.Longitude)
	if err != nil {
		mm.log.Fatal().Err(err).Msg("Unable to update installation")
	}
	//mm.mqttClient.PublishAndLogError("installation/"+installationId+"/name", installation.Name)
	//mm.mqttClient.PublishAndLogError("installation/"+installationId+"/site_ref", installation.SiteRef)
	//mm.mqttClient.PublishAndLogError("installation/"+installationId+"/timezone", installation.Timezone)
	//mm.mqttClient.PublishAndLogError("installation/"+installationId+"/creationDate", installation.CreationDate)
	//mm.mqttClient.PublishAndLogError("installation/"+installationId+"/latitude", fmt.Sprintf("%f", installation.Latitude))
	//mm.mqttClient.PublishAndLogError("installation/"+installationId+"/longitude", fmt.Sprintf("%f", installation.Longitude))
}

func (mm *MeterPostgresModule) updateMeterInfo(installationId string, meter climkit.MeterInfo) {
	//mm.mqttClient.PublishAndLogError("installation/"+installationId+"/meters/"+meter.Id+"/type", meter.Type)
	//mm.mqttClient.PublishAndLogError("installation/"+installationId+"/meters/"+meter.Id+"/prim_ad", fmt.Sprintf("%d", meter.PrimAd))
	//mm.mqttClient.PublishAndLogError("installation/"+installationId+"/meters/"+meter.Id+"/virtual", fmt.Sprintf("%d", meter.PrimAd))
}

func (mm *MeterPostgresModule) updateMetersLiveValue(installationId string, lastValues climkit.MeterData) {
	//timestamp := lastValues.Timestamp.Format(time.RFC3339)

	//mm.mqttClient.PublishAndLogError("installation/"+installationId+"/prod_total", fmt.Sprintf("%f", lastValues.ProdTotal*4))
	//mm.mqttClient.PublishAndLogError("installation/"+installationId+"/self", fmt.Sprintf("%f", lastValues.Self*4))
	//mm.mqttClient.PublishAndLogError("installation/"+installationId+"/to_ext", fmt.Sprintf("%f", lastValues.ToExt*4))
	//mm.mqttClient.PublishAndLogError("installation/"+installationId+"/timestamp", timestamp)
	//
	//for i := range lastValues.Meters {
	//	meterValue := lastValues.Meters[i]
	//
	//	mm.mqttClient.PublishAndLogError("installation/"+installationId+"/meters/"+meterValue.MeterId+"/ext", fmt.Sprintf("%f", meterValue.Ext*4))
	//	mm.mqttClient.PublishAndLogError("installation/"+installationId+"/meters/"+meterValue.MeterId+"/self", fmt.Sprintf("%f", meterValue.Self*4))
	//	mm.mqttClient.PublishAndLogError("installation/"+installationId+"/meters/"+meterValue.MeterId+"/total", fmt.Sprintf("%f", meterValue.Total*4))
	//	mm.mqttClient.PublishAndLogError("installation/"+installationId+"/meters/"+meterValue.MeterId+"/timestamp", timestamp)
	//}
}
