---
name: new-course-phase
description: Scaffold a new PROMPT 2.0 course phase end-to-end — a micro-frontend (clients/) plus a Go microservice (servers/) wired into Module Federation, the workspaces, and docker-compose. Use when asked to add a new course phase, component, or phase module/service.
---

Scaffold a new course phase. A phase is two parts: a React micro-frontend under `clients/` and a
Go microservice under `servers/`, both derived from the example phase in
[prompt-edu/prompt-example-phase](https://github.com/prompt-edu/prompt-example-phase) (its client
and server each have a README).

## Preferred: use the generator

```bash
make new-phase NAME=<snake_case_name> CLIENT_PORT=<port> SERVER_PORT=<port> [DB_PORT=<port>]
```

`scripts/new-course-phase.sh` clones prompt-example-phase, renames every identifier with its
`init.sh`, converts it to the workspace layout, registers the phase (workspaces, `clients/Dockerfile`,
core remote + phase mappings, docker-compose services, env templates, Makefile targets, CI
matrices), and verifies with `go build`, `yarn install`, `tsc`,
and biome. It needs network access to GitHub; set
`PHASE_TEMPLATE_SOURCE` (a git URL or local clone) and `PHASE_TEMPLATE_REF` to use another
source. Run it on a clean working tree so a failed run can be reverted. Taken client dev
ports: core 3000, interview 3002, matching 3003, intro_course_developer 3005,
github_challenge 3006, assessment 3007, team_allocation 3008, self_team_allocation 3009,
certificate 3010, presentation 3011, infrastructure_setup 3012, ai 3013. Taken server ports: core
8080, team_allocation 8083, self_team_allocation 8084, assessment 8085, interview 8087,
certificate 8088, presentation 8089, infrastructure_setup 8091, ai 8092. Pick free ones; the script aborts
on collisions.

Afterwards, complete the steps the generator prints:

1. **Course phase type** — add an `init<Name>()` in
   `servers/core/coursePhaseType/initializeTypes.go` (name key `<name>_component`, `BaseUrl`
   `{CORE_HOST}/<name>/api` in prod / `devServiceURL(<port>, "/<name>/api")` in dev, never a
   literal `http://localhost:<port>`, so parallel stacks reach their own server; required
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
`rspack.config.mjs` (`createRspackConfig({ name, port })`), `clients/package.json` +
`clients/lerna.json` workspaces, `clients/core/remotes.config.mjs` (the core remote list —
`module-federation-remote` skill), the three `PhaseMapping` files, and the docker-compose blocks in
`scripts/new-course-phase/compose.*.yml.tmpl`.

## Verify

- `cd servers/<name> && go build ./... && go test ./...`; `cd clients && yarn install`;
  `make lint`.
- Start the phase (`make db`, `make server-<name>`, `make clients`) and confirm core lazy-loads
  the remote for a course phase of type `<name>_component`.
- External (out-of-repo) phases: see the external-phase section of
  `docs/contributor/new_course_phase.md`; they use prompt-example-phase as a GitHub template.
