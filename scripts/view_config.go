package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github-copilot-invite/internal/encryption"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

// Config represents the configuration structure
type Config struct {
	Github struct {
		Token string `yaml:"token"`
	} `yaml:"github"`
	Smartsheet struct {
		Token   string `yaml:"token"`
		SheetID int64  `yaml:"sheet_id"`
	} `yaml:"smartsheet"`
	API struct {
		Token string `yaml:"token"`
	} `yaml:"api"`
	Server struct {
		Port        string `yaml:"port"`
		Environment string `yaml:"environment"`
		SSL         struct {
			Enabled  bool   `yaml:"enabled"`
			CertFile string `yaml:"cert_file"`
			KeyFile  string `yaml:"key_file"`
		} `yaml:"ssl"`
	} `yaml:"server"`
}

func main() {
	// Parse flags
	showAll := flag.Bool("all", false, "Show all configuration values")
	showSensitive := flag.Bool("sensitive", false, "Show only sensitive values")
	showDecrypted := flag.Bool("decrypt", false, "Show decrypted values (use with caution)")
	configFile := flag.String("config", "../config.yaml", "Path to config file")
	flag.Parse()

	// Initialize logger
	log.Logger = zerolog.New(os.Stderr).With().Timestamp().Logger()

	// Get absolute path for config file
	absConfigFile, err := filepath.Abs(*configFile)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get absolute path for config file")
	}

	// Read config file
	data, err := os.ReadFile(absConfigFile)
	if err != nil {
		log.Fatal().Err(err).Str("path", absConfigFile).Msg("Failed to read config file")
	}

	// Parse config
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		log.Fatal().Err(err).Msg("Failed to parse config file")
	}

	// Initialize encryption manager
	encryptionMgr, err := encryption.NewManager()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize encryption manager")
	}

	// Print configuration
	fmt.Println("\nConfiguration Values:")
	fmt.Println("====================")

	if *showAll || *showSensitive {
		// Show sensitive values
		printValue("GitHub Token", config.Github.Token, encryptionMgr, true, *showDecrypted)
		printValue("Smartsheet Token", config.Smartsheet.Token, encryptionMgr, true, *showDecrypted)
		printValue("API Token", config.API.Token, encryptionMgr, true, *showDecrypted)
	}

	if *showAll {
		// Show non-sensitive values
		fmt.Printf("\nServer Configuration:\n")
		fmt.Printf("  Port: %s\n", config.Server.Port)
		fmt.Printf("  Environment: %s\n", config.Server.Environment)
		fmt.Printf("  SSL:\n")
		fmt.Printf("    Enabled: %v\n", config.Server.SSL.Enabled)
		fmt.Printf("    Cert File: %s\n", config.Server.SSL.CertFile)
		fmt.Printf("    Key File: %s\n", config.Server.SSL.KeyFile)
		fmt.Printf("\nSmartsheet Configuration:\n")
		fmt.Printf("  Sheet ID: %d\n", config.Smartsheet.SheetID)
	}

	if *showDecrypted {
		fmt.Printf("\nCAUTION: Showing decrypted values. Handle this information securely!\n")
	}
}

func printValue(name, value string, encryptionMgr *encryption.Manager, sensitive bool, showDecrypted bool) {
	if value == "" {
		fmt.Printf("%s: <not set>\n", name)
		return
	}

	if encryption.IsEncrypted(value) {
		decrypted, err := encryptionMgr.Decrypt(value)
		if err != nil {
			fmt.Printf("%s: <error decrypting: %v>\n", name, err)
			return
		}
		if sensitive && !showDecrypted {
			// Show only first and last 4 characters of sensitive values
			masked := maskString(decrypted)
			fmt.Printf("%s: %s (encrypted)\n", name, masked)
		} else {
			fmt.Printf("%s: %s (encrypted)\n", name, decrypted)
		}
	} else {
		if sensitive && !showDecrypted {
			masked := maskString(value)
			fmt.Printf("%s: %s (plaintext)\n", name, masked)
		} else {
			fmt.Printf("%s: %s (plaintext)\n", name, value)
		}
	}
}

func maskString(s string) string {
	if len(s) <= 8 {
		return strings.Repeat("*", len(s))
	}
	return s[:4] + strings.Repeat("*", len(s)-8) + s[len(s)-4:]
}
