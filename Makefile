.PHONY: up down build logs ps clean dev

## ── Production ────────────────────────────────────────────────

# Build images and start all services
prod:
	docker compose -f docker-compose.yml -f docker-compose.prod.yml up --build -d
	@echo "\n✅  Frontend → https://localhost:3000"
	@echo "✅  API      → https://localhost:8080\n"
	
prod-no-cache:
	docker compose -f docker-compose.yml -f docker-compose.prod.yml build --no-cache
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
	docker compose -f docker-compose.yml -f docker-compose.dev.yml build --no-cache
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

test:
	cd backend/internal && go test ./...
