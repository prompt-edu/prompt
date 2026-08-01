# Migration Runbook: Example Phase → Template Repository

Moves `clients/example_component` and `servers/example_server` out of the monorepo into the new
GitHub template repository, preserving git history. Execute top to bottom; steps 1–5 touch only
the new repository, step 6 is the removal PR in this monorepo.

## 1. Create the repository

```bash
gh repo create prompt-edu/prompt-course-phase-template --public \
  --description "Template for building a custom PROMPT course phase (React micro-frontend + Go service)"
```

## 2. Import the example with history

```bash
git clone https://github.com/prompt-edu/prompt prompt-history && cd prompt-history
# keep only the two example directories, rewritten to client/ and server/
git filter-repo \
  --path clients/example_component --path servers/example_server \
  --path-rename clients/example_component:client \
  --path-rename servers/example_server:server
git remote add template git@github.com:prompt-edu/prompt-course-phase-template.git
git push template main
```

## 3. Add the staged root files

Copy from `template-repository/repo/` in the monorepo (restoring the leading dots):

```text
repo/README.md               -> README.md
repo/gitignore               -> .gitignore
repo/env.template            -> .env.template
repo/docker-compose.yml      -> docker-compose.yml
repo/docker-compose.prod.yml -> docker-compose.prod.yml
repo/client.package.json     -> client/package.json   (replaces the workspace package.json)
repo/client.Dockerfile       -> client/Dockerfile     (replaces the monorepo one)
repo/client.nginx.conf       -> client/nginx.conf
repo/server.Dockerfile       -> server/Dockerfile
repo/workflows/*.yml         -> .github/workflows/
template-repository/init.sh  -> init.sh
```

Then make the client standalone:

- `server/go.mod`: module `github.com/prompt-edu/prompt-course-phase-template/server`; update
  the internal import paths (`servers/example_server/...` → `.../server/...`).
- `client/`: remove `.yarnrc.yml` remnants of the workspace if present; run `yarn install` to
  produce a fresh `yarn.lock`; check `yarn typecheck && yarn build`.
- `cd server && go build ./... && go test ./...`.

## 4. Validate the template flow

1. `docker compose up` against a locally running PROMPT core + Keycloak (`make db`,
   `make server` in the monorepo) — the phase must load once registered in core.
2. On GitHub: Settings → check **Template repository**.
3. Create a scratch repo via "Use this template", run `./init.sh my_phase 3011 8089`, and verify
   build + typecheck pass with the renamed identifiers.

## 5. Register a phase built from the template

Core-side registration steps (one small PR to the monorepo per phase) are documented in
`docs/contributor/new_course_phase.md`, section "Register the phase in core".

## 6. Removal PR in the monorepo

Only after 1–5 are proven. Delete `clients/example_component`, `servers/example_server`, and
`template-repository/`, then unwind the wiring (the file list mirrors PR #1861):

- core: remote in `clients/core/rspack.config.mjs`, `ExampleRoutes`/`ExampleSidebar`, the three
  `PhaseMapping` entries, the `App.tsx` route, `EXAMPLE_HOST` in `public/env.js` +
  `public/env.template.js`
- workspaces: `clients/package.json`, `clients/lerna.json` (+ `yarn install` for the lockfile)
- compose: `server-example`, `client-example-component`, `db-example-server` (all files incl.
  prod + traefik labels), `EXAMPLE_HOST` pass-throughs
- env templates: `DB_EXAMPLE_*`, `EXAMPLE_HOST`, `EXAMPLE_IMAGE_TAG`, `SENTRY_DSN_EXAMPLE_SERVER`
- `Makefile` targets (`server-example`, `test-example`, `sqlc-example`) and aggregates
- CI: `quality-clients.yml` / `quality-servers.yml` / `test-servers.yml` matrices,
  `build-and-push-clients.yml` job + output, `deploy-docker.yml` / `dev.yml` / `prod.yml`
- api-stress catalog + service maps, e2e seed `Example` phase type, `.claude/rules` +
  skill references (point them at the template repository instead)
- `scripts/new-course-phase.sh` / `make new-phase`: either retire it or repoint it at cloning
  the template repository — decide in the removal PR
- ops notes: drop `EXAMPLE_*` env vars from deployed `.env` files, remove the deployed example
  client image/route and `example_component` phase-type rows

Verify with `make lint`, `make test`, a core client build, the e2e suite, and a repo-wide grep
for `example_component|example_server|EXAMPLE_`.
