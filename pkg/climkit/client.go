package climkit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const ClimkitTimeFormat = "2006-01-02 15:04:05"

type Client struct {
	options    ClientOptions
	httpClient *http.Client
	log        zerolog.Logger
}

type InstallationInfo struct {
	SiteRef      string `json:"site_ref"`
	Name         string `json:"name"`
	CreationDate string `json:"creation_date"`
	Address      struct {
		StreetName   string `json:"street_name"`
		StreetNumber string `json:"street_number"`
		CityName     string `json:"city_name"`
		CityRef      int16  `json:"city_ref"`
	} `json:"address"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timezone  string  `json:"timezone"`
}

type MeterInfo struct {
	Id      string `json:"id"`
	Type    string `json:"type"`
	PrimAd  int    `json:"prim_ad"`
	Virtual bool   `json:"virtual"`
}

type MeterDataItem struct {
	MeterId string
	Ext     float64
	Self    float64
	Total   float64
}

type MeterData struct {
	ConsoTotal              float64
	FromExt                 float64
	ProdTotal               float64
	Self                    float64
	StorageChargingTotal    float64
	StorageDischargingTotal float64
	ToExt                   float64
	Meters                  []MeterDataItem
	Timestamp               time.Time
}

type MeterType string

const (
	Electricity MeterType = "electricity"
	Heating               = "heating"
	ColdWater             = "cold_water"
	HotWater              = "hot_water"
	ChargePoint           = "charge_point"
)

type TimeSeriesRequest struct {
	// Should be ISO8601 but without timezone !
	StartTime string `json:"t_s"`
	EndTime   string `json:"t_e"`
}

func NewClient(options *ClientOptions) Client {
	logger := log.With().Str("Component", "Climkit").Logger()
	interceptor := NewInterceptor(logger, options)

	return Client{
		httpClient: &http.Client{
			Transport: interceptor,
		},
		options: *options,
		log:     logger,
	}
}

func (c *Client) GetInstallationIds() ([]string, error) {
	var obj []string
	err := c.get("installations", "v1/all_installations", &obj)
	return obj, err
}

func (c *Client) GetInstallationInfo(installationId string) (InstallationInfo, error) {
	var obj InstallationInfo
	err := c.get("installation info", "v1/installation_infos/"+installationId, &obj)
	return obj, err
}

func (c *Client) GetMetersInfos(installationId string) ([]MeterInfo, error) {
	var obj []MeterInfo
	err := c.get("meters info", "v1/meter_info/"+installationId, &obj)
	return obj, err
}

func (c *Client) GetMeterData(installationId string, meters []MeterInfo, meterType MeterType, startTime time.Time, endTime time.Time) ([]MeterData, error) {
	var obj []interface{}

	// implicit UTC
	formattedStartTime := startTime.UTC().Format(ClimkitTimeFormat)
	formattedEndTime := endTime.UTC().Format(ClimkitTimeFormat)

	request := TimeSeriesRequest{
		StartTime: formattedStartTime,
		EndTime:   formattedEndTime,
	}

	err := c.getHistory("meters data", "v1/site_data/"+installationId+"/"+string(meterType), request, &obj)

	var meterDataArray []MeterData

	for i := 0; i < len(obj); i++ {
		rawJson := obj[i].(map[string]interface{})

		var meterDataItemArray []MeterDataItem

		for j := 0; j < len(meters); j++ {
			meter := meters[j]
			item := MeterDataItem{
				MeterId: meter.Id,
			}
			if rawJson["ext_"+meter.Id] != nil {
				item.Ext = parse64AndLogError(rawJson["ext_"+meter.Id])
			}
			if rawJson["self_"+meter.Id] != nil {
				item.Ext = parse64AndLogError(rawJson["self_"+meter.Id])
			}
			if rawJson["total_"+meter.Id] != nil {
				item.Total = parse64AndLogError(rawJson["total_"+meter.Id])
			}
			meterDataItemArray = append(meterDataItemArray, item)
		}

		meterData := MeterData{
			ConsoTotal:              parse64AndLogError(rawJson["conso_total"]),
			FromExt:                 parse64AndLogError(rawJson["from_ext"]),
			ProdTotal:               parse64AndLogError(rawJson["prod_total"]),
			Self:                    parse64AndLogError(rawJson["self"]),
			StorageChargingTotal:    parse64AndLogError(rawJson["storage_charging_total"]),
			StorageDischargingTotal: parse64AndLogError(rawJson["storage_discharging_total"]),
			ToExt:                   parse64AndLogError(rawJson["to_ext"]),
			Timestamp:               parseTimeAndLogError(rawJson["timestamp"]),
			Meters:                  meterDataItemArray,
		}

		meterDataArray = append(meterDataArray, meterData)
	}

	return meterDataArray, err
}

func parse64AndLogError(input interface{}) float64 {
	//str := input.(string)
	//float, err := strconv.ParseFloat(str, 64)
	//if err != nil {
	//	log.Err(err).Str("str", str).Msg("Unable to parse float")
	//	return 0.0
	//}
	return input.(float64)
}

func parseTimeAndLogError(input interface{}) time.Time {
	str := input.(string)
	str = strings.Replace(str, " ", "T", 1) // fix timestamp format to ISO8601
	parsed, err := time.Parse(time.RFC3339, str)
	if err != nil {
		log.Err(err).Str("str", str).Msg("Unable to parse time")
	}
	return parsed
}

func (c *Client) get(methodName string, path string, returnObject any) error {
	c.log.Info().Str("methodName", methodName).Msg("Get request")
	resp, err := c.httpClient.Get(c.options.ApiUrl + path)
	return c.handleHttpResponse(methodName, resp, err, returnObject)
}

func (c *Client) getHistory(methodName string, path string, request TimeSeriesRequest, returnObject any) error {
	jsonRequest, err := json.Marshal(request)
	c.log.Info().Str("methodName", methodName).Str("request", string(jsonRequest)).Msg("Get history")
	if err != nil {
		return fmt.Errorf("cannot serialize request %s: %w", methodName, err)
	}
	resp, err := c.httpClient.Post(c.options.ApiUrl+path, "application/json", bytes.NewBuffer(jsonRequest))
	return c.handleHttpResponse(methodName, resp, err, returnObject)
}

func (c *Client) handleHttpResponse(methodName string, resp *http.Response, err error, returnObject any) error {
	if err != nil {
		return fmt.Errorf("unable to get %s: %w", methodName, err)
	}

	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		return fmt.Errorf("error reading the request for %s: %w", methodName, err)
	}
	err = json.Unmarshal(body, &returnObject)
	if err != nil {
		return fmt.Errorf("unable to unmarshal body when getting %s. Status: %s, body: %s, err=%w", methodName, resp.Status, string(body), err)
	} else if resp.StatusCode != 200 {
		return fmt.Errorf("error with request to get %s. Status: %s, body: %s", methodName, resp.Status, string(body))
	} else {
		return nil
	}
}
