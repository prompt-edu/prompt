---
paths:
  - "docker-compose*.yml"
---

# docker-compose

Local orchestration uses Docker Compose **v2** (`docker compose`, not `docker-compose`). Variants:
`docker-compose.yml` (full local), `.minimal.yml`, `.prod.yml`, `.extern.prod.yml`.

## Per-phase services

Each course phase contributes three services (mirror the `*-template` blocks):

- `server-<name>` — build context `./servers/<name>`, `dockerfile: ../Dockerfile`, unique published
  port → `8080`, `depends_on` its db + keycloak with `condition: service_healthy`, and
  `CORE_HOST` / `SERVER_CORE_HOST` / `DB_*_<NAME>_*` / `KEYCLOAK_*` / `DEBUG` env vars.
- `client-<name>-component` — build context `./clients/<name>_component`.
- `db-<name>` — its OWN PostgreSQL instance + dedicated `postgres_<name>_data` volume. Every
  microservice uses a separate database; never share one.

## Conventions

- Container names follow `prompt-<service>` / `prompt-db-<name>`.
- Add new `DB_*` and `SENTRY_*` vars to `.env.template` and `.env.dev.template`.
- See the `new-course-phase` skill for the end-to-end scaffold.
