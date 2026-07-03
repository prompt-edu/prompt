---
name: docker-patterns
description: Docker Compose patterns for PROMPT's local multi-service stack — networking, volumes, container security, and debugging. Use when editing docker-compose*.yml, adding a service, or troubleshooting containers.
metadata:
  origin: ECC
---

<!--
Vendored from Everything Claude Code (ECC) — https://github.com/affaan-m/everything-claude-code
License: MIT (Copyright 2026 Affaan Mustafa). Adapted for PROMPT 2.0 — the Node/npm/Redis
sample stack and Dockerfiles were replaced with pointers to our real compose files, keeping
only the stack-agnostic networking, volume, security, and debugging guidance.
-->

# Docker Patterns

Docker Compose practices for PROMPT's local stack. PROMPT runs on **Compose v2** (`docker compose`,
not `docker-compose`) with per-phase Go services, React micro-frontends, a **separate PostgreSQL
per microservice**, and Keycloak.

For the concrete service layout — build contexts, the shared `servers/../Dockerfile`, per-phase
`server-<name>` / `client-<name>-component` / `db-<name>` blocks, and the
`.yml` / `.minimal.yml` / `.prod.yml` / `.extern.prod.yml` variants — see the **`docker/compose.md`
rule** and the **`new-course-phase`** skill. This skill covers the stack-agnostic mechanics.

## When to Activate

- Editing `docker-compose*.yml` or adding a service
- Troubleshooting container networking, volumes, or startup ordering
- Reviewing a Compose change for security or reproducibility

## Networking

Services on the same Compose network resolve each other by **service name** — this is why the
`DB_*_HOST` and `KEYCLOAK_*` values in `.env` use container hostnames while `.env.dev` overrides
them to `localhost` for host-side runs:

```txt
# From a server container, the db and keycloak resolve by service name:
db-team-allocation:5432
keycloak:8080
```

Restrict exposure to what the host actually needs:

```yaml
services:
  db-team-allocation:
    ports:
      - "127.0.0.1:5433:5432"   # Reachable from the host only, not the wider network
    # Omit `ports` entirely for services only other containers talk to.
```

## Startup Ordering

A service that needs its database or Keycloak up first must wait on a **healthcheck**, not just
`depends_on` (which only waits for the container to start, not to be ready):

```yaml
services:
  server-team-allocation:
    depends_on:
      db-team-allocation:
        condition: service_healthy
      keycloak:
        condition: service_healthy

  db-team-allocation:
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 3s
      retries: 5
```

## Volumes

Each database gets its **own named volume** so data survives restarts and services never share
state:

```yaml
services:
  db-team-allocation:
    volumes:
      - postgres_team_allocation_data:/var/lib/postgresql/data

volumes:
  postgres_team_allocation_data:   # one dedicated volume per microservice db
```

- **Named volume** (above): persistent, Docker-managed — use for database data.
- **Bind mount** (`./path:/in/container`): maps a host directory in — use for init SQL or local
  config, not for production data.
- `docker compose down -v` **deletes** these volumes (all local DB data). `down` without `-v` keeps
  them.

## Container Security

```dockerfile
# Pin a specific tag — never :latest — for reproducible builds
FROM golang:1.26-alpine3.20

# Run as a non-root user
RUN addgroup -S app && adduser -S app -G app
USER app
```

```yaml
services:
  server-team-allocation:
    security_opt:
      - no-new-privileges:true
    cap_drop:
      - ALL
```

**Secrets** come from the environment, never image layers or committed files:

```yaml
services:
  server-team-allocation:
    env_file:
      - .env          # gitignored — see security.md and the .gitignore AI-config block
    environment:
      - DB_TEAM_ALLOCATION_PASSWORD   # inherited from the host / .env, not hardcoded
```

Never bake a secret into the image (`ENV API_KEY=...`) or commit it to `docker-compose.yml`.

## Debugging

```bash
docker compose logs -f server-team-allocation   # follow one service's logs
docker compose ps                                # what's running / health status
docker compose exec server-team-allocation sh    # shell into a container
docker compose exec db-team-allocation psql -U postgres   # open psql in the db container
docker compose up --build                        # rebuild after code/Dockerfile changes
docker compose down                              # stop & remove containers (volumes kept)
```

Network issues — verify service-name DNS and reachability from inside a container:

```bash
docker compose exec server-team-allocation nslookup db-team-allocation
docker compose exec server-team-allocation wget -qO- http://keycloak:8080/health
```

## Anti-Patterns

- **`:latest` tags** — unpinned images break reproducibility; pin a version.
- **Data in the container, not a volume** — containers are ephemeral; DB data must live in a named
  volume.
- **Running as root** — create and use a non-root user.
- **Sharing one database across services** — every PROMPT microservice owns its own DB.
- **Secrets in `docker-compose.yml` or image layers** — use `.env` (gitignored) instead.
- **`depends_on` without `condition: service_healthy`** for services that need a ready DB/Keycloak.

## Related

- Rule: `docker/compose.md` — PROMPT's per-phase service layout and compose variants
- Rule: `common/security.md` — never-commit-secrets policy and `${VAR}` interpolation
- Skill: `new-course-phase` — scaffolds the three per-phase services end to end
