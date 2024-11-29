package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Init initializes the global logger with the given configuration
func Init(environment string) {
	// Set up the logger output
	var output io.Writer = os.Stdout
	if environment == "production" {
		// In production, write to both file and stdout
		logFile, err := os.OpenFile(
			"app.log",
			os.O_APPEND|os.O_CREATE|os.O_WRONLY,
			0664,
		)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to open log file")
		}
		output = zerolog.MultiLevelWriter(os.Stdout, logFile)
	}

	// Configure zerolog
	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if environment != "production" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// Create a console writer for development
	if environment != "production" {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	// Set global logger
	log.Logger = zerolog.New(output).With().Timestamp().Caller().Logger()
}

// Logger returns the global logger instance
func Logger() *zerolog.Logger {
	return &log.Logger
}
