---
paths:
  - "docker-compose*.yml"
---

# docker-compose

Local orchestration uses Docker Compose **v2** (`docker compose`, not `docker-compose`). Variants:
`docker-compose.yml` (full local), `.minimal.yml`, `.prod.yml`, `.extern.prod.yml`, and `.e2e.yml`
(driven by the `make test-e2e*` targets, on its own host ports so it coexists with the dev stack).

## Per-phase services

Each course phase contributes three services (mirror the `*-example` blocks):

- `server-<name>` — build context `./servers/<name>`, `dockerfile: ../Dockerfile`, unique published
  port → `8080` (in the stack form below), `depends_on` its db + keycloak with
  `condition: service_healthy`, and
  `CORE_HOST` / `SERVER_CORE_HOST` / `DB_*_<NAME>_*` / `KEYCLOAK_*` / `DEBUG` env vars.
- `client-<name>-component` — build context `./clients/<name>_component`.
- `db-<name>` — its OWN PostgreSQL instance + dedicated `postgres_<name>_data` volume. Every
  microservice uses a separate database; never share one. List it in `DB_SERVICES` in the
  `Makefile` so `make db` starts it in every stack.

## Parallel stacks

Every worktree runs its own stack (`docs/contributor/setup.md`, "Run Several Stacks Side by
Side"), so `docker-compose.yml` and `docker-compose.e2e.yml` never hardcode a container name or a
host port:

- Container names: `container_name: ${PROMPT_CONTAINER_PREFIX:-prompt}-<service>` (e.g.
  `${PROMPT_CONTAINER_PREFIX:-prompt}-db-<name>`, `${PROMPT_CONTAINER_PREFIX:-prompt}-e2e-<service>`).
- Published ports: `"${PROMPT_PORT_<port>:-<port>}:<container port>"`, where `<port>` is the
  stack 0 host port. The Makefile shifts every `PROMPT_PORT_<port>` it finds in the Compose files.
- Every database or other stateful service also gets a named volume in `docker-compose.stack.yml`
  (service override plus a top-level `volumes:` entry). Without it, stack n keeps the data in a
  directory inside the worktree that `make stack-down` doesn't remove.
- A new database port also needs its `export DB_PORT_<NAME> := $(PROMPT_PORT_<port>)` line in the
  Makefile's stack section, so the host-run server reaches the stack's database.
- Start services through the `make` targets, never a bare `docker compose` in a worktree: that
  runs with stack 0's names and ports.

## Conventions

- Add new `DB_*` and `SENTRY_*` vars to `.env.template` and `.env.dev.template`.
- See the `new-course-phase` skill for the end-to-end scaffold.
