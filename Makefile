.PHONY: help server servers client-core client-certificate client-presentation client-assessment \
	client-interview client-matching clients db db-down \
	server-core server-assessment server-interview \
	server-team-allocation server-self-team-allocation server-example \
	server-certificate server-presentation server-infrastructure-setup \
	lint lint-clients lint-servers \
	test test-clients test-core test-assessment test-interview \
	test-team-allocation test-self-team-allocation test-example \
	test-certificate test-presentation test-infrastructure-setup \
	test-e2e test-e2e-shard test-e2e-ui test-e2e-down \
	verify-up verify-down stack stack-down s3 \
	sqlc sqlc-core sqlc-assessment sqlc-interview \
	sqlc-team-allocation sqlc-self-team-allocation sqlc-example \
	sqlc-certificate sqlc-presentation sqlc-infrastructure-setup \
	swagger install-clients install-hooks setup-skills new-phase \
	seed seed-check

# A linked worktree has no .env or .env.dev of its own (both are gitignored),
# so it falls back to the main checkout's copy of each.
MAIN_CHECKOUT := $(patsubst %/.git,%,$(shell git rev-parse --path-format=absolute --git-common-dir 2>/dev/null))
ENV_FILE := $(firstword $(wildcard .env $(MAIN_CHECKOUT)/.env))
ENV_DEV_FILE := $(firstword $(wildcard .env.dev $(MAIN_CHECKOUT)/.env.dev))

# Load .env file if it exists (base configuration)
ifneq (,$(ENV_FILE))
    include $(ENV_FILE)
    export
endif

# Load .env.dev file if it exists (local development overrides)
ifneq (,$(ENV_DEV_FILE))
    include $(ENV_DEV_FILE)
    export
endif

# ─── Stack ─────────────────────────────────────────────────────────────────────

# Every worktree can run its own stack next to the others. The main checkout is
# stack 0 and keeps the ports and container names in the Compose files. A linked
# worktree gets its own number n on first use (scripts/stack-number.sh), and
# PROMPT_STACK=<n> overrides it. Stack n shifts every published host port by
# n * 100, names its containers prompt-<n>-* and runs under the Compose projects
# prompt-<n> and prompt-<n>-e2e. Docs: docs/contributor/setup.md.
#
# Only the targets that touch a stack ask for a number, so `make lint` or
# `make test` in a fresh worktree never takes one.
STACK_GOALS := stack stack-down server servers server-% clients client-% db db-down s3 \
	seed test-e2e test-e2e-shard test-e2e-down test-e2e-ui verify-up verify-down

ifneq (,$(filter $(STACK_GOALS),$(MAKECMDGOALS)))
STACK := $(shell PROMPT_STACK='$(PROMPT_STACK)' ./scripts/stack-number.sh)
ifeq (,$(STACK))
$(error No stack number for this checkout, see the message above)
endif
export PROMPT_STACK := $(STACK)
STACK_OFFSET := $(shell echo $$(($(STACK) * 100)))

# Every PROMPT_PORT_<port> these files mention is <port> shifted to this stack.
# The Compose files fall back to <port> itself when the variable is unset.
STACK_PORT_FILES := Makefile docker-compose.yml docker-compose.stack.yml \
	docker-compose.e2e.yml e2e/docker-compose.browser.yml
$(foreach port,$(shell grep -ohsE 'PROMPT_PORT_[0-9]+' $(STACK_PORT_FILES) | sort -u | \
	awk -F_ '{ print $$0 "=" $$3 + $(STACK_OFFSET) }'),$(eval export $(port)))

# Read by core to register the phase services of a development stack, and by
# the client dev servers (clients/shared/rspack/devPorts.mjs).
export PROMPT_PORT_OFFSET := $(STACK_OFFSET)
export PROMPT_CLIENT_PORT := $(PROMPT_PORT_3000)

ifeq ($(STACK),0)
STACK_PREFIX := prompt
DEV_COMPOSE := docker compose
else
STACK_PREFIX := prompt-$(STACK)
DEV_COMPOSE := docker compose -p $(STACK_PREFIX) -f docker-compose.yml -f docker-compose.stack.yml
export PROMPT_CONTAINER_PREFIX := $(STACK_PREFIX)

