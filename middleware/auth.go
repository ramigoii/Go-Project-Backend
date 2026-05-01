package middleware

import (
	"MovieDatabase/services"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

var authClient *services.AuthClient

func InitAuthMiddleware() {
	authClient = services.NewAuthClient()
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			c.Abort()
			return
		}

		tokenString := bearerToken[1]

		// Валидация токена через auth service
		validation, err := authClient.ValidateToken(tokenString)
		if err != nil || !validation.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Добавляем user_id в контекст
		c.Set("user_id", validation.UserID)
		c.Set("username", validation.Username)
		c.Next()
	}
}
