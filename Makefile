.PHONY: prod prod-no-cache prod-build prod-logs prod-ps prod-down prod-down-v \
        dev dev-no-cache dev-build dev-logs dev-ps dev-down dev-down-v dev-clean \
        build push publish pull build-multi push-multi \
        health export-ca fix-certs tls-check test test-coverage test-docker \
        seed-users seed-events seed-registrations seed-all \
        seed-matchmaking-all seed-matchmaking-cleanup gen-keys \
        infra-init infra-fmt infra-validate infra-plan infra-apply infra-destroy \
        gcp-push gcp-migrate gcp-db-bootstrap gcp-deploy

## ── Configuration ─────────────────────────────────────────────
# IMAGE is the base repository (registry/namespace/name); each service appends
# its own suffix, e.g. $(IMAGE)-api and $(IMAGE)-frontend.
# Override on the command line, e.g. `make publish IMAGE=youruser/matchmaker VERSION=v1.2.3`
IMAGE   ?= matchmaker
VERSION ?= latest
PROJECT ?= matchmaker
DOMAIN  ?= matchmaker.localhost
FEEDBACK_URL ?= https://github.com/KatieSuth/Matchmaker/issues

# GCP (manual image deploy — see infra/README.md)
GCP_PROJECT ?=
GCP_REGION  ?= us-central1
GCP_TAG     ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo latest)
AR_REPO     ?= matchmaker-docker

# Exported so Docker Compose picks them up during interpolation
export IMAGE VERSION
export COMPOSE_PROJECT_NAME := $(PROJECT)

COMPOSE := docker compose
DEV     := $(COMPOSE) -f docker-compose.yml -f docker-compose.dev.yml
PROD    := $(COMPOSE) -f docker-compose.yml -f docker-compose.prod.yml
BAKE    := docker buildx bake -f docker-compose.yml -f docker-compose.prod.yml -f docker-bake.hcl

TF_DIR  := infra/terraform

# Caddy's container name
CADDY   := caddy

## ── Production ────────────────────────────────────────────────

# Build images and start all services (detached)
prod:
	$(PROD) up --build -d
	@echo "\nApp → https://$(DOMAIN)\n"

# Build without cache and start (detached)
prod-no-cache:
	$(PROD) build --no-cache --pull
	$(PROD) up -d
	@echo "\nApp → https://$(DOMAIN)\n"

# Build images without starting
prod-build:
	$(PROD) build

# Stream logs from all services
prod-logs:
	$(PROD) logs -f

# Show running containers
prod-ps:
	$(PROD) ps

# Stop and remove production containers
prod-down:
	$(PROD) down

# Stop and remove production containers and volumes
prod-down-v:
	$(PROD) down -v


## ── Development (hot reload) ──────────────────────────────────

# Start dev environment
dev:
	$(DEV) up --build
	@echo "\nDev mode — hot reload active"
	@echo "    App → https://$(DOMAIN)\n"

# Start dev environment without cache
dev-no-cache:
	$(DEV) build --no-cache --pull
	$(DEV) up
	@echo "\nDev mode — hot reload active"
	@echo "    App → https://$(DOMAIN)\n"

# Build images without starting
dev-build:
	$(DEV) build

# Stream logs from all services
dev-logs:
	$(DEV) logs -f

# Show running containers
dev-ps:
	$(DEV) ps

# Stop and remove dev containers
dev-down:
	$(DEV) down

# Stop and remove dev containers and volumes
dev-down-v:
	$(DEV) down -v

# Remove containers, images, and volumes
dev-clean:
	$(DEV) down --rmi all --volumes --remove-orphans


## ── Publishing (Docker Hub) ───────────────────────────────────

# Build production images tagged $(IMAGE)-<service>:$(VERSION)
build: prod-build

# Push production images (run `docker login` first)
push:
	$(PROD) push api frontend

# Build then push in one step
publish: build push

# Pull the published images from the registry
pull:
	$(PROD) pull api frontend

# Build multi-arch images (linux/amd64 + linux/arm64) into the build cache.
# Multi-platform builds can't be loaded into the local daemon — use push-multi
# to publish. Requires a buildx builder with QEMU emulation, created once via:
#   docker buildx create --name multiarch --driver docker-container --use --bootstrap
build-multi:
	$(BAKE)

# Build and push the multi-arch manifest to the registry (run `docker login` first)
push-multi:
	$(BAKE) --push


## ── Utilities ─────────────────────────────────────────────────

# Check API health
health:
	@curl -sk https://$(DOMAIN)/api/health | python3 -m json.tool

# Export Caddy's local root CA for browser trust (see README Firefox steps)
export-ca:
	docker exec $(CADDY) cat /data/caddy/pki/authorities/local/root.crt > caddy-root.crt
	@echo "Wrote caddy-root.crt — import into Firefox (Authorities tab) if not already trusted"

# Recover from expired/stuck Caddy TLS certs (stale lock or deleted cert files)
fix-certs:
	-docker exec $(CADDY) rm -f /data/caddy/locks/issue_cert_$(DOMAIN).lock
	docker restart $(CADDY)
	@bash scripts/tls-check.sh --wait $(DOMAIN)

