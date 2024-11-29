package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// BearerAuth middleware validates the bearer token in the Authorization header
func BearerAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			return
		}

		// Check if it's a Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format. Expected 'Bearer <token>'",
			})
			return
		}

		token := parts[1]
		expectedToken := viper.GetString("api.token")

		// Validate the token
		if expectedToken == "" {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "API token not configured",
			})
			return
		}

		if token != expectedToken {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			return
		}

		// Token is valid, proceed with the request
		c.Next()
	}
}
