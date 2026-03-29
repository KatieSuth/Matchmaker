# Matchmaker

Matchmaker is a free, open-source web application for organizing custom competitive games over Discord. It does this by allowing a game organizer to configure events that players can sign up to join, and it will ask players to provide their competitive rank (to be seen by the organizer but by default hidden to other players for privacy--can be configured in preferences). Once the organizer is ready, they can click a button to create 2 teams of the configured player count and sort players into those teams or a substitute pool as fairly as possibly based on their provided competitive ranks. It will also allow the organizer to make manual edits to the teams as needed.

It's still very much in the early phases but is intended to one day support Valorant and League of Legends, along with the ability for users to create their own custom settings for non-supported games or modify supported games (for example, 3v3 games instead of 5v5 games). See the roadmap for more information on where things are at & where we're going.

# Roadmap
- [x] Initial infrastructure setup (containers, DB, etc)
- [ ] Basic login functionality via Discord
- [ ] User preferences
- [ ] One-off event admin configuration
- [ ] One-off event sign-up for players
- [ ] One-off event admin team creation
- [ ] 1.0 web hosting with public availability
- [ ] Player duo requests
- [ ] Attendance tracking and host alerts for repetitive no-show players
- [ ] Opt-in notifications for updates regarding events
- [ ] Riot API Linking for automatic competitive rank detection
- [ ] Non-Riot game support (Overwatch 2, Marvel Rivals, etc.)
- [ ] Non-Discord login support
- [ ] Tournament bracket event support

# Quick Start

## Tech Stack

| Layer         | Technology                           |
|---------------|--------------------------------------|
| Frontend      | Next.js 16, TypeScript, App Router   |
| Backend       | Go 1.25, Gin, CORS middleware, sqlc  |
| Container     | Docker, Docker Compose v2            |
| Dev DX        | `air` (Go), `next dev` (Node)        |
| Database      | Postgres, pgx (v5), Goose            |
| Load balancer | Caddy                                |

---

### Prerequisites
- [Docker](https://docs.docker.com/get-docker/) + Docker Compose v2
- `make` (optional but handy)

### Development (hot reload)

Both services reload on file save.

Prior to first run, navigate to the frontend directory and run `npm install` if using VSCode for development.

```bash
make dev
# or:
docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build
```

- Next.js uses `next dev` with volume-mounted source
- Gin uses [`air`](https://github.com/air-verse/air) for live-reload

---

## API Reference

| Method | Path                  | Description        |
|--------|-----------------------|--------------------|
| GET    | `/health`             | Health check       |

### Example

```bash
# Health
curl http://api.matchmaker.localhost/health
```

---

## Environment Variables

### Frontend (`frontend/.env.example`)

| Variable               | Default                             | Description              |
|------------------------|-------------------------------------|--------------------------|
| `NEXT_PUBLIC_API_URL`  | `https://api.matchmaker.localhost`  | Gin API base URL         |
| `NODE_ENV`             | `development`                       | Node environment flag    |

### Backend (`backend/.env.example`)

| Variable             | Default                       | Description                                               |
|----------------------|-------------------------------|-----------------------------------------------------------|
| `PORT`               | `8080`                        | Server port                                               |
| `DATABASE_URL`       | required, no default          | URL to connect to the DB                                  |
| `POSTGRES_USER`      | required, no default          | Postgres Database username                                |
| `POSTGRES_PASSWORD`  | required, no default          | Postgres Database username                                |
| `COOKIE_HASH_KEY`    | required, no default          | A hash key for the secure cookie used on OAuth2 login     |
| `COOKIE_ENCRYPT_KEY` | required, no default          | An encrypt key for the secure cookie used on OAuth2 login |
| `POSTGRES_DB`        | required, no default          | Postgres Database name                                    |
| `GIN_MODE`           | `release`                     | `debug` or `release`                                      |
| `FRONTEND_URL`       | `http://matchmaker.localhost` | CORS allowed origin                                       |

---

## Useful Commands

```bash
make prod            # Build & start production containers (detached)
make prod-no-cache   # Build production containers without cache and start
make prod-build      # Build production images without starting
make prod-logs       # Stream logs from all production containers
make prod-ps         # Show running production containers

make dev           # Start development containers with hot reload
make dev-no-cache  # Build development containers without cache and start with hot reload
make dev-down      # Stop and remove development containers
make dev-down-v    # Stop and remove development containers and volumes
make dev-clean     # Remove development containers, images, and volumes
make dev-build     # Build development containers without running
make dev-logs      # Stream logs from all development containers
make dev-ps        # Show running development containers

make health        # Quick API health check
```

# FAQ
### Why do I need this when Valorant custom games have an autobalance button?

The built-in team balancer is a great resource if you have exactly 10 players and just don't know how to configure the teams. If you have a pool of more than 10 players and could potentially be running multiple games at once though, it cannot be used to fairly determine who should be in which lobby. That said, this isn't just for enormous Discord servers that will have tens of players joining at once; it can also be used to determine good 2v2 or 3v3 matches from a pool of available competitors for Skirmish or other modes.

### You've got a lot of roadmap there and not a lot of journey. Is this thing ever going to be done?

This application currently has one developer and I work on it when I can, so things might take a while. Thanks for the interest though and please feel free to add issues & contribute! I will review them as I'm able.
