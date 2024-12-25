package internal

import (
	"github-copilot-invite/internal/handlers"
	"github-copilot-invite/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all the routes for the application
func SetupRoutes(r *gin.Engine, h *handlers.Handler) {
	// Health check endpoint (unprotected)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
		})
	})

	// API routes (protected with bearer token)
	api := r.Group("/api/v1")
	api.Use(middleware.BearerAuth())
	{
		// GitHub Organization endpoints
		api.GET("/orgs", h.ListOrganizations)
		api.GET("/orgs/:org/teams", h.ListTeams)
		api.POST("/orgs/:org/teams", h.CreateTeam)

		// GitHub Copilot invite endpoint
		api.POST("/copilot/invite", h.SendCopilotInvite)

		// Health check endpoint (unprotected)
		r.GET("/healthcheck", h.HealthCheck)
	}
}
