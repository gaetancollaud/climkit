package modules

import (
	"database/sql"
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
	mm.fetchAndUpdateInstallationHistory()

	ticker := time.NewTicker(15 * time.Minute)
	mm.timerQuitChannel = make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:
				mm.fetchAndUpdateInstallationHistory()
			case <-mm.timerQuitChannel:
				mm.log.Info().Msg("Stopping interval requests")
				ticker.Stop()
				return
			}
		}
	}()
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

		meters, err := mm.climkit.GetMetersInfos(installationIds[i])
		metersStr, _ := json.Marshal(meters)
		mm.log.Info().RawJSON("meters", metersStr).Msg("got installation meters")

		for j := range meters {
			meterInfo := meters[j]
			mm.updateMeterInfo(installationId, meterInfo)
		}

		mm.installations[installationId] = meters
	}
}
func (mm *MeterPostgresModule) fetchAndUpdateInstallationHistory() {
	now := time.Now()
	interval := time.Hour * 24 * 30 // 1 month
	for installationId, meters := range mm.installations {
		startTime := mm.getLastHistoryTime(installationId)
		for startTime.Before(now) {
			endTime := startTime.Add(interval)
			mm.log.Info().Str("installation", installationId).Time("startTime", startTime).Time("endTime", endTime).Msg("Getting history")

			// TODO multiple call for multiple type
			data, err := mm.climkit.GetMeterData(installationId, meters, climkit.Electricity, startTime, endTime)
			if err != nil {
				log.Fatal().Str("installation", installationId).Time("startTime", startTime).Err(err).Msg("Unable to get data")
			}
			for _, instalData := range data {
				timestamp := instalData.Timestamp

				query := `INSERT INTO t_installation_values (installation_id, date_time, prod_total, self, to_ext)
							  VALUES ($1, $2, $3, $4, $5)
							  ON CONFLICT (installation_id, date_time)
							  DO UPDATE SET prod_total=$3, self=$4, to_ext=$5`
				err := mm.postgresClient.Execute(query,
					installationId, timestamp, instalData.ProdTotal, instalData.Self, instalData.ToExt)
				if err != nil {
					mm.log.Error().Str("installation", installationId).Time("Timestamp", timestamp).Err(err).Msg("Unable to insert meter data")
				}

				for _, meterData := range instalData.Meters {
					query := `INSERT INTO t_meter_values (meter_id, date_time, total, self, ext)
							  VALUES ($1, $2, $3, $4, $5)
							  ON CONFLICT (meter_id, date_time)
							  DO UPDATE SET total=$3, self=$4, ext=$5`
					err := mm.postgresClient.Execute(query,
						meterData.MeterId, timestamp, meterData.Total, meterData.Self, meterData.Ext)
					if err != nil {
						mm.log.Error().Str("installation", installationId).Time("Timestamp", timestamp).Err(err).Msg("Unable to insert meter data")
					}
				}
			}

			// sleep to avoid "too many requests"
			time.Sleep(2 * time.Second)

			startTime = endTime
		}
	}
}

func (mm *MeterPostgresModule) getLastHistoryTime(installationId string) time.Time {
	row := mm.postgresClient.Select(`SELECT date_time FROM t_installation_values WHERE installation_id=$1 ORDER BY date_time DESC LIMIT 1 `, installationId)
	var lastTime time.Time
	err := row.Scan(&lastTime)
	if err != nil {
		if err != sql.ErrNoRows {
			mm.log.Error().Err(err).Str("installationId", installationId).Msg("Unable to get last installation values")
		}

		//lastTime, _ = time.Parse(time.RFC3339, "2022-08-14T00:00:00Z")

		row = mm.postgresClient.Select(`SELECT creation_date FROM t_installations WHERE installation_id=$1 LIMIT 1 `, installationId)
		err = row.Scan(&lastTime)
		if err != nil {
			mm.log.Error().Err(err).Str("installationId", installationId).Msg("Unable to get installation creation date")
		}
	}

	return lastTime
}

func (mm *MeterPostgresModule) updateInstallation(installationId string, installation climkit.InstallationInfo) {
	query := `INSERT INTO t_installations(installation_id, site_ref, name, timezone, creation_date, latitude, longitude)
		VALUES($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (installation_id) DO UPDATE set site_ref=$2, name=$3, timezone=$4, creation_date=$5, latitude=$6, longitude=$7`

	err := mm.postgresClient.Execute(query, installationId, installation.SiteRef, installation.Name, installation.Timezone, installation.CreationDate, installation.Latitude, installation.Longitude)
	if err != nil {
		mm.log.Fatal().Err(err).Str("installationId", installationId).Msg("Unable to update installation")
	}
}

func (mm *MeterPostgresModule) updateMeterInfo(installationId string, meter climkit.MeterInfo) {
	query := `INSERT INTO t_meters(meter_id, installation_id, meter_type, prim_ad, virtual)
		VALUES($1, $2, $3, $4, $5)
		ON CONFLICT (meter_id) DO UPDATE set installation_id=$2, meter_type=$3, prim_ad=$4, virtual=$5`

	err := mm.postgresClient.Execute(query, meter.Id, installationId, meter.Type, meter.PrimAd, meter.Virtual)
	if err != nil {
		mm.log.Fatal().Err(err).Str("installationId", installationId).Str("MeterId", meter.Id).Msg("Unable to update meter")
	}
}
