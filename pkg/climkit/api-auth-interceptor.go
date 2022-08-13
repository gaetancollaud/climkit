package climkit

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog"
	"io/ioutil"
	"net/http"
	"time"
)

type Interceptor struct {
	core        http.RoundTripper
	options     ClientOptions
	accessToken string
	validUntil  time.Time
	log         zerolog.Logger
}

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	AccessToken string `json:"access_token"`
	ValidUntil  struct {
		Date int64 `json:"$date"`
	} `json:"valid_until"`
}

func NewInterceptor(logger zerolog.Logger, options *ClientOptions) *Interceptor {
	return &Interceptor{
		core: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			}},
		options:     *options,
		accessToken: "",
		validUntil:  time.Now().Add(-time.Hour),
		log:         logger,
	}
}

func (i *Interceptor) refreshTokenIfNecessary() (string, error) {
	if time.Now().Add(-10 * time.Second).After(i.validUntil) {
		i.log.Info().Msg("Token is expired, renewing")

		requestBody := AuthRequest{
			Username: i.options.Username,
			Password: i.options.Password,
		}
		jsonRequest, _ := json.Marshal(requestBody)
		response, err := http.Post(i.options.ApiUrl+"v1/auth", "application/json", bytes.NewBuffer(jsonRequest))
		if err != nil {
			return "", fmt.Errorf("unable to get accessToken: %w", err)
		}

		body, readErr := ioutil.ReadAll(response.Body)
		if readErr != nil {
			return "", fmt.Errorf("error reading the response: %w", err)
		}

		var jsonResponse AuthResponse
		json.Unmarshal(body, &jsonResponse)

		i.accessToken = jsonResponse.AccessToken
		i.validUntil = time.UnixMilli(jsonResponse.ValidUntil.Date)

		i.log.Debug().Str("access_token", i.accessToken).Time("valid_until", i.validUntil).Msg("Token received")
	}
	return i.accessToken, nil
}

func (i *Interceptor) modifyRequest(r *http.Request) *http.Request {
	token, err := i.refreshTokenIfNecessary()
	if err != nil {
		i.log.Err(err).Msg("Unable to get accessToken ")
	}
	i.log.Trace().Str("token", token).Msg("Injecting accessToken")
	r.Header.Set("Authorization", "Bearer "+token)
	return r
}

func (i *Interceptor) RoundTrip(r *http.Request) (*http.Response, error) {
	defer func() {
		if r.Body != nil {
			_ = r.Body.Close()
		}
	}()

	// modify before the request is sent
	newReq := i.modifyRequest(r)

	// send the request using the DefaultTransport
	return i.core.RoundTrip(newReq)
}
