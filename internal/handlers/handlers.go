package handlers

import (
	"net/http"

	"github-copilot-invite/internal/github"
	"github-copilot-invite/internal/smartsheet"

	"github.com/gin-gonic/gin"
	gh "github.com/google/go-github/v60/github"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	githubClient *github.Client
	validator    *smartsheet.LicenseValidator
}

func NewHandler(githubToken string, smartsheetToken string, sheetID int64) *Handler {
	return &Handler{
		githubClient: github.NewClient(githubToken),
		validator:    smartsheet.NewLicenseValidator(smartsheetToken, sheetID),
	}
}

func (h *Handler) ListOrganizations(c *gin.Context) {
	orgs, err := h.githubClient.ListOrganizations()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, orgs)
}

func (h *Handler) ListTeams(c *gin.Context) {
	org := c.Param("org")
	if org == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "organization name is required"})
		return
	}

	teams, err := h.githubClient.ListTeams(org)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, teams)
}

func (h *Handler) CreateTeam(c *gin.Context) {
	org := c.Param("org")
	if org == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "organization name is required"})
		return
	}

	var newTeam gh.NewTeam
	if err := c.BindJSON(&newTeam); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	team, err := h.githubClient.CreateTeam(org, &newTeam)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, team)
}

type CopilotInviteRequest struct {
	Organization string `json:"organization" binding:"required"`
	Team         string `json:"team" binding:"required"`
	Username     string `json:"username" binding:"required"`
}

func (h *Handler) SendCopilotInvite(c *gin.Context) {
	var req CopilotInviteRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate request
	if req.Organization == "" || req.Team == "" || req.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "organization, team and username are required"})
		return
	}

	// validate if Organization exists
	// orgs, err := h.githubClient.ListOrganizations()
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// 	return
	// }
	// var exists bool
	// var orgID int64
	// for _, org := range orgs {
	// 	if org.Name == &req.Organization {
	// 		exists = true
	// 		orgID = *org.ID
	// 		break
	// 	}
	// }
	// if !exists {
	// 	c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
	// 	return
	// }

	// validate if Team exists
	teams, err := h.githubClient.ListTeams(req.Organization)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var teamExists bool
	for _, t := range teams {
		if t.GetName() == req.Team {
			teamExists = true
			break
		}
	}
	if !teamExists {
		// Create a new team
		newTeam := gh.NewTeam{
			Name: req.Team,
		}
		_, err := h.githubClient.CreateTeam(req.Organization, &newTeam)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		log.Info().Str("organization", req.Organization).Str("team", req.Team).Msg("Created new team")
	}

	// Check license availability
	available, err := h.validator.CheckLicenseAvailability(req.Organization)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check license availability"})
		return
	}

	if !available {
		c.JSON(http.StatusConflict, gin.H{"error": "no licenses available for this organization"})
		return
	}

	// Send invite
	if err := h.githubClient.SendCopilotInvite(req.Organization, req.Team, req.Username); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Decrement license count
	if err := h.validator.DecrementLicense(req.Organization); err != nil {
		// Note: We might want to roll back the invite if this fails
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update license count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "invite sent successfully"})
}

// create a dummy endpoint for health check
func (h *Handler) HealthCheck(c *gin.Context) {
	log.Info().Msg("Health check")

	// print c
	log.Debug().Interface("c", c).Msg("c")
	if c.Request == nil {
		log.Debug().Msg("c.Request is nil")
		return
	} else {
		log.Debug().Msg("c.Request is not nil")
	}

	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}
