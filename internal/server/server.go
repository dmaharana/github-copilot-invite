package server

import (
	"fmt"

	"github-copilot-invite/internal"
	"github-copilot-invite/internal/handlers"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// Server represents the HTTP server
type Server struct {
	config  *Config
	router  *gin.Engine
	handler *handlers.Handler
}

// New creates a new server instance
func New() *Server {
	log.Debug().Msg("Initializing server...")

	// Initialize configuration
	config := NewConfig()

	// Set Gin mode based on environment
	if config.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize handler
	handler := handlers.NewHandler(
		viper.GetString("github.token"),
		viper.GetString("smartsheet.token"),
		viper.GetInt64("smartsheet.sheet_id"),
	)

	log.Debug().Msg("Handler initialized")

	// Initialize router
	router := gin.Default()

	// Setup routes
	internal.SetupRoutes(router, handler)

	log.Debug().Msg("Routes configured")

	return &Server{
		config:  config,
		router:  router,
		handler: handler,
	}
}

// Start starts the server
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%s", s.config.Port)

	// Validate SSL configuration
	if err := s.config.ValidateSSL(); err != nil {
		log.Warn().
			Err(err).
			Msg("SSL validation failed, falling back to HTTP")
		s.config.SSL.Enabled = false
	}

	// Start server with appropriate protocol
	if s.config.SSL.Enabled {
		log.Info().
			Str("address", addr).
			Msg("Starting HTTPS server")
		return s.router.RunTLS(addr, s.config.SSL.CertFile, s.config.SSL.KeyFile)
	}

	log.Info().
		Str("address", addr).
		Msg("Starting HTTP server")
	return s.router.Run(addr)
}
