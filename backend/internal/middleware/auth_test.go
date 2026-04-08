package middleware_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/KatieSuth/MatchmakerAPI/internal/middleware"
	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/KatieSuth/MatchmakerAPI/internal/testutil"
	"github.com/golang-jwt/jwt/v5"
)

func TestValidateAuth(t *testing.T) {
	jwtSecret, err := testutil.GetJWTSecret(t)
	if err != nil {
		t.Error(err)
	}

	//make sure it fails on empty
	_, numStatus := middleware.ValidateAuth(jwtSecret, "")
	if numStatus != http.StatusUnauthorized {
		t.Errorf("Token is empty; 401 expected if no Bearer token is provided")
	}

	testutil.WithTestTx(t, func(q *db.Queries, s *store.PostgresStore) {
		//create a user for the token
		discordId := "discordID1234"
		discordName := "discordName1234"

		dbUser, err := q.CreateUser(t.Context(), db.CreateUserParams{
			DiscordID:   &discordId,
			DiscordName: &discordName,
		})

		if err != nil {
			t.Errorf("Could not create user: %v", err)
			return
		}

		//create a token
		accessToken, _, err := model.GenerateTokens(dbUser.ID.String(), jwtSecret)
		if err != nil {
			t.Errorf("Could not create tokens: %v", err)
			return
		}

		//try to use the token but without proper format
		_, numStatus := middleware.ValidateAuth(jwtSecret, accessToken)
		if numStatus != http.StatusUnauthorized {
			t.Errorf("Token was provided incorrectly; 401 expected if not 'Bearer <token>'; got %d", numStatus)
			return
		}

		//try to use the token with invalid secret
		_, numStatus = middleware.ValidateAuth([]byte{}, fmt.Sprintf("Bearer %s", accessToken))
		if numStatus != http.StatusUnauthorized {
			t.Errorf("Token secret was incorrect, 401 expected; got %d", numStatus)
			return
		}

		//make sure user matches
		userId, numStatus := middleware.ValidateAuth(jwtSecret, fmt.Sprintf("Bearer %s", accessToken))
		if userId != dbUser.ID.String() || numStatus != 0 {
			t.Errorf("User should have matched what was set in the token (%s); got %s with status %d", dbUser.ID.String(), userId, numStatus)
			return
		}

		//create expired token
		accessClaims := model.Claims{
			UserID: dbUser.ID.String(),
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-5 * time.Minute)),
				IssuedAt:  jwt.NewNumericDate(time.Now().Add(-15 * time.Minute)),
			},
		}

		accessToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(jwtSecret)
		if err != nil {
			t.Errorf("could not create token with expired claims")
			return
		}

		_, numStatus = middleware.ValidateAuth(jwtSecret, fmt.Sprintf("Bearer %s", accessToken))
		if numStatus != http.StatusUnauthorized {
			t.Errorf("Token is expired, 401 expected; got %d", numStatus)
			return
		}
	})
}
