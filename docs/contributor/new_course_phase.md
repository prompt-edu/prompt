---
sidebar_position: 4
---

# How to Create a New Course Phase

A course phase consists of two parts:

- a **React micro-frontend** loaded into the core client via Module Federation, and
- a **Go microservice** with its own PostgreSQL database, secured through
  [`prompt-sdk`](https://github.com/prompt-edu/prompt-sdk).

There are two ways to build one:

| Model                        | Lives in                | Good for                                        |
| ---------------------------- | ----------------------- | ----------------------------------------------- |
| **In-repo phase**            | this monorepo           | phases maintained by the PROMPT team            |
| **External phase**           | its own repository      | institution- or course-specific custom phases   |

Both start from the same reference implementation:
[prompt-edu/prompt-example-phase](https://github.com/prompt-edu/prompt-example-phase), a minimal
working phase whose `client/` and `server/` each have a README explaining the file layout. External
phases follow the model of [prompt-intro-course](https://github.com/prompt-edu/prompt-intro-course)
and [prompt-github-challenge](https://github.com/prompt-edu/prompt-github-challenge).

Whichever model you pick, the phase must be **registered in core** (section 2). That part is
identical for both.

---

## 1. Scaffold the phase

### In-repo: use the generator

```bash
make new-phase NAME=feedback CLIENT_PORT=3014 SERVER_PORT=8093
```

The script (`scripts/new-course-phase.sh`) clones prompt-example-phase, renames every identifier
with the repository's own `init.sh`, converts it to the workspace layout of this monorepo, and
applies all registration steps below except the phase type (section 2.4) and the deployment
workflows (section 3). It prints a checklist of what remains manual. Pick ports that are free; the
script aborts if they collide with an existing phase.

The script needs network access to GitHub. To scaffold from another source, for example a local
clone with unpublished changes, set `PHASE_TEMPLATE_SOURCE` (a git URL or path) and
`PHASE_TEMPLATE_REF` (a branch or tag, default `main`). The workspace-side files it writes (client
`package.json`, `Dockerfile`, `biome.json`, the core loaders, and the docker-compose blocks) are
templates in `scripts/new-course-phase/`.

<details>
<summary>Manual checklist (what the generator automates)</summary>

**Client** (`clients/<name>_component`, name must end in `_component`; `core/src/declaration.d.ts`
types remotes by that suffix):

1. Clone prompt-example-phase, run `./init.sh <name> <client_port> <server_port>` in it, and copy
   its `client/` to `clients/<name>_component`.
2. Delete the standalone parts: `shared/` (the monorepo has `clients/shared/`), `yarn.lock`,
   `nginx.conf`, and `.yarn/`. Point the `shared` imports in `rspack.config.mjs` and
   `src/bootstrap.tsx` one level up, and set `rootDir` in `tsconfig.json` to `..`.
3. Replace `package.json`, `Dockerfile`, `.dockerignore`, and `biome.json` with the workspace
   versions: copy them from an existing component such as `clients/presentation_component` and
   rename. `biome.json` extends the root (`"extends": "//"`).
4. Register the workspace in `clients/package.json` (`workspaces.packages`),
   `clients/lerna.json` (`packages`), and the `COPY <workspace>/package.json` list in
   `clients/Dockerfile`, then `cd clients && yarn install`.

**Server** (`servers/<name>`):

1. Copy the renamed `server/` to `servers/<name>` and delete its `Dockerfile` (the monorepo
   builds every service from `servers/Dockerfile`).
2. `go.mod`: module path `github.com/prompt-edu/prompt/servers/<name>`; update all imports.
3. Keep the `config/` and `copy/` packages. They implement the SDK phase config/copy contracts
   that course templating and deep copy rely on.
4. Schema in `db/migration/`, queries in `db/query/`, then `sqlc generate` (see the
   [server guide](./guide/server.md)).

**Infrastructure:**

1. `docker-compose.yml`: add `server-<name>`, `client-<name>-component`, and `db-<name>-server`
   services, kebab-case like the existing ones (fill in the templates in
   `scripts/new-course-phase/`; separate database per service, own `postgres_<name>_data`
   volume). `DB_HOST_<NAME>_SERVER` in `.env.template` must match the db service name, and
   `DB_PORT_<NAME>_SERVER` there is `5432` (the port inside the Docker network).
2. `.env.template` + `.env.dev.template`: add the `DB_*_<NAME>_*` entries and `<NAME>_HOST`.
3. `Makefile`: add `server-<name>`, `test-<name>`, `sqlc-<name>` targets and wire them into the
   aggregate `servers`, `test`, `sqlc`, `lint-servers`, and `DB_SERVICES` lists.
4. CI: add the client to the `quality-clients.yml` matrix and the server to the
   `quality-servers.yml` and `test-servers.yml` matrices.

</details>

### External: own repository

1. Open [prompt-example-phase](https://github.com/prompt-edu/prompt-example-phase) and click
   **Use this template**.
2. Clone your new repository and run `./init.sh <name> <client_port> <server_port>`. It renames
   every placeholder, sets the Go module path from your `origin` remote, and checks the build.
3. Follow its README to run it against a local PROMPT stack.

The repository comes with `client/` + `server/` + `docker-compose.yml` (dev and production with
traefik) + CI that builds and pushes both images. Your client is built and served by your own
deployment; your server runs against its own database. You then need one small PR to this repo
for core registration, exactly the steps in section 2, with your deployed URLs instead of
localhost ones.

The client shares React, React Router, React Query, recharts, and the two `@tumaet/prompt-*`
packages with core as Module Federation singletons. Keep their versions in your
`client/package.json` the same as in `clients/package.json` of the PROMPT version you deploy
against.

**Styling.** In-repo phases get their utilities from core, which scans every component directory
(see the styling section of `docs/contributor/guide/client.md`). Core cannot scan a repository it
does not have, so an external phase has two constraints. Utilities it uses that no in-repo client
uses will not be in core's stylesheet, so stay close to the shared UI library and to the utilities
core already emits. And it must not inject a Tailwind build of its own into the host: that
stylesheet lands in `<head>` after core's and overrides core's utilities at equal specificity,
which is issue #2086.

---

## 2. Register the phase in core

These steps make core aware of your phase. All paths are under `clients/core`.

### 2.1 Module Federation remote

Add an entry to `clients/core/remotes.config.mjs`:

```js
{
  name: 'feedback_component',
  phaseTypeName: 'feedback_component', // your course phase type name, see 2.3 and 2.4
  devPort: 3014,
  prodPath: '/feedback',
},
```

`clients/core/rspack.config.mjs` turns every entry into a federation remote, with a `Date.now()`
suffix that busts the remote-entry cache, and the admin System Status page probes each entry.
`clients/core/remotes.config.test.ts` fails when an entry's `phaseTypeName` has no matching key in
`PhaseRouterMapping.tsx`.

For external phases the production URL typically comes from an env var exposed via
`core/public/env.template.js` (see 2.5).

### 2.2 External routes and sidebar loaders

Add `<Name>Routes.tsx` (`src/managementConsole/PhaseMapping/ExternalRoutes/`) and
`<Name>Sidebar.tsx` (`.../ExternalSidebars/`). Copy an existing pair such as
`PresentationRoutes.tsx` / `PresentationSidebar.tsx`, or fill in `Routes.tsx.tmpl` /
`Sidebar.tsx.tmpl` from `scripts/new-course-phase/`, and update the import string to your remote.
The import must stay a **static string literal**: the bundler needs it for code splitting, and the
copy-per-phase pattern keeps loading/caching behavior correct.

### 2.3 Phase mappings

Map your **course phase type name** (as stored in the DB, see 2.4) in three files under
`src/managementConsole/PhaseMapping/`:

- `PhaseRouterMapping.tsx` → your `<Name>Routes`
- `PhaseSidebarMapping.tsx` → your `<Name>Sidebar`
- `PhaseStudentDetailMapping.tsx` → your student-detail provider (or `Fallback`)

The management console serves every phase through the generic
`/management/course/:courseId/:phaseId/*` route in `src/App.tsx`, so no route of your own is
needed.

### 2.4 Course phase type

Core seeds phase types at startup in `servers/core/coursePhaseType/initializeTypes.go`. Add an
`init<Name>()` following the existing pattern: name (must match the mapping keys from 2.3),
`BaseUrl` (`{CORE_HOST}/<name>/api` in production, `devServiceURL(<port>, "/<name>/api")` in
dev, which shifts the port to the stack a worktree runs in), description, and the phase's required inputs / provided outputs. For a purely private deployment
you can instead insert the row into `course_phase_type` manually, but the initializer is the
supported path.

### 2.5 Runtime environment variable (production)

Add `<NAME>_HOST` to `clients/core/public/env.template.js` (substituted by `envsubst` in the
core client's entrypoint) and to `env.js` for local dev. Note: the typed `EnvType` lives in
`@tumaet/prompt-shared-state`; until your var is added there, read it from `window.env` (see
`client/src/example_component/network/exampleServerConfig.ts` in prompt-example-phase for the
pattern).

---

## 3. Deployment (in-repo phases)

1. `docker-compose.prod.yml`: add the client service with traefik labels (copy an existing block
   such as `client-presentation-component`: router rule `PathPrefix(/<name>)`, strip-prefix and
   compress middlewares) and the server + db services.
2. `build-and-push-clients.yml`: add a build job and an `<name>_image_tag` output. Pass
   `MF_BUILD_VERSION` in its `build-args` like the other remote jobs, so the System Status page
   shows the client's version.
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
  phase of your new type to a course. Core must lazy-load your remote, and the sidebar entry
  must appear with the configured permissions.