# The host-run servers reach this stack's databases, Keycloak and S3 on its own
# ports. These are the port settings of .env.dev, which stack 0 keeps as is.
export DB_PORT := $(PROMPT_PORT_5432)
export DB_CORE_PORT := $(PROMPT_PORT_5432)
export DB_PORT_TEAM_ALLOCATION := $(PROMPT_PORT_5434)
export DB_PORT_SELF_TEAM_ALLOCATION := $(PROMPT_PORT_5436)
export DB_PORT_ASSESSMENT := $(PROMPT_PORT_5435)
export DB_PORT_CERTIFICATE := $(PROMPT_PORT_5439)
export DB_PORT_EXAMPLE_SERVER := $(PROMPT_PORT_5437)
export DB_PORT_INTERVIEW := $(PROMPT_PORT_5438)
export DB_PORT_PRESENTATION := $(PROMPT_PORT_5440)
export DB_PORT_INFRASTRUCTURE_SETUP := $(PROMPT_PORT_5441)
export KEYCLOAK_HOST := http://localhost:$(PROMPT_PORT_8081)
# The stack's Keycloak imports keycloakConfig.json on its first start, which
# stores this placeholder as the prompt-server secret.
export KEYCLOAK_CLIENT_SECRET := **********
export S3_ENDPOINT := http://localhost:$(PROMPT_PORT_8334)
export S3_PUBLIC_ENDPOINT := http://localhost:$(PROMPT_PORT_8334)
export SERVER_CORE_HOST := http://localhost:$(PROMPT_PORT_8080)
export CORE_HOST := http://localhost:$(PROMPT_PORT_3000)
endif
E2E_PROJECT := $(STACK_PREFIX)-e2e
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
	@$(MAKE) server-example &
	@$(MAKE) server-certificate &
	@$(MAKE) server-presentation &
	@$(MAKE) server-infrastructure-setup &
	@wait
	@echo "All servers started."

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

server-example: ## Start example server (port 8086)
	cd servers/example_server && go run main.go

server-certificate: ## Start certificate server (port 8088)
	cd servers/certificate && go run main.go

server-presentation: ## Start presentation server (port 8089)
	cd servers/presentation && go run main.go

server-infrastructure-setup: ## Start infrastructure setup server (port 8091)
	cd servers/infrastructure_setup && go run main.go

clients: ## Start all client micro-frontends
	cd clients && yarn install && yarn run dev

client-core: ## Start only the core client
	cd clients/core && yarn dev

client-certificate: ## Start only the certificate client
	cd clients/certificate_component && yarn dev

client-presentation: ## Start only the presentation client
	cd clients/presentation_component && yarn dev

client-assessment: ## Start only the assessment client
	cd clients/assessment_component && yarn dev

client-interview: ## Start only the interview client
	cd clients/interview_component && yarn dev

client-matching: ## Start only the matching client
	cd clients/matching_component && yarn dev

DB_SERVICES = db db-team-allocation db-assessment db-self-team-allocation \
	db-example-server db-interview db-certificate db-presentation \
	db-infrastructure-setup

db: ## Start every service database and Keycloak
	docker compose up -d $(DB_SERVICES) keycloak

db-down: ## Stop every service database and Keycloak
	docker compose stop $(DB_SERVICES) keycloak

# Runs in a container so no host psql is needed and the code path matches the
# one the e2e stack uses. The servers own the schemas, so start them once first.
seed: ## Load the demo course into every service database (re-runnable)
	docker compose run --rm seed

seed-check: ## Verify the cross-database references in seed/ resolve
	./scripts/seed-check.sh

