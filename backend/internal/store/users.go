package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *PostgresStore) GetUserByDiscordID(ctx context.Context, discordId string, errorOnNoRows bool) (model.User, error) {
	dbUser, err := s.q.GetUserByDiscordID(ctx, &discordId)
	if err != nil {
		return model.User{}, fmt.Errorf("looking up user: %w", err)
	}

	if errorOnNoRows && errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, fmt.Errorf("looking up user: %w", err)
	}

	return model.MapDbUserToUser(dbUser), nil
}

func (s *PostgresStore) CreateNewUser(ctx context.Context, discordUser model.DiscordUser) (model.User, error) {
	dbUser, err := s.q.CreateUser(ctx, db.CreateUserParams{
		DiscordID:   &discordUser.ID,
		DiscordName: &discordUser.Username,
		ImageUrl:    &discordUser.Avatar,
	})
	if err != nil {
		return model.User{}, fmt.Errorf("creating user: %w", err)
	}
	return model.MapDbUserToUser(dbUser), nil
}

func (s *PostgresStore) UpdateUserFromLogin(ctx context.Context, userId uuid.UUID, discordUser model.DiscordUser) (model.User, error) {
	dbUser, err := s.q.UpdateUserFromLogin(ctx, db.UpdateUserFromLoginParams{
		DiscordName: &discordUser.Username,
		ImageUrl:    &discordUser.Avatar,
		ID:          userId,
	})
	if err != nil {
		return model.User{}, fmt.Errorf("updating user on login: %w", err)
	}
	return model.MapDbUserToUser(dbUser), nil
}
