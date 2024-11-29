package config

import (
	"github-copilot-invite/internal/encryption"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Manager handles configuration with encryption support
type Manager struct {
	encryptionMgr *encryption.Manager
	configFile    string
}

// NewManager creates a new configuration manager
func NewManager(configFile string) (*Manager, error) {
	encryptionMgr, err := encryption.NewManager()
	if err != nil {
		return nil, err
	}

	return &Manager{
		encryptionMgr: encryptionMgr,
		configFile:    configFile,
	}, nil
}

// Load loads and processes the configuration file
func (m *Manager) Load() error {
	// Read the configuration file
	viper.SetConfigFile(m.configFile)
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	// Get all settings
	settings := viper.AllSettings()

	// Process sensitive fields
	modified := false
	sensitiveKeys := []string{
		"github.token",
		"smartsheet.token",
		"api.token",
	}

	// Process each sensitive key
	for _, key := range sensitiveKeys {
		parts := strings.Split(key, ".")
		if value := getNestedValue(settings, parts...); value != "" {
			strValue := value.(string)
			if !encryption.IsEncrypted(strValue) {
				// Encrypt the value
				encrypted, err := m.encryptionMgr.Encrypt(strValue)
				if err != nil {
					log.Error().Err(err).Str("key", key).Msg("Failed to encrypt value")
					continue
				}
				setNestedValue(settings, encrypted, parts...)
				modified = true
				log.Info().Str("key", key).Msg("Encrypted sensitive value")
			}
		}
	}

	// If any values were encrypted, update the config file
	if modified {
		// Marshal the updated configuration
		yamlData, err := yaml.Marshal(settings)
		if err != nil {
			return err
		}

		// Write the updated configuration back to the file
		if err := os.WriteFile(m.configFile, yamlData, 0644); err != nil {
			return err
		}

		// Reload the configuration
		if err := viper.ReadInConfig(); err != nil {
			return err
		}

		log.Info().Msg("Updated configuration file with encrypted values")
	}

	return nil
}

// GetDecrypted gets a decrypted configuration value
func (m *Manager) GetDecrypted(key string) string {
	value := viper.GetString(key)
	if value == "" || !encryption.IsEncrypted(value) {
		return value
	}

	decrypted, err := m.encryptionMgr.Decrypt(value)
	if err != nil {
		log.Error().Err(err).Str("key", key).Msg("Failed to decrypt value")
		return ""
	}

	return decrypted
}

// Helper function to get nested map value
func getNestedValue(m map[string]interface{}, keys ...string) interface{} {
	current := m
	for i, key := range keys {
		if i == len(keys)-1 {
			return current[key]
		}
		if current[key] == nil {
			return nil
		}
		current = current[key].(map[string]interface{})
	}
	return nil
}

// Helper function to set nested map value
func setNestedValue(m map[string]interface{}, value string, keys ...string) {
	current := m
	for i, key := range keys {
		if i == len(keys)-1 {
			current[key] = value
			return
		}
		if current[key] == nil {
			current[key] = make(map[string]interface{})
		}
		current = current[key].(map[string]interface{})
	}
}
