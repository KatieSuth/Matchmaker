### Project Status
[![Backend](https://github.com/KatieSuth/Matchmaker/actions/workflows/backend.yml/badge.svg)](https://github.com/KatieSuth/Matchmaker/actions)
[![Backend Coverage](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/KatieSuth/c21c84f4fba3f91f41a5be25dd59326a/raw/matchmaker-coverage.json)](https://github.com/KatieSuth/Matchmaker/actions)
[![Frontend](https://github.com/KatieSuth/Matchmaker/actions/workflows/frontend.yml/badge.svg)](https://github.com/KatieSuth/Matchmaker/actions)

[Matchmaker](https://matchmaker.games)

# Matchmaker

Matchmaker is a free, open-source web application for organizing custom competitive games over Discord. It does this by allowing a game organizer to configure events that players can sign up to join, and it will ask players to provide their competitive rank (to be seen by the organizer but by default hidden to other players for privacy--can be configured in preferences). Once the organizer is ready, they can click a button to create 2 teams of the configured player count and sort players into those teams or a substitute pool as fairly as possibly based on their provided competitive ranks. It will also allow the organizer to make manual edits to the teams as needed.

It's now available live at https://matchmaker.games and currently supports Valorant and League of Legends. Eventually, users will have the ability to create their own custom settings for non-supported games or modify supported games (for example, 3v3 games instead of 5v5 games). See the roadmap for more information on where things are at & where we're going.

# Roadmap
- [x] Initial infrastructure setup (containers, DB, etc)
- [x] Basic login functionality via Discord
- [x] User preferences
- [x] One-off event admin configuration
- [x] One-off event sign-up for players
- [x] One-off event admin team creation
- [x] Player duo requests (best-effort at lobby and team assignment; balance takes priority)
- [x] 1.0 web hosting with public availability
- [ ] Riot API Linking for automatic competitive rank detection
- [ ] Discord server membership requirement as host option
- [ ] Attendance tracking and host alerts for repetitive no-show players
- [ ] Opt-in notifications for updates regarding events
- [ ] Allow hosts to ban users from their events
- [ ] Host option to auto create teams X time before match
- [ ] "Available Games" dashboard tab populated based on server membership/upcoming games
- [ ] Non-Riot game support (Overwatch 2, Marvel Rivals, user-defined, etc.)
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
| Local edge    | Caddy (Compose)                      |
| Production    | GCP Cloud Run + Cloud SQL + Cloudflare Worker — [runbook](infra/README.md), [design](infra/INFRASTRUCTURE.md) |

---

### Prerequisites
- [Docker](https://docs.docker.com/get-docker/) + Docker Compose v2
- `make` (optional but handy)

### Environment Setup

For a local setup, create your `.env` file by copying the existing `.env.example` in the repository root to `.env`. See below for a list of the expected environment variables. You can use `make gen-keys` to generate JWT_SECRET, COOKIE_HASH_KEY, and COOKIE_ENCRYPT_KEY values; alternatively, you can use `openssl rand -hex 32`. Your PostgreSQL database will be initialized with whatever values you've provided, so ensure you've replaced each value (including the ones in the Database URLs) with the value you want it to be and that you've picked a strong password.

Updating the root `.env` is essential for local development and Docker runs alike. Defaults for Docker Compose are provided but can be overridden here.

Update the URLs in the `.env` files. By default your domain is expected to be `matchmaker.localhost`, but this can be changed by setting the `DOMAIN` variable in the root `.env` file. Caddy and the derived URLs pick up that value automatically.

You will also need to create a Discord application: https://discord.com/developers/applications. Once you've created it, navigate to the OAuth2 tab and create/copy the Client ID and Secret into the appropriate places in the root `.env` file. Add your redirect URI; it should be `https://${DOMAIN}/api/auth/discord_redirect`. You do not need to generate an OAuth2 URL; the application fills in the scopes it needs in `backend/cmd/server/main.go` (it only uses identify and guilds).

### Development (hot reload)

Both services reload on file save.
- Next.js uses `next dev` with volume-mounted source
- Gin uses [`air`](https://github.com/air-verse/air) for live-reload

1. Prior to first run, navigate to the frontend directory and run `npm install` if using VSCode for development.

2.
```bash
make dev
# or:
docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build
```

3: Manually export and install the CA cert file from caddy's container. From the same directory as the Caddyfile while the containers are running:

```bash
# 3a. Copy the CA cert out of the caddy_data volume
docker exec caddy cat /data/caddy/pki/authorities/local/root.crt > caddy-root.crt
# or run
make export-ca
```

``` bash
# choose the right one for you:
# 3b. install on macOS
sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain caddy-root.crt

# 3b. install on Windows (run in elevated PowerShell)
Import-Certificate -FilePath "caddy-root.crt" -CertStoreLocation Cert:\LocalMachine\Root

# 3b. install on Linux (Ubuntu/Debian)
sudo cp caddy-root.crt /usr/local/share/ca-certificates/caddy.crt
sudo update-ca-certificates
```

#### Firefox-specific steps

1. Firefox requires a couple of manual changes to run as expected in a development environment. For SSL to work, Caddy's local CA cert must be exported from the container and imported into browser manually since Firefox has its own cert store separate from the OS. Run `make export-ca` (or follow the `docker exec` step in 3a) to download the cert. Next, import it directly into Firefox:

```
1. Settings → Privacy & Security → search for Certificates -> Manage certificates
2. Authorities tab → Import
3. Select the caddy-root.crt file you exported earlier
4. Check "Trust this CA to identify websites"
5. OK → restart Firefox
```

2. Firefox treats localhost itself as a secure context by default for cookie purposes, but it doesn't treat .localhost subdomains the same way. To prevent Firefox from rejecting the auth-related cookies:
```
1. In Firefox, go to about:config
2. Search for network.dns.localDomains
3. Set or add the value: your `${DOMAIN}`
```

#### TLS troubleshooting (Firefox "security issue" / expired cert)

Caddy stores dev TLS certificates in a Docker volume. If the stack was stopped for a while, the site certificate can expire. Caddy may then keep serving the expired cert while renewal gets stuck on a stale lock file.

```bash
make fix-certs    # clears stale lock, restarts Caddy, re-issues cert
make tls-check    # prints cert dates and verify status
make export-ca    # re-export root CA if needed (usually unchanged)
```

If Firefox still warns after `make fix-certs`, confirm `caddy-root.crt` is imported under **Authorities** (not **Your Certificates**) and restart Firefox.

---

## API Reference

In a developer environment setup, the API is available at `https://${DOMAIN}/api/`

The available paths are documented using the OpenAPI specification. See [openapi.yaml](https://github.com/KatieSuth/Matchmaker/blob/main/backend/openapi.yaml)

### Example

```bash
# Health
curl https://${DOMAIN}/api/health
```

---

## Environment Variables

Frontend and API settings live in the **root** [`.env.example`](.env.example) (Compose interpolates `NEXT_PUBLIC_*` from `DOMAIN`).

### Frontend (from root `.env` / Compose)

| Variable                               | Default                             | Description                                                                       |
|----------------------------------------|-------------------------------------|-----------------------------------------------------------------------------------|
| `NEXT_PUBLIC_API_URL`                  | `https://${DOMAIN}/api`             | Gin API base URL                                                                  |
| `NEXT_PUBLIC_FRONTEND_DOMAIN`          | `${DOMAIN}`                         | Frontend host or origin (`allowedDevOrigins`, `auth_session` cookie `Domain`)     |
| `NEXT_PUBLIC_COOKIE_AUTH_EXPIRE_LIMIT` | `604800` (7 days)                   | Length of time for auth expiration (should match REFRESH_EXPIRE_LIMIT in backend) |
| `NODE_ENV`                             | `development`                       | Node environment flag                                                             |
| `ORIGIN_VERIFY_SECRET`                 | unset (gate off)                    | When set (Cloud Run), requires `X-Origin-Verify`; `/health` exempt                |
| `SENTRY_DSN`                           | unset                               | When set, enables Sentry in the Next.js server runtime                            |


### Environment Variables (`.env.example`)

| Variable                        | Default                                                  | Description                                                                           |
|---------------------------------|----------------------------------------------------------|---------------------------------------------------------------------------------------|
| `COOKIE_HASH_KEY`               | required, no default                                     | A hash key for the secure cookie used on OAuth2 login                                 |
| `COOKIE_ENCRYPT_KEY`            | required, no default                                     | An encrypt key for the secure cookie used on OAuth2 login                             |
| `COOKIE_DOMAIN`                 | `${DOMAIN}`                                              | The domain used for setting cookies on login                                          |
| `DATABASE_URL`                  | required, no default                                     | URL to connect to the DB                                                              |
| `DATABASE_URL_TESTS`            | no default                                               | URL to connect to the DB for testing                                                  |
| `DB_MAX_CONNS`                  | unset (pgx default)                                      | Optional pgx pool max connections (prod sets `5`)                                     |
| `DISCORD_CLIENT_ID`             | required, no default                                     | Client ID provided by Discord developer portal app                                    |
| `DISCORD_CLIENT_SECRET`         | required, no default                                     | Client Secret provided by Discord developer portal app                                |
| `DISCORD_REDIRECT_URI`          | `https://${DOMAIN}/api/auth/discord_redirect`            | Discord redirect URI for OAuth2, configured in developer portal                       |
| `DISCORD_API_URL`               | `https://discord.com/api`                                | Discord API URL (for test flexibility)                                                |
| `DOMAIN`                        | `matchmaker.localhost`                                   | Public domain served by Caddy; the other public URLs above default to this value      |
| `FRONTEND_URL`                  | `https://${DOMAIN}`                                      | CORS allowed origin                                                                   |
| `GIN_MODE`                      | `release`                                                | `debug` or `release`                                                                  |
| `JWT_SECRET`                    | required, no default                                     | A key for signing the JWT access tokens                                               |
| `ORIGIN_VERIFY_SECRET`          | unset (gate off)                                         | When set, require `X-Origin-Verify` (Cloudflare Worker); `/health` exempt             |
| `PORT`                          | `8080`                                                   | Server port                                                                           |
| `POSTGRES_DB`                   | required, no default                                     | Postgres Database name                                                                |
| `POSTGRES_PASSWORD`             | required, no default                                     | Postgres Database password                                                            |
| `POSTGRES_USER`                 | required, no default                                     | Postgres Database username                                                            |
| `REFRESH_EXPIRE_LIMIT`          | `604800` (7 days)                                        | Time in seconds for the expiration of the refresh tokens                              |
| `SENTRY_DSN`                    | unset                                                    | When set, enables Sentry error reporting                                              |
| `TRUSTED_PROXIES`               | `172.20.0.0/16`                                          | Comma-separated CIDRs trusted for `X-Forwarded-For` (Gin)                             |
| `FAIRNESS_OUTLIER_GAP`          | `6`                                                      | Baseline outlier rank gap for per-lobby fairness warnings (at reference tier count)   |
| `FAIRNESS_TEAM_SEPARATION`      | `3`                                                      | Baseline team average rank separation for fairness warnings (at reference tier count) |
| `FAIRNESS_REFERENCE_TIER_COUNT` | `25`                                                     | Tier count the fairness baselines are calibrated for (Valorant)                       |

### Production (GCP)

- **How to deploy:** [infra/README.md](infra/README.md) (accounts, Terraform, images, teardown). You create GCP / Cloudflare / Sentry / Discord accounts manually.
- **What / why:** [infra/INFRASTRUCTURE.md](infra/INFRASTRUCTURE.md) (architecture, Terraform layout, Worker, origin-verify, cost decisions).

---

## Useful Commands

```bash
make prod               # Build & start production containers (detached)
make prod-no-cache      # Build production containers without cache and start (detached)
make prod-build         # Build production images without starting
make prod-logs          # Stream logs from all production containers
make prod-ps            # Show running production containers
make prod-down          # Stop and remove production containers
make prod-down-v        # Stop and remove production containers and volumes

make dev                # Start development containers with hot reload
make dev-no-cache       # Build development containers without cache and start with hot reload
make dev-down           # Stop and remove development containers
make dev-down-v         # Stop and remove development containers and volumes
make dev-clean          # Remove development containers, images, and volumes
make dev-build          # Build development containers without running
make dev-logs           # Stream logs from all development containers
make dev-ps             # Show running development containers

make build              # Build production images tagged for the configured registry
make push               # Push production images to the registry (run `docker login` first)
make publish            # Build then push production images in one step
make pull               # Pull the published images from the registry
make build-multi        # Build multi-arch images (linux/amd64 + linux/arm64) into the build cache
make push-multi         # Build and push the multi-arch manifest to the registry

make health             # Quick API health check
make test               # Run Go tests
make test-coverage      # Run Go tests with coverage percentage output, filtering for test-supporting files

make seed-users         # Dev-only: seed 30 users
make seed-events        # Dev-only: seed 20 groups and adjacent events (run after seed-users)
make seed-registrations # Dev-only: seed event registrations (run after seed-events)
make seed-all           # Dev-only: run all seed scripts in required order

make seed-matchmaking-all HOST=YourDiscordName  # Dev-only: seed matchmaking test event groups (see below)
make seed-matchmaking-cleanup                   # Dev-only: remove all matchmaking test data created by the script

make gen-keys           # Generate new JWT, Cookie Hash, and Cookie Encrypt keys

# GCP / Terraform (see infra/README.md — accounts are created manually)
make infra-init         # terraform init (requires backend.hcl)
make infra-fmt          # terraform fmt
make infra-validate     # terraform validate
make infra-plan         # terraform plan
make infra-apply        # terraform apply
make infra-destroy      # terraform destroy
make gcp-push GCP_PROJECT=... DOMAIN=matchmaker.games   # build/push images to Artifact Registry
make gcp-deploy GCP_PROJECT=... DOMAIN=matchmaker.games  # push + update Cloud Run
```

Seed commands are for local development data only and are not part of production build, deploy, or runtime flows.

### Publishing images to Docker Hub

The `api` and `frontend` services are tagged as `${IMAGE}-api:${VERSION}` and `${IMAGE}-frontend:${VERSION}`, where `IMAGE` is the base repository (registry/namespace/name) and `VERSION` is the tag (defaults to `latest`). Set those in the root `.env` or pass them as `make` variables, log in, then publish:

```bash
docker login
make publish IMAGE=youruser/matchmaker VERSION=v1.0.0
```

`make publish` runs `make build` (production images via the production Dockerfiles) followed by `make push`. Use `make pull` on the target host to fetch the published images.

#### Multi-architecture images

To publish images that run on both `linux/amd64` and `linux/arm64` (e.g. Intel/AMD servers and ARM hosts like Apple Silicon or Raspberry Pi), use the Buildx-based targets backed by `docker-bake.hcl`. Create a builder once, then push:

```bash
docker buildx create --name multiarch --driver docker-container --use --bootstrap
docker login
make push-multi IMAGE=youruser/matchmaker VERSION=v1.0.0
```

`make push-multi` builds both architectures and pushes a single multi-arch manifest in one step (multi-platform images can't be loaded into the local Docker daemon, so use `make build-multi` only to warm the build cache). The platforms are defined in `docker-bake.hcl`, which reuses the build contexts, image names, and build args from the Compose files.

### Matchmaking test data

Seeds ~54 event groups with scenario-specific users (controlled counts, ranks, subs, and duo pairings) to exercise matchmaking manually. Each group is left open for you to review registrations and click **Lock In & Create Teams**. Users are owned by each scenario — not shared — so ranks stay stable across manual re-testing.

**Prerequisites:** migrations applied; your host account must already exist in the database (log in via the app once).

```bash
make seed-matchmaking-all HOST=YourDiscordName
```

Replace `YourDiscordName` with your `users.discord_name` value. The script prints a table of `event_group.id` values with a short description of what each group is meant to test (insufficient players, single lobby, single lobby with overflow subs/unplaced, two lobbies with subs, fairness warnings, balanced vs ranked modes, etc.).

Optional flags via direct CLI (not exposed in Makefile):

```bash
cd backend && go run ./cmd/scripts/matchmaking all --host=YourDiscordName
cd backend && go run ./cmd/scripts/matchmaking all --host=YourDiscordName --json
```

Re-running is idempotent: existing scenario groups are skipped. To re-seed from scratch:

```bash
make seed-matchmaking-cleanup
make seed-matchmaking-all HOST=YourDiscordName
```

---

# FAQ
### Why do I need this when Valorant custom games have an autobalance button?

The built-in team balancer is a great resource if you have exactly 10 players and just don't know how to configure the teams. If you have a pool of more than 10 players and could potentially be running multiple games at once though, it cannot be used to fairly determine who should be in which lobby. That said, this isn't just for enormous Discord servers that will have tens of players joining at once; it can also be used to determine good 2v2 or 3v3 matches from a pool of available competitors for Skirmish or other modes.

### You've got a lot of roadmap there and not a lot of journey. Is (the thing on the roadmap I'm waiting for) ever going to be done?

This application currently has one developer and I work on it when I can, so things might take a while. Thanks for the interest though and please feel free to add issues & contribute! I will review them as I'm able.

### How does matchmaking work?

When the host clicks **Lock In & Create Teams**, Matchmaker looks at everyone signed up for each game and builds teams automatically using the rank information players provided (current rank and peak rank).

For each player, it calculates a single skill number by averaging those two ranks. That number is what gets used for sorting and balancing — not just current rank alone.

The host chooses a **matchmaking mode** when creating the event:

- **Balanced** (default) — Best for casual games. Matchmaker tries to spread different skill levels across the lobby and split players into two even teams. If more people signed up than can play at once, it picks a mix from high, middle, and low ranks rather than only taking the highest-ranked or lowest-ranked players.
- **Rank Grouping** — Best for serious practice. Matchmaker groups players of similar rank into the same lobby. If not everyone can play, it keeps whichever skill level most players belong to (for example, if most signups are lower rank, lower-ranked players are prioritized for roster spots).

After teams are formed, a few other things happen automatically:

- Players who marked **Can substitute** during sign-up may be placed in a sub pool instead of a team if there are more signups than spots.
- Players who did **not** mark **Can substitute** and did not make a team stay signed up but are listed as unplaced for that game.
- If enough players signed up, multiple lobbies can be created so more than one game can run at once.
- One player per lobby is assigned as lobby host (whoever volunteered first, or the earliest sign-up if no one volunteered).
- **Duo requests** — If two players list each other's Discord name as their duo partner, Matchmaker tries to put them in the same lobby and on the same team. Balance comes first; duos are not guaranteed if fairness requires otherwise. Only one duo pair per team is attempted.

The host can still review the results and make manual changes afterward.

For more detailed technical information, see [backend/internal/matchmaking/MATCHMAKING.md](backend/internal/matchmaking/MATCHMAKING.md).

### How do I know the matches are fair?

Matchmaker does its best to create even teams, but perfect balance is not always possible — especially when signups span a very wide range of ranks (for example, Iron through Immortal in the same Valorant lobby).

After teams are created, each lobby is checked for balance. If Matchmaker could not get a good result, you will see a warning on that game or lobby. The message explains that teams were formed with the best available balance, but the rank spread was too wide for fully fair teams in that lobby.

A warning usually means one of two things:

- One player is much higher or lower ranked than everyone else in the lobby (a large skill gap).
- The two teams' average skill levels are still noticeably different after balancing.

If you are the event host, you can also see each team's average rank on the team headers to judge balance yourself.

Fairness warnings are a heads-up, not a failure — the teams are still playable. The host can always adjust rosters manually if something looks off.

### Why does my event name hit the 50-character limit so fast when I use emojis?

Event names are capped at 50 characters. Ordinary letters and numbers usually count as one character each, but many emojis (and especially ones made of several parts, like some family or flag emojis) count as more than one toward that limit even though they look like a single symbol on screen. That’s just how computers store those characters; it isn’t a bug in the counter. The counter on the form shows how much of the limit you have left; if it turns red, shorten the name or use fewer emojis.

### How can teams be edited?

After you click **Lock In & Create Teams**, the host can adjust rosters without deleting and recreating everything. Edits are done **one game at a time** (each scheduled match in the group has its own teams).

To move players around:

1. Open the event group page and select the game you want to change.
2. On any player card in a **team**, **subs**, or **unplaced** list, open the **⋯** menu.
3. Choose **Swap**.
4. Pick another player from the dropdown — anyone on a different team, in another lobby, in the sub pool, or on the unplaced list — and click **Submit**.

A swap exchanges the two players' spots. Each player takes the other's exact placement (lobby, team, sub slot, or unplaced status). You can use this to rebalance teams, move someone into or out of subs, pull an unplaced player onto a roster, or send a rostered player to unplaced.

A few rules to be aware of:

- You **cannot** swap two players already in the **same group**. You can't swap two players on the same team; you can't swap two substitutes in the same lobby with each other; you can't swap two unplaced players with each other. They won't appear as options for your swap.
- If a rostered player who did **not** sign up as a substitute is swapped into a sub slot, they become **unplaced** instead — only substitute volunteers can fill sub spots.
- If your event requires a minimum number of substitutes **and** has more than one lobby, a swap is blocked when it would leave any lobby below that minimum.

After each swap, Matchmaker refreshes lobby hosts and fairness warnings for the affected lobbies so you can see whether balance changed. The original warning from lock-in is kept separately so you can still tell what the auto-generated teams looked like.

If you want to undo many changes at once, see the next FAQ entry about **Delete teams**.

### I've edited a team, but I don't like what I've done. How do I get back to where it started?

To get back to the default teams that were created, simply click "Delete teams" in the event header. Doing this will reset the event back to registrations-only from before teams were created, but registrations will remain closed. To regenerate the teams, click "Create Teams" in the event header and the teams will be created again the same way they were before.

### I've closed registrations and created teams, but a new player wants to join. How do I add them?

As the host, you can't manually add players; all players must register themselves. To let the new player in, click "Delete teams" in the event header. This will reset the event back to the registrations, but registrations will remain closed. To reopen them, click "Edit" in the event header. At the bottom of the modal, there is a "Registration Status" toggle. Toggle it on and click "Save Settings" to allow new players to register.