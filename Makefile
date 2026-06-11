.PHONY: up down build logs ps clean dev seed-users seed-events seed-registrations seed-all seed-matchmaking-all seed-matchmaking-cleanup

## ── Production ────────────────────────────────────────────────

# Build images and start all services
prod:
	docker compose -f docker-compose.yml -f docker-compose.prod.yml up --build -d
	@echo "\n✅  Frontend → https://localhost:3000"
	@echo "✅  API      → https://localhost:8080\n"
	
prod-no-cache:
	docker compose -f docker-compose.yml -f docker-compose.prod.yml build --no-cache --pull
	docker compose -f docker-compose.yml -f docker-compose.prod.yml up
	@echo "\n✅  Frontend → https://localhost:3000"
	@echo "✅  API      → https://localhost:8080\n"

# Build images without starting
prod-build:
	docker compose -f docker-compose.yml -f docker-compose.prod.yml build

# Stream logs from all services
prod-logs:
	docker compose -f docker-compose.yml -f docker-compose.prod.yml logs -f

# Show running containers
prod-ps:
	docker compose -f docker-compose.yml -f docker-compose.prod.yml ps


## ── Development (hot reload) ──────────────────────────────────

# Start dev environment
dev:
	docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build
	@echo "\n🔥  Dev mode — hot reload active"
	@echo "    Frontend → https://matchmaker.localhost"
	@echo "    API      → https://matchmaker.localhost/api\n"
	
# Start dev environment without cache
dev-no-cache:
	docker compose -f docker-compose.yml -f docker-compose.dev.yml build --no-cache --pull
	docker compose -f docker-compose.yml -f docker-compose.dev.yml up
	@echo "\n🔥  Dev mode — hot reload active"
	@echo "    Frontend → https://matchmaker.localhost"
	@echo "    API      → https://matchmaker.localhost/api\n"

# Stop and remove dev containers
dev-down:
	docker compose -f docker-compose.yml -f docker-compose.dev.yml down

# Stop and remove dev containers and volumes
dev-down-v:
	docker compose -f docker-compose.yml -f docker-compose.dev.yml down -v

# Build images without starting
dev-build:
	docker compose -f docker-compose.yml -f docker-compose.dev.yml build

# Stream logs from all services
dev-logs:
	docker compose -f docker-compose.yml -f docker-compose.dev.yml logs -f

# Show running containers
dev-ps:
	docker compose -f docker-compose.yml -f docker-compose.dev.yml ps

# Remove containers, images, and volumes
dev-clean:
	docker compose -f docker-compose.yml -f docker-compose.dev.yml down --rmi all --volumes --remove-orphans

# Check API health
health:
	@curl -sk https://matchmaker.localhost/api/health | python3 -m json.tool

# Export Caddy's local root CA for browser trust (see README Firefox steps)
export-ca:
	docker exec matchmaker-caddy cat /data/caddy/pki/authorities/local/root.crt > caddy-root.crt
	@echo "✅  Wrote caddy-root.crt — import into Firefox (Authorities tab) if not already trusted"

# Recover from expired/stuck Caddy TLS certs (stale lock or deleted cert files)
fix-certs:
	-docker exec matchmaker-caddy rm -f /data/caddy/locks/issue_cert_matchmaker.localhost.lock
	docker restart matchmaker-caddy
	@bash scripts/tls-check.sh --wait matchmaker.localhost

# Check TLS certificate served by Caddy
tls-check:
	@bash scripts/tls-check.sh matchmaker.localhost

# Test the code
test:
	cd backend/internal && go test ./...

# Test the code and output coverage percentage, excluding generated files (./backend/internal/db/*, ./backend/internal/test_util/*, and ./backend/internal/store/mock_store.go)
test-coverage:
	@cd backend/internal && \
	PKGS=$$(go list ./... | grep -vE "db|test_util") && \
	go test -coverprofile=coverage.out $$PKGS && \
	sed -i.bak '/mock_store.go/d' coverage.out && rm coverage.out.bak && \
	go tool cover -func=coverage.out

# Seed local development users
seed-users:
	cd backend && go run ./cmd/scripts/users

# Seed local development event groups/events
seed-events:
	cd backend && go run ./cmd/scripts/events

# Seed local development registrations
seed-registrations:
	cd backend && go run ./cmd/scripts/registrations

# Run all local development seeders in required order
seed-all: seed-users seed-events seed-registrations

# Seed matchmaking test event groups for manual lock-in testing
seed-matchmaking-all:
	cd backend && go run ./cmd/scripts/matchmaking all --host=$(HOST)

# Remove all matchmaking test data created by seed-matchmaking-all
seed-matchmaking-cleanup:
	cd backend && go run ./cmd/scripts/matchmaking cleanup
	
# Generate new secure keys
gen-keys:
	cd backend && go run ./cmd/scripts/gen_keys