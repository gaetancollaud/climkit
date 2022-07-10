package main

import (
	"github.com/gaetancollaud/climkit-to-mqtt/pkg/climkit"
	"github.com/gaetancollaud/climkit-to-mqtt/pkg/config"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	config, err := config.ReadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Error found when reading the config.")
	}

	if config.LogLevel == "TRACE" {
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	} else if config.LogLevel == "DEBUG" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else if config.LogLevel == "INFO" {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	} else if config.LogLevel == "WARN" {
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	} else if config.LogLevel == "ERROR" {
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	}

	log.Info().Msg("Starting climkit to MQTT!")

	climkit := climkit.New(config.Climkit.ApiUrl, config.Climkit.Username, config.Climkit.Password)

	err = climkit.GetAll()
	if err != nil {
		log.Err(err).Msg("Unable to get installations")
	}

	//mqtt := climkit_mqtt.New(config, ds)

	//mqtt.Start()

	// Subscribe for interruption happening during execution.
	exitSignal := make(chan os.Signal)
	signal.Notify(exitSignal, os.Interrupt, syscall.SIGTERM)
	<-exitSignal

	// Graceful stop the connections.
	//mqtt.Stop()
}
