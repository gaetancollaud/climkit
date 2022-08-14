package config

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"strings"
)

type Mode string

const (
	Mqtt     Mode = "mqtt"
	Postgres      = "postgres"
)

type ConfigClimkit struct {
	ApiUrl   string
	Username string
	Password string
}
type ConfigMqtt struct {
	MqttUrl     string
	Username    string
	Password    string
	TopicPrefix string
	Retain      bool
}
type ConfigPostgres struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
	SslMode  string
}
type Config struct {
	Climkit  ConfigClimkit
	Mqtt     ConfigMqtt
	Postgres ConfigPostgres
	Mode     Mode
	LogLevel string
}

const (
	undefined              string = "__undefined__"
	configFile             string = "config.yaml"
	envKeyMode             string = "mode"
	envKeyLogLevel         string = "log.level"
	envKeyClimkitApiUrl    string = "climkit.api-url"
	envKeyClimkitUsername  string = "climkit.username"
	envKeyClimkitPassword  string = "climkit.password"
	envKeyMqttUrl          string = "mqtt.url"
	envKeyMqttUsername     string = "mqtt.username"
	envKeyMqttPassword     string = "mqtt.password"
	envKeyMqttTopicPrefix  string = "mqtt.topic-prefix"
	envKeyMqttRetain       string = "mqtt.retain"
	envKeyPostgresHost     string = "postgres.host"
	envKeyPostgresPort     string = "postgres.port"
	envKeyPostgresDatabase string = "postgres.database"
	envKeyPostgresUsername string = "postgres.username"
	envKeyPostgresPassword string = "postgres.password"
	envKeyPostgresSslMode  string = "postgres.ssl-mode"
)

var defaultConfig = map[string]interface{}{
	envKeyMode:             undefined,
	envKeyClimkitApiUrl:    "https://api.climkit.io/api/",
	envKeyClimkitUsername:  undefined,
	envKeyClimkitPassword:  undefined,
	envKeyMqttUrl:          "",
	envKeyMqttUsername:     "",
	envKeyMqttPassword:     "",
	envKeyMqttTopicPrefix:  "climkit",
	envKeyMqttRetain:       false,
	envKeyLogLevel:         "INFO",
	envKeyPostgresHost:     "localhost",
	envKeyPostgresPort:     "5432",
	envKeyPostgresDatabase: "postgres",
	envKeyPostgresUsername: "postgres",
	envKeyPostgresPassword: "postgres",
	envKeyPostgresSslMode:  "disable",
}

// FromEnv returns a Config from env variables
func ReadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	// Set the current directory where the binary is being run.
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	for key, value := range defaultConfig {
		if value != undefined {
			viper.SetDefault(key, value)
		}
	}

	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			log.Info().Msg("No config file found, using default and environment variable")
		} else {
			return nil, fmt.Errorf("ReadInConfig error: %w", err)
		}
	}

	// Check for deprecated and undefined fields.
	for fieldName, defaultValue := range defaultConfig {
		if defaultValue == undefined && !viper.IsSet(fieldName) {
			return nil, fmt.Errorf("required field not found in config: %s", fieldName)
		}
	}

	config := &Config{
		Climkit: ConfigClimkit{
			ApiUrl:   viper.GetString(envKeyClimkitApiUrl),
			Username: viper.GetString(envKeyClimkitUsername),
			Password: viper.GetString(envKeyClimkitPassword),
		},
		Mqtt: ConfigMqtt{
			MqttUrl:     viper.GetString(envKeyMqttUrl),
			Username:    viper.GetString(envKeyMqttUsername),
			Password:    viper.GetString(envKeyMqttPassword),
			TopicPrefix: viper.GetString(envKeyMqttTopicPrefix),
			Retain:      viper.GetBool(envKeyMqttRetain),
		},
		Postgres: ConfigPostgres{
			Host:     viper.GetString(envKeyPostgresHost),
			Port:     viper.GetInt(envKeyPostgresPort),
			Database: viper.GetString(envKeyPostgresDatabase),
			Username: viper.GetString(envKeyPostgresUsername),
			Password: viper.GetString(envKeyPostgresPassword),
			SslMode:  viper.GetString(envKeyPostgresSslMode),
		},
		Mode:     Mode(viper.GetString(envKeyMode)),
		LogLevel: viper.GetString(envKeyLogLevel),
	}

	return config, nil
}

func (c *Config) String() string {
	return fmt.Sprintf("%+v\n", c.Climkit)
}
