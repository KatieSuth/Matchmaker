package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func Auth(jwtSecret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		userId, response := validateAuth(jwtSecret, authHeader)
		if response != 0 {
			c.AbortWithStatus(response)
			return
		}

		// Attach the user ID to the context so handlers can access it
		c.Set("userID", userId)
		c.Next()
	}
}

func validateAuth(jwtSecret []byte, authHeader string) (string, int) {
	//verify the header isn't empty
	if authHeader == "" {
		return "", http.StatusUnauthorized
	}

	//verify header is in correct pattern (must be "Bearer <token>")
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", http.StatusUnauthorized
	}

	token, err := jwt.ParseWithClaims(parts[1], &model.Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return "", http.StatusUnauthorized
	}

	claims, ok := token.Claims.(*model.Claims)
	if !ok {
		return "", http.StatusUnauthorized
	}

	return claims.UserID, 0
}
