package model

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// generate auth token and refresh tokens
func GenerateTokens(userID string, jwtSecret []byte) (accessToken, refreshToken string, err error) {
	if userID == "" {
		err = errors.New("userID must not be empty")
		return
	}
	//access token
	accessClaims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	accessToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(jwtSecret)
	if err != nil {
		return
	}

	//refresh token
	key := make([]byte, 32)
	_, err = rand.Read(key)
	if err != nil {
		return
	}
	refreshToken = hex.EncodeToString(key)

	return
}
