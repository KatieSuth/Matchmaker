package handler_test

import (
	"io"
	"log"
	"os"
	"testing"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/KatieSuth/MatchmakerAPI/internal/handler"
	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/KatieSuth/MatchmakerAPI/internal/testutil"
	"github.com/google/uuid"
)

func TestGetUser(t *testing.T) {
	log.SetOutput(io.Discard)
	defer log.SetOutput(os.Stderr)

	testutil.WithTestTx(t, func(q *db.Queries, s *store.PostgresStore) {
		//try getting user with invalid ID
		user, err := handler.GetUserById("123456789", s, t.Context())
		if err == nil {
			t.Errorf("Looked for with invalid ID (123456789), expected error. Got %v", user)
			return
		}

		//try getting user that doesn't exist
		userID := uuid.New()
		user, err = handler.GetUserById(userID.String(), s, t.Context())
		if err == nil {
			t.Errorf("Looked for user that didn't exist (%s), expected error. Got %v", userID.String(), user)
			return
		}

		//add a user
		discordId := "discordID1234"
		discordName := "discordName1234"
		avatar := "discordAvatar1234"

		dbUser, err := q.CreateUser(t.Context(), db.CreateUserParams{
			DiscordID:   &discordId,
			DiscordName: &discordName,
			ImageUrl:    &avatar,
		})

		//get the user
		user, err = handler.GetUserById(dbUser.ID.String(), s, t.Context())
		if err != nil {
			t.Errorf("Error fetching user (%s): %v", dbUser.ID.String(), err)
			return
		}

		//make sure user is correct
		if user.ID != dbUser.ID || *user.DiscordID != discordId || *user.DiscordName != discordName || *user.ImageUrl != avatar {
			t.Errorf("User returned doesn't match expected. ID (e: %s r: %s) | DiscordID (e: %s r: %s) | DiscordName (e: %s r: %s) | Avatar (e: %s, r: %s)",
				dbUser.ID.String(), user.ID.String(),
				discordId, *user.DiscordID,
				discordName, *user.DiscordName,
				avatar, *user.ImageUrl,
			)
			return
		}
	})
}