# Check TLS certificate served by Caddy
tls-check:
	@bash scripts/tls-check.sh $(DOMAIN)

# Test the code (requires local Go install)
test:
	cd backend/internal && go test -p 1 ./...

# Test the code in Docker (no local Go install required)
test-docker:
	docker compose -f docker-compose.test.yml run --rm test
	docker compose -f docker-compose.test.yml down -v

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

## ── GCP / Terraform (see infra/README.md) ─────────────────────

infra-init:
	cd $(TF_DIR) && terraform init -backend-config=backend.hcl

infra-fmt:
	cd $(TF_DIR) && terraform fmt -recursive

infra-validate:
	cd $(TF_DIR) && terraform validate

infra-plan:
	cd $(TF_DIR) && terraform plan

infra-apply:
	cd $(TF_DIR) && terraform apply

infra-destroy:
	cd $(TF_DIR) && terraform destroy

# Build/push linux/amd64 images to Artifact Registry. Requires GCP_PROJECT.
# DOMAIN should be your public domain (e.g. matchmaker.games) for NEXT_PUBLIC_* bake.
gcp-push:
	@test -n "$(GCP_PROJECT)" || (echo "Set GCP_PROJECT=..." && exit 1)
	gcloud auth configure-docker $(GCP_REGION)-docker.pkg.dev --quiet
	$(eval AR := $(GCP_REGION)-docker.pkg.dev/$(GCP_PROJECT)/$(AR_REPO))
	docker build --platform linux/amd64 -t $(AR)/api:$(GCP_TAG) -f backend/Dockerfile backend
	docker push $(AR)/api:$(GCP_TAG)
	docker build --platform linux/amd64 -t $(AR)/frontend:$(GCP_TAG) -f frontend/Dockerfile frontend \
		--build-arg NEXT_PUBLIC_API_URL=https://$(DOMAIN)/api \
		--build-arg NEXT_PUBLIC_FRONTEND_DOMAIN=$(DOMAIN) \
		--build-arg NEXT_PUBLIC_COOKIE_AUTH_EXPIRE_LIMIT=604800 \
		--build-arg NEXT_PUBLIC_FEEDBACK_URL=$(FEEDBACK_URL)
	docker push $(AR)/frontend:$(GCP_TAG)
	@echo "Pushed $(AR)/api:$(GCP_TAG) and $(AR)/frontend:$(GCP_TAG)"

# One-shot least-privilege role bootstrap (postgres admin). Run after Apply A,
# before setting db_roles_bootstrapped=true. Not part of every gcp-deploy.
gcp-db-bootstrap:
	@test -n "$(GCP_PROJECT)" || (echo "Set GCP_PROJECT=..." && exit 1)
	$(eval AR := $(GCP_REGION)-docker.pkg.dev/$(GCP_PROJECT)/$(AR_REPO))
	gcloud run jobs update matchmaker-db-bootstrap \
		--project=$(GCP_PROJECT) --region=$(GCP_REGION) \
		--image=$(AR)/api:$(GCP_TAG)
	gcloud run jobs execute matchmaker-db-bootstrap \
		--project=$(GCP_PROJECT) --region=$(GCP_REGION) --wait

# Run goose via the matchmaker-migrate Cloud Run Job (private VPC → Cloud SQL).
# Point the job at the same API image tag you are about to roll out (or already pushed).
gcp-migrate:
	@test -n "$(GCP_PROJECT)" || (echo "Set GCP_PROJECT=..." && exit 1)
	$(eval AR := $(GCP_REGION)-docker.pkg.dev/$(GCP_PROJECT)/$(AR_REPO))
	gcloud run jobs update matchmaker-migrate \
		--project=$(GCP_PROJECT) --region=$(GCP_REGION) \
		--image=$(AR)/api:$(GCP_TAG)
	gcloud run jobs execute matchmaker-migrate \
		--project=$(GCP_PROJECT) --region=$(GCP_REGION) --wait

# Push images, migrate DB with the new API image, then roll Cloud Run services.
gcp-deploy: gcp-push
	@test -n "$(GCP_PROJECT)" || (echo "Set GCP_PROJECT=..." && exit 1)
	$(eval AR := $(GCP_REGION)-docker.pkg.dev/$(GCP_PROJECT)/$(AR_REPO))
	gcloud run jobs update matchmaker-migrate \
		--project=$(GCP_PROJECT) --region=$(GCP_REGION) \
		--image=$(AR)/api:$(GCP_TAG)
	gcloud run jobs execute matchmaker-migrate \
		--project=$(GCP_PROJECT) --region=$(GCP_REGION) --wait
	gcloud run services update matchmaker-api --project=$(GCP_PROJECT) --region=$(GCP_REGION) --image=$(AR)/api:$(GCP_TAG)
	gcloud run services update matchmaker-frontend --project=$(GCP_PROJECT) --region=$(GCP_REGION) --image=$(AR)/frontend:$(GCP_TAG)
