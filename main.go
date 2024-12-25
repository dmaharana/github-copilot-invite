package main

import (
	"github-copilot-invite/internal/config"
	"github-copilot-invite/internal/handlers"
	"github-copilot-invite/internal/logger"
	"github-copilot-invite/internal/server"
	"github-copilot-invite/internal/updater"
	"github.com/spf13/viper"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

const configFile = "config.yaml"

func init() {
	// Initialize configuration manager
	configPath, err := filepath.Abs(configFile)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get config file path")
	}

	configMgr, err := config.NewManager(configPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create config manager")
	}

	// Load and process configuration
	if err := configMgr.Load(); err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Initialize logger
	logger.Init(configMgr.GetDecrypted("server.environment"))
	log.Info().Msg("Application starting...")
}

func main() {
	// create handler
	h := handlers.NewHandler(
		viper.GetString("github.token"),
		viper.GetString("smartsheet.token"),
		viper.GetInt64("smartsheet.sheet_id"),
	)

	// Create and start server
	updater.NewOrgsTrigger(h)

	// Create and start server
	srv := server.New()
	if err := srv.Start(); err != nil {
		log.Fatal().Err(err).Msg("Server error")
	}
}
