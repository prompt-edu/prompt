# PROMPT 2.0 - AI Assistant Guide

This document provides essential **general** context for AI assistants and contributors working on
the PROMPT 2.0 codebase. Language- and path-specific conventions live in `.claude/rules/`
(see [AI Assistant Tooling](#ai-assistant-tooling)).

## Project Overview

**PROMPT 2.0** is a modular course management platform for project-based university teaching, originally developed for the iPraktikum at TU Munich. It uses a **micro-frontend + microservices architecture** with:

- **Core System**: React frontend (Webpack Module Federation) + Go backend (Gin framework)
- **Course Phase Modules**: Independent frontend components and backend services dynamically loaded based on course configuration
- **Authentication**: Keycloak for identity management with RBAC

**Live Instance:** <https://prompt.aet.cit.tum.de/>

**Community & Support:** [Join our Discord Server](https://discord.gg/eybNUqD8gf) for coordination and support.

## Repository Structure

```text
clients/
  core/                    # Main React app shell (port 3000)
  shared_library/          # Shared UI components, hooks, network utilities
  *_component/             # Course phase micro-frontends:
    - interview_component (port 3002)
    - matching_component (port 3003)
    - example_component (port 3001)
    - team_allocation_component (port 3008)
    - self_team_allocation_component (port 3009)
    - assessment_component (port 3007)
  external remotes:
    - intro_course_developer_component (served by prompt-intro-course, typically port 3005 in local dev)
    - devops_challenge_component (served by prompt-github-challenge)

servers/
  core/                    # Main Go service (port 8080)
  interview/               # Interview scheduling (port 8087)
  team_allocation/         # Team matching (port 8083)
  self_team_allocation/    # Self-managed teams (port 8084)
  assessment/              # Rubric-based grading (port 8085)
  example_server/          # Example phase service (port 8086)

docs/                      # Docusaurus documentation
```

## Quick Start Commands

Use the Makefile for cross-shell compatible commands:

```bash
# Start all micro-frontends
make clients

# Start core server (loads .env and .env.dev automatically)
make server

# Start database and Keycloak (detached)
make db

# Stop database and Keycloak
make db-down

# Run linting
make lint

# Run tests
make test

# Regenerate .claude/skills symlinks from .agents/skills
make setup-skills
```

**Environment Setup:** Copy `.env.template` to `.env` and `.env.dev.template` to `.env.dev`. The `.env.dev` file contains localhost overrides for local development (vs Docker hostnames in `.env`). Each microservice has separate DB configuration (e.g., `DB_CORE_*`, `DB_TEAM_ALLOCATION_*`). For auth/Keycloak setup and common failures, use the `keycloak-local-setup` skill.

## Technology Stack

### Frontend

- React 19, TypeScript 5.9, Webpack 5 (Module Federation)
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
    `module-federation-remote`, `keycloak-local-setup`, `open-pr`, `address-pr-comments`,
    `github-release-creation`.
  - Reference patterns (vendored from ECC, MIT): `golang-patterns`, `postgres-patterns`,
    `react-performance`, `docker-patterns`.
- **Subagents** — focused reviewers in `.claude/agents/`: `go-service-reviewer`,
  `frontend-reviewer`, `migration-auditor`.
- Full rollout rationale and community-skill references: `AI_TOOLING_PLAN.md`.

## Creating New Course Phases

A phase is a micro-frontend (`clients/<name>_component`) plus a Go service (`servers/<name>`), wired
into Module Federation, the workspaces, and `docker-compose.yml`. **Use the `new-course-phase`
skill** for the full end-to-end checklist; the `module-federation-remote` and `sqlc-migration`
skills cover the sub-steps.

## Testing

Run `make lint` and `make test` before completing a change. Go tests use `testcontainers-go`;
end-to-end tests use Playwright (`e2e-testing` skill). Details: `.claude/rules/common/testing.md`.

## Definition of Done

- `make lint` and `make test` pass.
- Backend: protected routes under `/api/course_phase/:coursePhaseID` with correct roles; `db/sqlc/`
  regenerated and committed with any migration/query change.
- Frontend: no `any`; reuse shared libraries over custom code.
- PR title uses the backtick feature-tag format (`open-pr` skill; `address-pr-comments` for review
  follow-ups).

**End-to-End Tests (core server + client):**

```bash
make test-e2e          # full Dockerized stack + Playwright runner (CI-identical)
make test-e2e-ui       # interactive Playwright UI in Docker (open http://127.0.0.1:8123)
make test-e2e-down     # stop the stack and remove volumes
```

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
- Use `yarn dlx shadcn add <component>` in `shared_library` for new UI components
- Course-specific roles are dynamically created with a naming convention including semester and course name
