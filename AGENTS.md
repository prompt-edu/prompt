# PROMPT 2.0 - AI Assistant Guide

This document provides essential **general** context for AI assistants and contributors working on
the PROMPT 2.0 codebase. Language- and path-specific conventions live in `.claude/rules/`
(see [AI Assistant Tooling](#ai-assistant-tooling)).

## Project Overview

**PROMPT 2.0** is a modular course management platform for project-based university teaching, originally developed for the iPraktikum at TU Munich. It uses a **micro-frontend + microservices architecture** with:

- **Core System**: React frontend (Module Federation) + Go backend (Gin framework)
- **Course Phase Modules**: Independent frontend components and backend services dynamically loaded based on course configuration
- **Authentication**: Keycloak for identity management with RBAC
- **Object Storage**: SeaweedFS (S3-compatible) for uploaded files, served through presigned URLs

**Live Instance:** <https://prompt.aet.cit.tum.de/>

**Community & Support:** [Join our Discord Server](https://discord.gg/eybNUqD8gf) for coordination and support.

## Repository Structure

```text
clients/
  core/                    # Main React app shell (port 3000)
  *_component/             # Course phase micro-frontends:
    - example_component (port 3001)
    - interview_component (port 3002)
    - matching_component (port 3003)
    - assessment_component (port 3007)
    - team_allocation_component (port 3008)
    - self_team_allocation_component (port 3009)
    - certificate_component (port 3010)
  external remotes:
    - intro_course_developer_component (served by prompt-intro-course, port 3005 in local dev)
    - github_challenge_component (served by prompt-github-challenge, port 3006 in local dev)
  shared/                  # Module Federation scaffolding shared by the remotes (not a workspace)

servers/
  core/                    # Main Go service (port 8080)
  team_allocation/         # Team matching (port 8083)
  self_team_allocation/    # Self-managed teams (port 8084)
  assessment/              # Rubric-based grading (port 8085)
  example_server/          # Example phase service (port 8086)
  interview/               # Interview scheduling (port 8087)
  certificate/             # Certificate generation (port 8088)

docs/                      # Docusaurus documentation
```

## Quick Start Commands

Use the Makefile for cross-shell compatible commands:

```bash
# Start all micro-frontends
make clients

# Start core server (loads .env and .env.dev automatically)
make server

# Start every server (core + all microservices)
make servers

# Start database and Keycloak (detached)
make db

# Stop database and Keycloak
make db-down

# Boot the seeded stack for browser verification (see the `verify` skill)
make verify-up
make verify-down

# Run linting
make lint

# Run tests (client unit tests + every server suite)
make test

# Run only the client unit tests
make test-clients

# Regenerate sqlc code for every service (or make sqlc-<service>)
make sqlc

# Regenerate the core server's swagger docs
make swagger

# Install the pre-commit git hooks
make install-hooks

# Regenerate .claude/skills symlinks from .agents/skills
make setup-skills
```

`make help` lists every target, including the per-service `server-*`, `client-*`, `test-*`, and
`sqlc-*` ones.

**Environment Setup:** Copy `.env.template` to `.env` and `.env.dev.template` to `.env.dev`. The `.env.dev` file contains localhost overrides for local development (vs Docker hostnames in `.env`). Each microservice has separate DB configuration (e.g., `DB_CORE_*`, `DB_TEAM_ALLOCATION_*`). For auth/Keycloak setup and common failures, use the `keycloak-local-setup` skill.

## Technology Stack

### Frontend

- React 19, TypeScript 6, rspack 2 (Module Federation)
- Tailwind CSS v4, shadcn/ui + Radix UI
- Zustand (state), TanStack React Query (data fetching)
- React Hook Form, Axios, React Router DOM 7

### Backend

- Go 1.26, Gin framework
- PostgreSQL with pgx driver
- sqlc for type-safe SQL generation
- golang-migrate for migrations

## AI Assistant Tooling

Agent configuration is shared with the team and split by purpose:

- **General guidance** (this file). Read it first. `AGENTS.md` is the cross-tool standard; Claude
  Code reads it via the `@AGENTS.md` import in `CLAUDE.md`.
- **Rules** — language/path-specific conventions in `.claude/rules/<stack>/<facet>.md`. Each stack
  file is path-scoped (`paths:` frontmatter) so Claude auto-loads it only when you touch matching
  files. Stacks: `common/` (always-on), `go/`, `react-typescript/`, `database/`,
  `module-federation/`, `docker/`. Other tools: see `AGENTS.override.md` (Codex).
- **Skills** — repeatable procedures in `.agents/skills/` (canonical source), symlinked into
  `.claude/skills/` via `make setup-skills`.
  - Repo-specific: `new-course-phase`, `sqlc-migration`, `add-shared-ui-component`,
    `module-federation-remote`, `keycloak-local-setup`, `verify`, `open-pr`,
    `address-pr-comments`, `github-release-creation`.
  - Reference patterns (vendored from ECC, MIT): `golang-patterns`, `postgres-patterns`,
    `react-performance`, `docker-patterns`.
- **Subagents** — focused reviewers in `.claude/agents/`: `go-service-reviewer`,
  `frontend-reviewer`, `migration-auditor`.

## Creating New Course Phases

A phase is a micro-frontend (`clients/<name>_component`) plus a Go service (`servers/<name>`), wired
into Module Federation, the workspaces, and `docker-compose.yml`. **Scaffold with
`make new-phase NAME=<name> CLIENT_PORT=<port> SERVER_PORT=<port>`** (the `new-course-phase` skill
and `docs/contributor/new_course_phase.md` cover the remaining manual steps); the
`module-federation-remote` and `sqlc-migration` skills cover the sub-steps. External (out-of-repo)
phases: see the external-phase section of the guide and `template-repository/`.

## Testing

Run `make lint` and `make test` before completing a change. Client unit tests run on Vitest from
`clients/` (`make test-clients`); Go tests use `testcontainers-go`; end-to-end tests use Playwright
and are documented in `e2e/README.md`. Details: `.claude/rules/common/testing.md`.

To observe a change in a real browser rather than through tests, use the `verify` skill:
`make verify-up` boots the same seeded stack in host-browser mode (client
<http://localhost:4000>), which is then driven with the Playwright MCP browser.

## Definition of Done

- `make lint` and `make test` pass.
- Backend: protected routes under `/api/course_phase/:coursePhaseID` with correct roles; `db/sqlc/`
  regenerated and committed with any migration/query change.
- Frontend: no `any`; reuse shared libraries over custom code.
- PR title uses the backtick feature-tag format (`open-pr` skill; `address-pr-comments` for review
  follow-ups).

**End-to-End Tests (core server + client):**

Run **one shard at a time**, never the whole suite:

```bash
make test-e2e-shard                    # list the available shards
make test-e2e-shard SHARD=interview    # run one CI module shard
make test-e2e-shard PATHS="tests/..."  # run a narrower slice
make test-e2e-ui                       # interactive Playwright UI in Docker (open http://127.0.0.1:8123)
make test-e2e-down                     # stop the stack and remove volumes
```

- `make test-e2e` runs every spec in one container. Do not use it: CI splits the suite by
  `e2e/shards.json` (one shard per microservice plus `core`) and never runs it whole either
- Playwright suite lives in `e2e/`; see **`e2e/README.md`** for how to run it
  (including interactive UI mode) and how to add tests
- Boots core server + client + Keycloak + Postgres + SeaweedFS via `docker-compose.e2e.yml`
- Uses the seeded Keycloak users and a fixed DB seed (`e2e/seed/e2e_seed.sql`)
- Runs on non-default host ports (client 4000 / API 18090 / Keycloak 18081), so
  it coexists with a running dev stack

## UI Guidelines

- Student main pages: Place key actions directly, avoid subpage navigation
- Lecturer main pages: State purpose, show status summary with progress indicators
- Recommended subpages: Participants, Student Preview, Mailing (optional), Configuration

## Documentation

- **User/Admin Docs:** `docs/` (Docusaurus) - run with `yarn start`
- **API Docs:** Swagger annotations in Go code (`@Summary`, `@Tags`, etc.)
- **Setup Guide:** `docs/contributor/setup.md`
- **Client Guide:** `docs/contributor/guide/client.md`
- **Server Guide:** `docs/contributor/guide/server.md`

## Important Notes

- All microservices use separate PostgreSQL databases
- Routes must be under `<server>/api/course_phase/:coursePhaseID` for SDK auth
- Shared UI lives in the `prompt-edu/prompt-lib` repository (`@tumaet/prompt-ui-components`), not in
  this one; new primitives are added there and released as a package version
- Course-specific roles are dynamically created with a naming convention including semester and course name
