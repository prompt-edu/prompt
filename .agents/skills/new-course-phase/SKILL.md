---
name: new-course-phase
description: Scaffold a new PROMPT 2.0 course phase end-to-end — a micro-frontend (clients/) plus a Go microservice (servers/) wired into Module Federation, the workspaces, and docker-compose. Use when asked to add a new course phase, component, or phase module/service.
---

Scaffold a new course phase. A phase is two parts: a React micro-frontend under `clients/` and a
Go microservice under `servers/`, both derived from the example phase
(`clients/example_component` + `servers/example_server` — each has a README).

## Preferred: use the generator

```bash
make new-phase NAME=<snake_case_name> CLIENT_PORT=<port> SERVER_PORT=<port> [DB_PORT=<port>]
```

`scripts/new-course-phase.sh` copies both parts, renames every identifier, registers the phase
(workspaces, core remote + phase mappings + `App.tsx` route, docker-compose services, env
templates, Makefile targets, CI matrices), and verifies with `go build`, `yarn install`, `tsc`,
and biome. Run it on a clean working tree so a failed run can be reverted. Existing client dev
ports: core 3000, example 3001, interview 3002, matching 3003, assessment 3007,
team_allocation 3008, self_team_allocation 3009 — pick a free one; the script aborts on
collisions.

Afterwards, complete the steps the generator prints:

1. **Course phase type** — add an `init<Name>()` in
   `servers/core/coursePhaseType/initializeTypes.go` (name key `<name>_component`, `BaseUrl`
   `{CORE_HOST}/<name>/api` in prod / `http://localhost:<port>/<name>/api` in dev, required
   inputs / provided outputs).
2. **Schema + logic** — migrations in `db/migration/` as `0001_<desc>.up.sql`, queries in
   `db/query/*.sql`, then `make sqlc-<name>` (use the `sqlc-migration` skill). Business logic in
   `<module>/service.go`, routes in `<module>/router.go` under
   `/api/course_phase/:coursePhaseID` (required for prompt-sdk auth — `go/auth-routing` rule).
   Keep the `config/` and `copy/` packages (phase config + copy endpoints).
3. **Deployment** — `docker-compose.prod.yml` service + traefik labels, and the
   `build-and-push-clients.yml` / `deploy-docker.yml` / `dev.yml` / `prod.yml` wiring.

## Manual fallback / verification checklist

If the generator can't be used, follow the manual checklist in
`docs/contributor/new_course_phase.md` (it mirrors what the script does). Key files:
`rspack.config.mjs` (`COMPONENT_NAME`, `COMPONENT_DEV_PORT`), `clients/package.json` +
`clients/lerna.json` workspaces, `clients/core/rspack.config.mjs` remotes (cache-busting pattern
— `module-federation-remote` skill), the three `PhaseMapping` files, and the `*-example`
docker-compose blocks.

## Verify

- `cd servers/<name> && go build ./... && go test ./...`; `cd clients && yarn install`;
  `make lint`.
- Start the phase (`make db`, `make server-<name>`, `make clients`) and confirm core lazy-loads
  the remote for a course phase of type `<name>_component`.
- External (out-of-repo) phases: see the external-phase section of
  `docs/contributor/new_course_phase.md` and the `template-repository/` staging directory.
