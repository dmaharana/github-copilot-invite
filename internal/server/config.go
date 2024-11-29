package server

import (
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// Config holds all server configuration
type Config struct {
	Port        string
	Environment string
	SSL         SSLConfig
}

// SSLConfig holds SSL-specific configuration
type SSLConfig struct {
	Enabled  bool
	CertFile string
	KeyFile  string
}

// NewConfig creates a new server configuration from viper settings
func NewConfig() *Config {
	port := viper.GetString("server.port")
	if port == "" {
		port = "8080"
		log.Debug().Msg("No port configured, using default: 8080")
	}

	config := &Config{
		Port:        port,
		Environment: viper.GetString("server.environment"),
		SSL: SSLConfig{
			Enabled:  viper.GetBool("server.ssl.enabled"),
			CertFile: viper.GetString("server.ssl.cert_file"),
			KeyFile:  viper.GetString("server.ssl.key_file"),
		},
	}

	log.Info().
		Str("port", config.Port).
		Str("environment", config.Environment).
		Bool("ssl_enabled", config.SSL.Enabled).
		Msg("Server configuration loaded")

	return config
}

// ValidateSSL checks if SSL certificates exist and are accessible
func (c *Config) ValidateSSL() error {
	if !c.SSL.Enabled {
		log.Debug().Msg("SSL is disabled, skipping certificate validation")
		return nil
	}

	log.Debug().
		Str("cert_file", c.SSL.CertFile).
		Str("key_file", c.SSL.KeyFile).
		Msg("Validating SSL certificates")

	// Check certificate file
	if _, err := os.Stat(c.SSL.CertFile); os.IsNotExist(err) {
		log.Error().
			Str("cert_file", c.SSL.CertFile).
			Msg("SSL certificate file not found")
		return err
	}

	// Check key file
	if _, err := os.Stat(c.SSL.KeyFile); os.IsNotExist(err) {
		log.Error().
			Str("key_file", c.SSL.KeyFile).
			Msg("SSL key file not found")
		return err
	}

	// Ensure certificate directory exists
	certDir := filepath.Dir(c.SSL.CertFile)
	if err := os.MkdirAll(certDir, 0755); err != nil {
		log.Error().
			Str("directory", certDir).
			Err(err).
			Msg("Failed to create certificate directory")
		return err
	}

	log.Info().Msg("SSL certificate validation successful")
	return nil
}
