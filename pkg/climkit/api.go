package climkit

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"net/http"
)

type ClimkitAPI struct {
	apiUrl     string
	httpClient *http.Client
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

func NewApi(apiUrl string, username string, password string) *ClimkitAPI {
	interceptor := NewInterceptor(apiUrl, username, password)

	return &ClimkitAPI{
		httpClient: &http.Client{
			Transport: interceptor,
		},
		apiUrl: apiUrl,
	}
}

func (c *ClimkitAPI) GetInstallations() ([]string, error) {
	var obj []string
	err := c.get("installations", "v1/all_installations", &obj)
	return obj, err
}

func (c *ClimkitAPI) getInstallationInfo(installationId string) (InstallationInfo, error) {
	var obj InstallationInfo
	err := c.get("installation info", "v1/installation_infos/"+installationId, &obj)
	return obj, err
}

func (c *ClimkitAPI) getMetersInfos(installationId string) ([]MeterInfo, error) {
	var obj []MeterInfo
	err := c.get("meters info", "v1/meter_info/"+installationId, &obj)
	return obj, err
}

func (c *ClimkitAPI) get(methodName string, path string, returnObject any) error {
	log.Info().Msg("Getting " + methodName)
	resp, err := c.httpClient.Get(c.apiUrl + path)
	if err != nil {
		return fmt.Errorf("unable to get %s: %w", methodName, err)
	}

	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		return fmt.Errorf("error reading the request for %s: %w", methodName, err)
	}
	//var jsonResponse []MeterInfo
	err = json.Unmarshal(body, &returnObject)
	if err != nil {
		return fmt.Errorf("unable to unmarshal body when getting %s. Status: %s, body: %s, err=%w", methodName, resp.Status, string(body), err)
	} else if resp.StatusCode != 200 {
		return fmt.Errorf("error with request to get %s. Status: %s, body: %s", methodName, resp.Status, string(body))
	} else {
		return nil
	}
}
