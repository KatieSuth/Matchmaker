-- +goose Up
CREATE TABLE "users" (
  "id" UUID PRIMARY KEY,
  "discord_id" TEXT UNIQUE,
  "discord_name" TEXT UNIQUE,
  "image_url" TEXT,
  "pronouns" TEXT,
  "show_pronouns" BOOLEAN NOT NULL DEFAULT true,
  "created_at" TIMESTAMPTZ NOT NULL,
  "updated_at" TIMESTAMPTZ NOT NULL
);

CREATE TABLE "api_links" (
  "id" UUID PRIMARY KEY,
  "user_id" UUID,
  "name" text,
  "refresh_token" text,
  "refresh_token_iv" text,
  "created_at" TIMESTAMPTZ NOT NULL,
  "updated_at" TIMESTAMPTZ NOT NULL
);

CREATE TABLE "refresh_tokens" (
  "token" TEXT PRIMARY KEY,
  "user_id" UUID NOT NULL,
  "expires_at" TIMESTAMPTZ NOT NULL,
  "revoked_at" timestamp,
  "created_at" TIMESTAMPTZ NOT NULL,
  "updated_at" TIMESTAMPTZ NOT NULL
);

CREATE TABLE "games" (
  "id" UUID PRIMARY KEY,
  "name" TEXT NOT NULL,
  "owner_id" UUID,
  "created_at" TIMESTAMPTZ NOT NULL,
  "updated_at" TIMESTAMPTZ NOT NULL
);

CREATE TABLE "game_modes" (
  "id" UUID PRIMARY KEY,
  "game_id" UUID NOT NULL,
  "name" TEXT NOT NULL,
  "team_size" INT NOT NULL DEFAULT 1,
  "owner_id" UUID,
  "duration" INT NOT NULL DEFAULT 0,
  "created_at" TIMESTAMPTZ NOT NULL,
  "updated_at" TIMESTAMPTZ NOT NULL
);

CREATE TABLE "game_ranks" (
  "id" UUID PRIMARY KEY,
  "game_id" UUID,
  "name" TEXT NOT NULL,
  "order" INT NOT NULL,
  "created_at" TIMESTAMPTZ NOT NULL,
  "updated_at" TIMESTAMPTZ NOT NULL
);

CREATE TABLE "user_games" (
  "user_id" UUID,
  "game_id" UUID,
  "in_game_name" TEXT NOT NULL,
  "current_rank" UUID,
  "peak_rank" UUID,
  "show_rank" BOOLEAN NOT NULL DEFAULT false,
  "api_permission" BOOLEAN NOT NULL DEFAULT false,
  "api_links_id" UUID,
  "created_at" TIMESTAMPTZ NOT NULL,
  "updated_at" TIMESTAMPTZ NOT NULL,
  PRIMARY KEY ("user_id", "game_id")
);

CREATE TABLE "event_groups" (
  "id" UUID PRIMARY KEY,
  "owner_id" uuid NOT NULL,
  "game_mode_id" UUID NOT NULL,
  "sub_min" int NOT NULL DEFAULT 0,
  "registration_open" BOOLEAN NOT NULL DEFAULT true,
  "is_public" BOOLEAN NOT NULL DEFAULT false,
  "deprioritize_noshows" BOOLEAN NOT NULL DEFAULT false,
  "max_noshows" INT NOT NULL DEFAULT 0,
  "discord_guild" TEXT,
  "region" TEXT NOT NULL,
  "created_at" TIMESTAMPTZ NOT NULL,
  "updated_at" TIMESTAMPTZ NOT NULL
);

CREATE TABLE "events" (
  "id" UUID PRIMARY KEY,
  "group_id" UUID,
  "start_time" TIMESTAMPTZ,
  "created_at" TIMESTAMPTZ NOT NULL,
  "updated_at" TIMESTAMPTZ NOT NULL
);

CREATE TABLE "registrations" (
  "event_id" UUID,
  "user_id" UUID,
  "can_substitute" BOOLEAN NOT NULL DEFAULT false,
  "can_lobby_host" BOOLEAN NOT NULL DEFAULT false,
  "duo_request" text,
  "created_at" TIMESTAMPTZ NOT NULL,
  "updated_at" TIMESTAMPTZ NOT NULL,
  PRIMARY KEY ("event_id", "user_id")
);

CREATE TABLE "lobbies" (
  "id" UUID PRIMARY KEY,
  "event_id" UUID,
  "host" UUID,
  "created_at" TIMESTAMPTZ NOT NULL,
  "updated_at" TIMESTAMPTZ NOT NULL
);

