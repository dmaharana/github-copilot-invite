package main

import (
	"github-copilot-invite/internal/logger"
	"github-copilot-invite/internal/server"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func init() {
	// Load configuration
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal().Err(err).Msg("Error reading config file")
	}

	// Initialize logger
	logger.Init(viper.GetString("server.environment"))
	log.Info().Msg("Application starting...")
}

func main() {
	// Create and start server
	srv := server.New()
	if err := srv.Start(); err != nil {
		log.Fatal().Err(err).Msg("Server error")
	}
}