stack: ## Show this checkout's stack number, ports and the allocated stacks
	@echo "stack       $(STACK) (containers $(STACK_PREFIX)-*, e2e containers $(STACK_PREFIX)-e2e-*)"
	@echo "client      http://localhost:$(PROMPT_PORT_3000) (remotes up to $(PROMPT_PORT_3012))"
	@echo "core API    http://localhost:$(PROMPT_PORT_8080)"
	@echo "servers     team allocation $(PROMPT_PORT_8083), self team allocation $(PROMPT_PORT_8084), assessment $(PROMPT_PORT_8085),"
	@echo "            example $(PROMPT_PORT_8086), interview $(PROMPT_PORT_8087), certificate $(PROMPT_PORT_8088),"
	@echo "            presentation $(PROMPT_PORT_8089), infrastructure setup $(PROMPT_PORT_8091)"
	@echo "keycloak    http://localhost:$(PROMPT_PORT_8081)"
	@echo "postgres    core $(PROMPT_PORT_5432), keycloak $(PROMPT_PORT_5431), phase services $(PROMPT_PORT_5434)-$(PROMPT_PORT_5441)"
	@echo "s3          http://localhost:$(PROMPT_PORT_8334)"
	@echo "e2e/verify  client $(PROMPT_PORT_4000), core API $(PROMPT_PORT_18090), keycloak $(PROMPT_PORT_18081), mailpit $(PROMPT_PORT_18025), s3 $(PROMPT_PORT_18333)"
	@echo "allocated   (<git common dir>/prompt-stacks)"
	@./scripts/stack-number.sh --list | sed 's/^/            /'

# ─── Code Quality ──────────────────────────────────────────────────────────────

lint: lint-clients lint-servers ## Lint all code

lint-clients: ## Lint all clients
	cd clients && yarn biome check
	./scripts/check-remote-styles.sh

lint-servers: ## Run go vet on all servers
	cd servers/core && go vet ./...
	cd servers/assessment && go vet ./...
	cd servers/interview && go vet ./...
	cd servers/team_allocation && go vet ./...
	cd servers/self_team_allocation && go vet ./...
	cd servers/example_server && go vet ./...
	cd servers/certificate && go vet ./...
	cd servers/presentation && go vet ./...
	cd servers/infrastructure_setup && go vet ./...

# ─── Testing ───────────────────────────────────────────────────────────────────

test: test-clients test-core test-assessment test-interview test-team-allocation test-self-team-allocation test-example test-certificate test-presentation test-infrastructure-setup ## Run all client and server tests

test-clients: ## Run all client unit tests
	cd clients && yarn install && yarn test

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

test-example: ## Run example server tests
	cd servers/example_server && go test ./...

test-certificate: ## Run certificate server tests
	cd servers/certificate && go test ./...

test-presentation: ## Run presentation server tests
	cd servers/presentation && go test ./...

test-infrastructure-setup: ## Run infrastructure setup server tests
	cd servers/infrastructure_setup && go test ./...

# ─── End-to-End Tests ──────────────────────────────────────────────────────────

E2E_COMPOSE = docker compose -f docker-compose.e2e.yml --env-file e2e/.env.e2e

# The Makefile exports .env / .env.dev (host-mode overrides like DB_CORE_HOST=localhost),
# which Compose ranks above --env-file. Unset the keys e2e/.env.e2e defines so it stays authoritative.
E2E_ENV_KEYS := $(shell sed -nE 's/^([A-Za-z_][A-Za-z0-9_]*)=.*/\1/p' e2e/.env.e2e)

# Set SKIP_BUILD=1 to reuse the images already on the machine. Only safe when
# nothing under clients/, servers/, or e2e/ changed since the last build.
E2E_BUILD = $(if $(SKIP_BUILD),true,$(E2E_COMPOSE) build)

test-e2e: ## Whole suite in one container - CI-only, use test-e2e-shard SHARD=<name> locally
	@mkdir -p e2e/playwright-report e2e/test-results e2e/blob-report
	unset $(E2E_ENV_KEYS); \
		$(E2E_BUILD); \
		$(E2E_COMPOSE) run --rm e2e-runner; status=$$?; \
		$(E2E_COMPOSE) down -v; \
		exit $$status

test-e2e-shard: ## Run one CI module shard locally, e.g. make test-e2e-shard SHARD=interview (SKIP_BUILD=1 reuses existing images)
	@mkdir -p e2e/playwright-report e2e/test-results e2e/blob-report
	@if [ -z "$(SHARD)$(PATHS)" ]; then \
		echo "make test-e2e-shard: set SHARD=<name> (or PATHS=\"<patterns>\"), else this would run the whole suite."; \
		echo "Available shards:"; cd e2e && node scripts/shards.mjs names | sed 's/^/  /'; \
		exit 2; \
	fi
	@set -e; \
		paths="$(PATHS)"; \
		if [ -z "$$paths" ]; then paths="$$(cd e2e && node scripts/shards.mjs paths "$(SHARD)")"; fi; \
		set -f; \
		unset $(E2E_ENV_KEYS); \
		$(E2E_BUILD); \
		set +e; \
		$(E2E_COMPOSE) run --rm e2e-runner npx playwright test $$paths; status=$$?; \
		$(E2E_COMPOSE) down -v; \
		exit $$status

