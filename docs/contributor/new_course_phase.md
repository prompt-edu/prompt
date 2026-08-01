---
sidebar_position: 4
---

# How to Create a New Course Phase

A course phase consists of two parts:

- a **React micro-frontend** loaded into the core client via Webpack Module Federation, and
- a **Go microservice** with its own PostgreSQL database, secured through
  [`prompt-sdk`](https://github.com/prompt-edu/prompt-sdk).

There are two ways to build one:

| Model                        | Lives in                | Good for                                        |
| ---------------------------- | ----------------------- | ----------------------------------------------- |
| **In-repo phase**            | this monorepo           | phases maintained by the PROMPT team            |
| **External phase**           | its own repository      | institution- or course-specific custom phases   |

Both start from the same reference implementation: `clients/example_component` and
`servers/example_server` (each has a README explaining its file layout). External phases follow
the model of [prompt-intro-course](https://github.com/prompt-edu/prompt-intro-course) and
[prompt-github-challenge](https://github.com/prompt-edu/prompt-github-challenge); a standalone
GitHub template repository is being prepared (see `template-repository/` in the repo root).

Whichever model you pick, the phase must be **registered in core** (section 2) — that part is
identical for both.

---

## 1. Scaffold the phase

### In-repo: use the generator

```bash
make new-phase NAME=feedback CLIENT_PORT=3011 SERVER_PORT=8089
```

The script (`scripts/new-course-phase.sh`) copies the example component and server, renames every
identifier, and applies all registration steps below except the phase type (section 2.4) and the
deployment workflows (section 3), printing a checklist of what it changed and what remains manual.
Pick ports that are free — the script aborts if they collide with an existing phase.

<details>
<summary>Manual checklist (what the generator automates)</summary>

**Client** (`clients/<name>_component`, name must end in `_component` — `core/src/declaration.d.ts`
types remotes by that suffix):

1. `cp -R clients/example_component clients/<name>_component`; rename the inner
   `src/example_component` directory.
2. `rspack.config.mjs`: set `COMPONENT_NAME` and a unique `COMPONENT_DEV_PORT`.
3. `package.json`: set `"name"`; keep `biome.json` extending the root (`"extends": "//"`).
4. Register the workspace in `clients/package.json` (`workspaces.packages`) and
   `clients/lerna.json` (`packages`), then `cd clients && yarn install`.

**Server** (`servers/<name>`):

1. `cp -R servers/example_server servers/<name>`; rename the `example/` module package.
2. `go.mod`: module path `github.com/prompt-edu/prompt/servers/<name>`; update all imports.
3. `main.go`: route prefix `<name>-service`, env vars `DB_HOST_<NAME>_SERVER` /
   `DB_PORT_<NAME>_SERVER` / `SENTRY_DSN_<NAME>_SERVER`, server port, swagger annotations.
4. Keep the `config/` and `copy/` packages — they implement the SDK phase config/copy contracts
   that course templating and deep copy rely on.
5. Schema in `db/migration/`, queries in `db/query/`, then `sqlc generate` (see the
   [server guide](./guide/server.md)).

**Infrastructure:**

1. `docker-compose.yml`: add `server-<name>`, `client-<name>-component`, and `db-<name>` services
   (copy the `*-example` blocks; separate database per service, own `postgres_<name>_data` volume).
2. `.env.template` + `.env.dev.template`: add the `DB_*_<NAME>_*` entries and `<NAME>_HOST`.
3. `Makefile`: add `server-<name>`, `test-<name>`, `sqlc-<name>` targets and wire them into the
   aggregate `server`, `test`, `sqlc`, and `lint-servers` targets.
4. CI: add the client to the `quality-clients.yml` matrix and the server to the
   `quality-servers.yml` and `test-servers.yml` matrices.

</details>

### External: own repository

Structure your repo like `prompt-github-challenge` (`client/` + `server/` + `docker-compose.yml`
+ own CI); the upcoming template repository gives you this out of the box. Your client is built
and served by your own deployment; your server runs against its own database. You then need one
small PR to this repo for core registration — exactly the steps in section 2, with your deployed
URLs instead of localhost ones.

---

## 2. Register the phase in core

These steps make core aware of your phase. All paths are under `clients/core`.

### 2.1 Module Federation remote

In `clients/core/rspack.config.mjs`:

```js
const feedbackURL = IS_DEV ? `http://localhost:3011` : '/feedback'   // or an env-provided URL
// in ModuleFederationPlugin > remotes — the Date.now() suffix busts the remote-entry cache:
feedback_component: `feedback_component@${feedbackURL}/remoteEntry.js?${Date.now()}`,
```

For external phases the production URL typically comes from an env var exposed via
`core/public/env.template.js` (see 2.5).

### 2.2 External routes and sidebar loaders

Copy `ExampleRoutes.tsx` (`src/managementConsole/PhaseMapping/ExternalRoutes/`) and
`ExampleSidebar.tsx` (`.../ExternalSidebars/`), rename them, and update the import string to your
remote. The import must stay a **static string literal** — webpack needs it for code splitting,
and the copy-per-phase pattern keeps loading/caching behavior correct.

### 2.3 Phase mappings and management route

Map your **course phase type name** (as stored in the DB, see 2.4) in three files under
`src/managementConsole/PhaseMapping/`:

- `PhaseRouterMapping.tsx` → your `<Name>Routes`
- `PhaseSidebarMapping.tsx` → your `<Name>Sidebar`
- `PhaseStudentDetailMapping.tsx` → your student-detail provider (or `Fallback`)

Also add the management route in `src/App.tsx` (copy the `example_component` block):

```tsx
<Route path='/management/course/:courseId/feedback_component/*' ... />
```

### 2.4 Course phase type

Core seeds phase types at startup in `servers/core/coursePhaseType/initializeTypes.go`. Add an
`init<Name>()` following the existing pattern: name (must match the mapping keys from 2.3),
`BaseUrl` (`{CORE_HOST}/<name>/api` in production, `http://localhost:<port>/<name>/api` in dev),
description, and the phase's required inputs / provided outputs. For a purely private deployment
you can instead insert the row into `course_phase_type` manually, but the initializer is the
supported path.

### 2.5 Runtime environment variable (production)

Add `<NAME>_HOST` to `clients/core/public/env.template.js` (substituted by `envsubst` in the
core client's entrypoint) and to `env.js` for local dev. Note: the typed `EnvType` lives in
`@tumaet/prompt-shared-state`; until your var is added there, read it from `window.env` (see
`exampleServerConfig.ts` for the pattern).

---

## 3. Deployment (in-repo phases)

1. `docker-compose.prod.yml`: add the client service with traefik labels (copy the
   `client-example-component` block: router rule `PathPrefix(/<name>)`, strip-prefix and compress
   middlewares) and the server + db services.
2. `build-and-push-clients.yml`: add a build job and an `<name>_image_tag` output.
3. `deploy-docker.yml`: add the image tag env var to the `.env.prod` step and the service to the
   `SERVICES` list; reference the new output in `dev.yml` and `prod.yml`.

External phases handle deployment in their own repository; only the core-side URL (2.1/2.5) and
phase type (2.4) concern this repo.

---

## 4. Verify

- `cd servers/<name> && go build ./... && go test ./...`
- `cd clients && yarn install && yarn tsc -p <name>_component/tsconfig.json --noEmit`
- `make lint`
- Start the stack (`make db`, `make server`, the new server target, `make clients`) and add a
  phase of your new type to a course — core must lazy-load your remote, and the sidebar entry
  must appear with the configured permissions.
