---
name: new-course-phase
description: Scaffold a new PROMPT 2.0 course phase end-to-end — a micro-frontend (clients/) plus a Go microservice (servers/) wired into Module Federation, the workspaces, and docker-compose. Use when asked to add a new course phase, component, or phase module/service.
---

Scaffold a new course phase. A phase is two parts: a React micro-frontend under `clients/` and a
Go microservice under `servers/`. Pick a `snake_case` phase name (e.g. `feedback`), a unique client
dev port, and a unique server port. Work through both checklists; do not skip the registration steps
— a phase that isn't registered in core + compose will not load.

## Frontend component

1. Copy the example: `cp -R clients/example_component clients/<name>_component`.
2. In `clients/<name>_component/webpack.config.mjs` update the two constants near the top:
   `const COMPONENT_NAME = '<name>_component'` and `const COMPONENT_DEV_PORT = <unique-port>`
   (existing ports: core 3000, example 3001, interview 3002, matching 3003, assessment 3007,
   team_allocation 3008, self_team_allocation 3009 — pick a free one).
3. Set `"name": "<name>_component"` in `clients/<name>_component/package.json`.
4. Register the workspace in BOTH `clients/lerna.json` (`packages`) and `clients/package.json`
   (`workspaces.packages`).
5. Register the remote in `clients/core/webpack.config.mjs` `remotes:` using the cache-busting
   pattern (see the `module-federation-remote` skill for the exact line and the matching URL env var).
6. Load it lazily in core: `const X = React.lazy(() => import('<name>_component/App'))`.
7. `cd clients && yarn install` so the workspace resolves.

## Backend service

1. Copy the example: `cp -R servers/example_server servers/<name>`.
2. In `servers/<name>/go.mod` set the module path to
   `github.com/prompt-edu/prompt/servers/<name>` and update internal imports accordingly.
3. Define schema + queries (use the `sqlc-migration` skill): migrations in `db/migration/` as
   `0001_<desc>.up.sql`, queries in `db/query/*.sql`, then run `sqlc generate` (config in
   `servers/<name>/sqlc.yaml`).
4. Implement business logic in `<module>/service.go`, routes in `<module>/router.go` under
   `/api/course_phase/:coursePhaseID` (required for prompt-sdk auth — see the `go/auth-routing` rule),
   validation in `<module>/validation.go`.
5. Keep the `config/` and `copy/` packages from the example (phase config + copy endpoints) and wire
   them in `main.go` (`InitAuthenticationMiddleware`, `RegisterConfigEndpoint`, `RegisterCopyEndpoint`).

## docker-compose wiring

1. Add a `server-<name>` service (copy the `server-example` block): build context
   `./servers/<name>`, `dockerfile: ../Dockerfile`, unique published port mapped to `8080`,
   `depends_on` its db + keycloak (`condition: service_healthy`), and the `DB_*`/`KEYCLOAK_*`/
   `CORE_HOST`/`SERVER_CORE_HOST` env vars (rename `*_EXAMPLE_*` → `*_<NAME>_*`).
2. Add a `client-<name>-component` service (copy `client-example-component`).
3. Add a dedicated database service (copy `db-example-server` → `db-<name>`) with its own
   `postgres_<name>_data` volume. Each microservice uses a SEPARATE Postgres database.
4. Add the matching `DB_*_<NAME>_*` entries to `.env.template` and `.env.dev.template`.

## Verify

- `make setup-skills` not needed here. Run `cd servers/<name> && go build ./...` and
  `go test ./...`; `cd clients && yarn install`; `make lint`.
- Start the phase (`make clients`, the new server target) and confirm core lazy-loads the remote.
- Source of truth for these conventions: `AGENTS.md` ("Creating New Course Phases") and the
  `.claude/rules/` stack files.
