package climkit

import (
	"time"
)

// ClientOptions contains configurable options for a DigitalStrom Client.
type ClientOptions struct {
	ApiUrl       string
	Username     string
	Password     string
	PollInterval time.Duration
}

func NewClientOptions() *ClientOptions {
	return &ClientOptions{
		ApiUrl:       "https://api.climkit.io/api/v1/",
		Username:     "",
		Password:     "",
		PollInterval: time.Minute * 5,
	}
}

func (o *ClientOptions) SetApiUrl(apiUrl string) *ClientOptions {
	o.ApiUrl = apiUrl
	return o
}

func (o *ClientOptions) SetUsername(u string) *ClientOptions {
	o.Username = u
	return o
}

func (o *ClientOptions) SetPassword(p string) *ClientOptions {
	o.Password = p
	return o
}

func (o *ClientOptions) SetPollInterval(pollInterval time.Duration) *ClientOptions {
	o.PollInterval = pollInterval
	return o
}
