package middleware

import (
	"strings"

	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/services"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(tokenService services.ITokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"message": "Authorization header is missing"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(401, gin.H{"message": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := tokenService.ExtractClaims(tokenString)
		if err != nil {
			c.JSON(401, gin.H{"message": "Invalid or expired token"})
			c.Abort()
			return
		}

		userID, _ := claims["user_id"].(string)
		username, _ := claims["username"].(string)

		c.Set("user_id", userID)
		c.Set("username", username)
		c.Next()
	}
}
