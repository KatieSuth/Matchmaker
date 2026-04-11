package middleware

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func Auth(jwtSecret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		userId, response, err := ValidateAuth(jwtSecret, authHeader)
		if response != 0 || err != nil {
			slog.WarnContext(c.Request.Context(), "failed to validated JWT", "error", err)
			c.AbortWithStatus(response)
			return
		}

		// Attach the user ID to the context so handlers can access it
		slog.InfoContext(c.Request.Context(), "validated JWT successfully", "user_id", userId)
		c.Set("userID", userId)
		c.Next()
	}
}

func ValidateAuth(jwtSecret []byte, authHeader string) (string, int, error) {
	//verify the header isn't empty
	if authHeader == "" {
		return "", http.StatusUnauthorized, errors.New("empty auth header")
	}

	//verify header is in correct pattern (must be "Bearer <token>")
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", http.StatusUnauthorized, errors.New("invalid header pattern")
	}

	token, err := jwt.ParseWithClaims(parts[1], &model.Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return "", http.StatusUnauthorized, errors.New("unable to parse token")
	}

	claims, ok := token.Claims.(*model.Claims)
	if !ok {
		return "", http.StatusUnauthorized, errors.New("invalid token claims")
	}

	return claims.UserID, 0, nil
}