CREATE TABLE "players" (
  "lobby_id" UUID,
  "user_id" UUID,
  "team_number" INT,
  "created_at" TIMESTAMPTZ NOT NULL,
  "updated_at" TIMESTAMPTZ NOT NULL,
  PRIMARY KEY ("lobby_id", "user_id")
);

CREATE TABLE "user_noshow" (
  "user_id" UUID,
  "event_id" UUID,
  PRIMARY KEY ("user_id", "event_id")
);

CREATE INDEX ON "users" ("discord_id");

CREATE INDEX ON "users" ("discord_name");

CREATE INDEX ON "api_links" ("user_id", "name");

CREATE INDEX ON "refresh_tokens" ("user_id");

CREATE INDEX ON "games" ("name");

CREATE INDEX ON "games" ("owner_id");

CREATE INDEX ON "game_modes" ("game_id");

CREATE INDEX ON "game_modes" ("owner_id");

CREATE INDEX ON "game_ranks" ("game_id");

CREATE INDEX ON "game_ranks" ("order");

CREATE INDEX ON "user_games" ("in_game_name");

CREATE INDEX ON "user_games" ("current_rank");

CREATE INDEX ON "user_games" ("peak_rank");

CREATE INDEX ON "user_games" ("api_links_id");

CREATE INDEX ON "event_groups" ("owner_id");

CREATE INDEX ON "event_groups" ("game_mode_id");

CREATE INDEX ON "events" ("group_id");

CREATE INDEX ON "events" ("start_time");

CREATE INDEX ON "registrations" ("duo_request");

CREATE INDEX ON "lobbies" ("event_id");

CREATE INDEX ON "lobbies" ("host");

ALTER TABLE "api_links" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE CASCADE DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "refresh_tokens" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE CASCADE DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "games" ADD FOREIGN KEY ("owner_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE CASCADE DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "game_modes" ADD FOREIGN KEY ("game_id") REFERENCES "games" ("id") ON DELETE CASCADE ON UPDATE CASCADE DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "game_modes" ADD FOREIGN KEY ("owner_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE CASCADE DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "game_ranks" ADD FOREIGN KEY ("game_id") REFERENCES "games" ("id") ON DELETE CASCADE ON UPDATE CASCADE DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "user_games" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE CASCADE DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "user_games" ADD FOREIGN KEY ("game_id") REFERENCES "games" ("id") ON DELETE CASCADE ON UPDATE CASCADE DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "user_games" ADD FOREIGN KEY ("current_rank") REFERENCES "game_ranks" ("id") ON DELETE SET NULL ON UPDATE CASCADE DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "user_games" ADD FOREIGN KEY ("peak_rank") REFERENCES "game_ranks" ("id") ON DELETE SET NULL ON UPDATE CASCADE DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "user_games" ADD FOREIGN KEY ("api_links_id") REFERENCES "api_links" ("id") ON DELETE SET NULL ON UPDATE CASCADE DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "event_groups" ADD FOREIGN KEY ("owner_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE CASCADE DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "event_groups" ADD FOREIGN KEY ("game_mode_id") REFERENCES "game_modes" ("id") ON DELETE SET NULL ON UPDATE CASCADE DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "events" ADD FOREIGN KEY ("group_id") REFERENCES "event_groups" ("id") ON DELETE CASCADE ON UPDATE CASCADE DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "registrations" ADD FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON DELETE CASCADE ON UPDATE CASCADE DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "registrations" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE CASCADE DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "lobbies" ADD FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON DELETE CASCADE ON UPDATE CASCADE DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "lobbies" ADD FOREIGN KEY ("host") REFERENCES "users" ("id") ON DELETE SET NULL ON UPDATE CASCADE DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "players" ADD FOREIGN KEY ("lobby_id") REFERENCES "lobbies" ("id") ON DELETE CASCADE ON UPDATE CASCADE DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "players" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE CASCADE DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "user_noshow" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE CASCADE DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "user_noshow" ADD FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON DELETE CASCADE ON UPDATE CASCADE DEFERRABLE INITIALLY IMMEDIATE;


-- +goose Down
DROP TABLE "users";
DROP TABLE "api_links";
DROP TABLE "refresh_tokens";
DROP TABLE "games";
DROP TABLE "game_modes";
DROP TABLE "game_ranks";
DROP TABLE "user_games";
DROP TABLE "event_groups";
DROP TABLE "events";
DROP TABLE "registrations";
DROP TABLE "lobbies";
DROP TABLE "players";
DROP TABLE "user_noshow";