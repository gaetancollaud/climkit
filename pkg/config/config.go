package config

import (
	"fmt"
	"github.com/spf13/viper"
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
type Config struct {
	Climkit  ConfigClimkit
	Mqtt     ConfigMqtt
	LogLevel string
}

const (
	undefined             string = "__undefined__"
	configFile            string = "config.yaml"
	envKeyClimkitApiUrl   string = "climkit.api-url"
	envKeyClimkitUsername string = "climkit.username"
	envKeyClimkitPassword string = "climkit.password"
	envKeyMqttUrl         string = "mqtt.url"
	envKeyMqttUsername    string = "mqtt.username"
	envKeyMqttPassword    string = "mqtt.password"
	envKeyMqttTopicPrefix string = "mqtt.topic-prefix"
	envKeyMqttRetain      string = "mqtt.retain"
	envKeyLogLevel        string = "log.level"
)

var defaultConfig = map[string]interface{}{
	envKeyClimkitApiUrl:   "https://api.climkit.io/api/",
	envKeyClimkitUsername: undefined,
	envKeyClimkitPassword: undefined,
	envKeyMqttUrl:         undefined,
	envKeyMqttUsername:    "",
	envKeyMqttPassword:    "",
	envKeyMqttTopicPrefix: "climkit",
	envKeyMqttRetain:      false,
	envKeyLogLevel:        "INFO",
}

// FromEnv returns a Config from env variables
func ReadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	// Set the current directory where the binary is being run.
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	for key, value := range defaultConfig {
		if value != undefined {
			viper.SetDefault(key, value)
		}
	}

	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("ReadInConfig error: %w", err)
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
		LogLevel: viper.GetString(envKeyLogLevel),
	}

	return config, nil
}

func (c *Config) String() string {
	return fmt.Sprintf("%+v\n", c.Climkit)
}
