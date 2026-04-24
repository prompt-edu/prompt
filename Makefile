.PHONY: help server servers servers-1000 client-core client-certificate client-assessment \
	client-interview client-matching clients clients-1000 db db-template db-1000 db-down \
	server-core server-assessment server-interview \
	server-team-allocation server-self-team-allocation server-template \
	server-certificate \
	lint lint-clients lint-servers \
	test test-core test-assessment test-interview \
	test-team-allocation test-self-team-allocation test-template \
	test-certificate \
	sqlc sqlc-core sqlc-assessment sqlc-interview \
	sqlc-team-allocation sqlc-self-team-allocation sqlc-template \
	sqlc-certificate \
	swagger install-clients install-hooks

# Load .env file if it exists (base configuration)
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Load .env.dev file if it exists (local development overrides)
ifneq (,$(wildcard ./.env.dev))
    include .env.dev
    export
endif

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

# ─── Development ───────────────────────────────────────────────────────────────

server: server-core ## Start the core server (alias)

servers: ## Start all servers (core + all microservices)
	@echo "Starting all servers..."
	@$(MAKE) server-core &
	@$(MAKE) server-assessment &
	@$(MAKE) server-interview &
	@$(MAKE) server-team-allocation &
	@$(MAKE) server-self-team-allocation &
	@$(MAKE) server-template &
	@$(MAKE) server-certificate &
	@wait
	@echo "All servers started."

servers-1000: ## Start project week servers (core, team allocation, self team allocation, template)
	@echo "Starting project week servers..."
	@trap 'kill 0' INT TERM EXIT; \
	$(MAKE) server-core SERVER_ADDRESS=0.0.0.0:8080 & \
	$(MAKE) server-team-allocation SERVER_ADDRESS=0.0.0.0:8083 & \
	$(MAKE) server-self-team-allocation SERVER_ADDRESS=0.0.0.0:8084 & \
	$(MAKE) server-template SERVER_ADDRESS=0.0.0.0:8086 & \
	wait

server-core: ## Start core server (port 8080)
	cd servers/core && go run main.go

server-assessment: ## Start assessment server (port 8085)
	cd servers/assessment && go run main.go

server-interview: ## Start interview server (port 8087)
	cd servers/interview && go run main.go

server-team-allocation: ## Start team allocation server (port 8083)
	cd servers/team_allocation && go run main.go

server-self-team-allocation: ## Start self team allocation server (port 8084)
	cd servers/self_team_allocation && go run main.go

server-template: ## Start template server (port 8086)
	cd servers/template_server && SERVER_ADDRESS=0.0.0.0:8086 go run main.go

server-certificate: ## Start certificate server (port 8088)
	cd servers/certificate && go run main.go

clients: ## Start all client micro-frontends
	cd clients && yarn install && yarn run dev

clients-1000: ## Start only the core, team allocation, and self team allocation clients
	@echo "Starting project week clients..."
	( cd clients/core && yarn dev ) & \
	( cd clients/team_allocation_component && yarn dev ) & \
	( cd clients/self_team_allocation_component && yarn dev ) & \
	wait

client-core: ## Start only the core client
	cd clients/core && yarn dev

client-certificate: ## Start only the certificate client
	cd clients/certificate_component && yarn dev

client-assessment: ## Start only the assessment client
	cd clients/assessment_component && yarn dev

client-interview: ## Start only the interview client
	cd clients/interview_component && yarn dev

client-matching: ## Start only the matching client
	cd clients/matching_component && yarn dev

db: ## Start database, template database, and Keycloak
	docker compose up -d db db-template-server keycloak

db-template: ## Start only the template server database
	docker compose up -d db-template-server

db-1000: ## Start databases and infrastructure needed for project week services, including template DB
	docker compose up -d \
	db \
	db-template-server \
	db-team-allocation \
	db-self-team-allocation \
	keycloak-db \
	keycloak \
	seaweedfs-master \
	seaweedfs-volume \
	seaweedfs-filer \
	seaweedfs-s3

db-down: ## Stop database, template database, and Keycloak
	docker compose stop db db-template-server keycloak

# ─── Code Quality ──────────────────────────────────────────────────────────────

lint: lint-clients lint-servers ## Lint all code

lint-clients: ## Lint all clients
	cd clients && yarn eslint "core" --config "core/eslint.config.mjs"
	cd clients && yarn eslint "shared_library" --config "shared_library/eslint.config.mjs"
	cd clients && yarn eslint "assessment_component" --config "assessment_component/eslint.config.mjs"
	cd clients && yarn eslint "interview_component" --config "interview_component/eslint.config.mjs"
	cd clients && yarn eslint "matching_component" --config "matching_component/eslint.config.mjs"
	cd clients && yarn eslint "self_team_allocation_component" --config "self_team_allocation_component/eslint.config.mjs"
	cd clients && yarn eslint "team_allocation_component" --config "team_allocation_component/eslint.config.mjs"
	cd clients && yarn eslint "template_component" --config "template_component/eslint.config.mjs"
	cd clients && yarn eslint "certificate_component" --config "certificate_component/eslint.config.mjs"

lint-servers: ## Run go vet on all servers
	cd servers/core && go vet ./...
	cd servers/assessment && go vet ./...
	cd servers/interview && go vet ./...
	cd servers/team_allocation && go vet ./...
	cd servers/self_team_allocation && go vet ./...
	cd servers/template_server && go vet ./...
	cd servers/certificate && go vet ./...

# ─── Testing ───────────────────────────────────────────────────────────────────

test: test-core test-assessment test-interview test-team-allocation test-self-team-allocation test-template test-certificate ## Run all server tests

test-core: ## Run core server tests
	cd servers/core && go test ./...

test-assessment: ## Run assessment server tests
	cd servers/assessment && go test ./...

test-interview: ## Run interview server tests
	cd servers/interview && go test ./...

test-team-allocation: ## Run team allocation server tests
	cd servers/team_allocation && go test ./...

test-self-team-allocation: ## Run self team allocation server tests
	cd servers/self_team_allocation && go test ./...

test-template: ## Run template server tests
	cd servers/template_server && go test ./...

test-certificate: ## Run certificate server tests
	cd servers/certificate && go test ./...

# ─── Code Generation ──────────────────────────────────────────────────────────

sqlc: sqlc-core sqlc-assessment sqlc-interview sqlc-team-allocation sqlc-self-team-allocation sqlc-template sqlc-certificate ## Generate sqlc code for all servers

sqlc-core: ## Generate sqlc code for core server
	cd servers/core && sqlc generate

sqlc-assessment: ## Generate sqlc code for assessment server
	cd servers/assessment && sqlc generate

sqlc-interview: ## Generate sqlc code for interview server
	cd servers/interview && sqlc generate

sqlc-team-allocation: ## Generate sqlc code for team allocation server
	cd servers/team_allocation && sqlc generate

sqlc-self-team-allocation: ## Generate sqlc code for self team allocation server
	cd servers/self_team_allocation && sqlc generate

sqlc-template: ## Generate sqlc code for template server
	cd servers/template_server && sqlc generate

sqlc-certificate: ## Generate sqlc code for certificate server
	cd servers/certificate && sqlc generate

swagger: ## Generate swagger docs for core server
	cd servers/core && swag init

# ─── Setup ─────────────────────────────────────────────────────────────────────

install-clients: ## Install client dependencies
	cd clients && yarn install

install-hooks: ## Install git hooks
	./scripts/install-githooks.sh
