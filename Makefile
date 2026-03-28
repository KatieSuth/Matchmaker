.PHONY: up down build logs ps clean dev

## ── Production ────────────────────────────────────────────────

# Build images and start all services
up:
	docker compose -f docker-compose.yml -f docker-compose.prod.yml up --build -d
	@echo "\n✅  Frontend → https://localhost:3000"
	@echo "✅  API      → https://localhost:8080\n"

# Stop and remove containers
down:
	docker compose -f docker-compose.yml down
	
# Stop and remove containers and volumes
down-v:
	docker compose -f docker-compose.yml down -v

# Build images without starting
build:
	docker compose -f docker-compose.yml build

# Stream logs from all services
logs:
	docker compose -f docker-compose.yml logs -f

# Show running containers
ps:
	docker compose -f docker-compose.yml ps

# Remove containers, images, and volumes
clean:
	docker compose -f docker-compose.yml down --rmi all --volumes --remove-orphans


## ── Development (hot reload) ──────────────────────────────────

# Start dev environment
dev:
	docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build
	@echo "\n🔥  Dev mode — hot reload active"
	@echo "    Frontend → https://localhost:3000"
	@echo "    API      → https://localhost:8080\n"
	
# Start dev environment without cache
dev-no-cache:
	docker compose -f docker-compose.yml -f docker-compose.dev.yml build --no-cache
	docker compose -f docker-compose.yml -f docker-compose.dev.yml up
	@echo "\n🔥  Dev mode — hot reload active"
	@echo "    Frontend → https://localhost:3000"
	@echo "    API      → https://localhost:8080\n"

# Check API health
health:
	@curl -s https://localhost:8080/health | python3 -m json.tool
