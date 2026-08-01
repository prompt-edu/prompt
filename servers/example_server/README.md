# Example Server

This is the reference Go microservice for a PROMPT course phase. Together with
[`clients/example_component`](../../clients/example_component) it forms a complete, minimal
course phase: copy both, rename, and extend.

> **Fastest start:** run `make new-phase NAME=<name> CLIENT_PORT=<port> SERVER_PORT=<port>` from
> the repository root — it copies both parts and performs every registration step automatically.

## How a phase server plugs into PROMPT

Every phase service is an independent Gin server with its **own PostgreSQL database** (never query
another service's DB). Auth, CORS, and inter-service communication come from
[`prompt-sdk`](https://github.com/prompt-edu/prompt-sdk). Two rules are load-bearing:

- Course-data routes MUST live under `<prefix>/api/course_phase/:coursePhaseID` — the SDK auth
  middleware depends on this path shape.
- Protect every route with `promptSDK.AuthenticationMiddleware(<roles...>)`.

## File map

```text
main.go              Wiring: env config, migrations, DB pool, Keycloak auth, route groups,
                     service info endpoint (capabilities), Sentry (optional)
example/             The actual phase module — the part you replace with your logic
  main.go            InitExampleModule: registers routes, creates the service singleton
  router.go          Gin routes under /api/course_phase/:coursePhaseID with role middleware
  service.go         Business logic (keep handlers thin, logic here)
config/              Phase config endpoint (PhaseConfigHandler) — lets core ask "is this
                     phase fully configured?"; placeholder returns 404 until implemented
copy/                Phase copy endpoint (PhaseCopyHandler) — called on course deep copy and
                     course templating to duplicate phase data; placeholder returns 404
db/
  migration/         golang-migrate SQL (0001_schema.up.sql creates the demo example_table)
  query/             sqlc query definitions
  sqlc/              GENERATED type-safe Go — never edit, run `make sqlc-example`
docs/                GENERATED swagger (swag) — regenerate after annotation changes
utils/getQueries.go  Test helper providing db.Queries
sqlc.yaml            sqlc config (postgresql, pgx/v5, output db/sqlc)
```

## What the code demonstrates

- **SDK auth wiring** — `initKeycloak` + `promptSDK.InitAuthenticationMiddleware` once in
  `main.go`, then per-route `promptSDK.AuthenticationMiddleware(promptSDK.PromptAdmin, ...)`.
- **Module layout** — `example/` follows the repo convention `main.go` (init) / `router.go`
  (thin handlers) / `service.go` (logic); add `validation.go` for input validation.
- **Config & copy endpoints** — `config/` and `copy/` implement the SDK's `PhaseConfigHandler`
  and `PhaseCopyHandler` contracts. Keep both packages in your phase; core relies on the
  capabilities advertised via `RegisterInfoEndpoint`.
- **Database access** — migrations run on startup; queries go through generated sqlc methods.

## Run it locally

```bash
make db              # starts Postgres (this service's DB on port 5437) + Keycloak
make server-example  # runs this server on port 8086
curl http://localhost:8086/example-service/api/course_phase/<uuid>/hello
```

Tests: `make test-example`. Regenerate sqlc: `make sqlc-example`.

## Creating your own phase from this server

Prefer `make new-phase` (see above). To do it manually:

1. `cp -R servers/example_server servers/<name>`, rename the `example/` package, and set the
   module path in `go.mod` to `github.com/prompt-edu/prompt/servers/<name>`.
2. Update the route prefix (`example-service` → `<name>-service`), env var names
   (`*_EXAMPLE_*` → `*_<NAME>_*`), and swagger annotations.
3. Write your schema in `db/migration/`, queries in `db/query/`, run `sqlc generate`.
4. Register the compose services, env vars, Makefile targets, and CI — the full checklist lives
   in `docs/contributor/new_course_phase.md`.

Keep this server generic: it is the source for new phases and is planned to move into a
standalone GitHub template repository (see `template-repository/` at the repo root).