test-e2e-down: ## Tear down the e2e stack and remove volumes
	unset $(E2E_ENV_KEYS); $(E2E_COMPOSE) down -v

test-e2e-ui: ## Interactive Playwright UI in Docker - then open http://127.0.0.1:8123
	@mkdir -p e2e/playwright-report e2e/test-results
	@echo "Starting the interactive Playwright UI - open http://127.0.0.1:8123 once it prints 'Listening'."
	unset $(E2E_ENV_KEYS); \
		$(E2E_COMPOSE) run --rm --build -p 8123:8123 \
		-v "$(CURDIR)/e2e/tests:/work/tests" \
		-v "$(CURDIR)/e2e/src:/work/src" \
		-v "$(CURDIR)/e2e/playwright-report:/work/playwright-report" \
		-v "$(CURDIR)/e2e/test-results:/work/test-results" \
		e2e-runner npx playwright test --ui-host=0.0.0.0 --ui-port=8123; \
		$(E2E_COMPOSE) down -v

# ─── Browser Verification ─────────────────────────────────────────────────────

# Same seeded stack as the e2e suite, overlaid so the client, Keycloak and the
# core API are reachable as localhost:<published port>. That lets a browser on
# the host complete the login redirect and mint a token the servers accept.
# Driven by the `verify` skill.
VERIFY_COMPOSE = docker compose -f docker-compose.e2e.yml -f e2e/docker-compose.browser.yml --env-file e2e/.env.e2e

verify-up: ## Boot the seeded stack for host-browser verification (SKIP_BUILD=1 reuses existing images)
	set -e; unset $(E2E_ENV_KEYS); \
		$(if $(SKIP_BUILD),true,$(VERIFY_COMPOSE) build); \
		$(VERIFY_COMPOSE) up -d client-core server-core seed
	@echo ""
	@echo "client    http://localhost:4000/management"
	@echo "core API  http://localhost:18090"
	@echo "keycloak  http://localhost:18081"
	@echo "mailpit   http://localhost:18025"
	@echo "logins    username == password (lecturer, admin, student, ...)"

verify-down: ## Tear down the host-browser stack and remove volumes
	unset $(E2E_ENV_KEYS); $(VERIFY_COMPOSE) down -v

# ─── Code Generation ──────────────────────────────────────────────────────────

sqlc: sqlc-core sqlc-assessment sqlc-interview sqlc-team-allocation sqlc-self-team-allocation sqlc-example sqlc-certificate sqlc-presentation sqlc-infrastructure-setup ## Generate sqlc code for all servers

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

sqlc-example: ## Generate sqlc code for example server
	cd servers/example_server && sqlc generate

sqlc-certificate: ## Generate sqlc code for certificate server
	cd servers/certificate && sqlc generate

sqlc-presentation: ## Generate sqlc code for presentation server
	cd servers/presentation && sqlc generate

sqlc-infrastructure-setup: ## Generate sqlc code for infrastructure setup server
	cd servers/infrastructure_setup && sqlc generate

swagger: ## Generate swagger docs for core server
	cd servers/core && swag init

# ─── Setup ─────────────────────────────────────────────────────────────────────

install-clients: ## Install client dependencies
	cd clients && yarn install

install-hooks: ## Install git hooks (via the pre-commit framework)
	git config --local --unset-all core.hooksPath 2>/dev/null || true
	pre-commit install

setup-skills: ## Regenerate .claude/skills symlinks from .agents/skills (canonical source)
	./scripts/setup-skills.sh

# DB_PORT is also defined in .env (included above), so only forward it when it
# was given on the command line; otherwise the script picks the next free port.
new-phase: ## Scaffold a new course phase (make new-phase NAME=<name> CLIENT_PORT=<port> SERVER_PORT=<port> [DB_PORT=<port>])
	./scripts/new-course-phase.sh "$(NAME)" "$(CLIENT_PORT)" "$(SERVER_PORT)" "$(if $(filter command line,$(origin DB_PORT)),$(DB_PORT),)"
