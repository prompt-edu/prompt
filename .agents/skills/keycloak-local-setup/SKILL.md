---
name: keycloak-local-setup
description: Get Keycloak auth working for local PROMPT 2.0 development — the .env vs .env.dev split, realm import, and common auth failures (401/403/CORS). Use when setting up local dev, fixing login/token issues, or onboarding a new contributor.
---

PROMPT 2.0 uses Keycloak (RBAC) for auth. Local dev runs Keycloak + Postgres via docker-compose;
the Go services validate tokens through `prompt-sdk`.

## Environment files

- Copy templates: `.env.template` → `.env`, `.env.dev.template` → `.env.dev`.
- `.env` holds Docker-hostname values (service-to-service inside compose); `.env.dev` holds the
  localhost overrides for running a server on the host. Both are loaded by the Makefile (`.env`
  first, then `.env.dev` overrides). Never commit `.env*` (gitignored); they are copied into
  worktrees via `.worktreeinclude`.
- Key vars: `KEYCLOAK_HOST`, `KEYCLOAK_REALM_NAME`, `CORE_HOST` (frontend URL for CORS),
  `SERVER_CORE_HOST` (inter-service base URL).

## Bring up Keycloak + DB

- `make db` starts Postgres + Keycloak (detached); `make db-down` stops them.
- The realm is imported from `keycloakConfig.json` at the repo root on first start. If roles/clients
  look wrong, stop, remove the `keycloak_postgres_data/` volume, and re-import.

## Common failures

- **401 everywhere:** `KEYCLOAK_HOST`/`KEYCLOAK_REALM_NAME` mismatch between the running Keycloak
  and the service env, or a stale realm — re-import.
- **403 on a route:** route not under `/api/course_phase/:coursePhaseID`, or the user lacks the
  required role (`PromptAdmin`, `PromptLecturer`, `CourseLecturer`, `CourseEditor`, `CourseStudent`).
  See the `go/auth-routing` rule.
- **CORS errors:** `CORE_HOST` doesn't match the frontend origin; `prompt-sdk` `CORSMiddleware` uses it.
- **Token works in Docker but not on host (or vice-versa):** you're using the wrong `.env` vs
  `.env.dev` value for `KEYCLOAK_HOST`/`SERVER_CORE_HOST`.

## Verify

- `make db`, then start the core server; log in via the frontend and confirm a protected
  `/api/course_phase/...` route returns 200 with a valid token.
