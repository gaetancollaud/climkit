package climkit

import (
	"encoding/json"
	"github.com/rs/zerolog/log"
	"time"
)

type Climkit struct {
	api *ClimkitAPI
}

func New(apiUrl string, username string, password string) *Climkit {
	return &Climkit{
		api: NewApi(apiUrl, username, password),
	}
}

func (c *Climkit) GetAll() error {
	installations, err := c.api.GetInstallations()
	if err != nil {

		return err
	}
	log.Info().Strs("installations", installations).Msg("installations retrieved")

	for i := range installations {
		info, err := c.api.getInstallationInfo(installations[i])
		if err != nil {
			return err
		}
		infoStr, _ := json.Marshal(info)
		log.Info().RawJSON("info", infoStr).Msg("got installation info")

		meters, err := c.api.getMetersInfos(installations[i])
		metersStr, _ := json.Marshal(meters)
		log.Info().RawJSON("meters", metersStr).Msg("got installation meters")

		timeSeries, err := c.api.getMeterData(installations[i], meters, Electricity, time.Now().Add(-time.Minute*30))
		timeSeriesStr, _ := json.Marshal(timeSeries)
		log.Info().RawJSON("timeSeries", timeSeriesStr).Msg("got data")
	}
	return nil
}
