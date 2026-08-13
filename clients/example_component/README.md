# Example Component

This is the reference micro-frontend for a PROMPT course phase. It is intentionally minimal:
copy it, rename it, and you have a working phase client that plugs into the core management
console. Its backend counterpart is [`servers/example_server`](../../servers/example_server).

> **Fastest start:** run `make new-phase NAME=<name> CLIENT_PORT=<port> SERVER_PORT=<port>` from
> the repository root — it copies this component and the example server and performs every
> registration step listed below automatically.

## How a phase client plugs into PROMPT

PROMPT's core client (`clients/core`) is a Webpack Module Federation host. Each course phase is a
**remote** that core loads at runtime based on the course configuration. A phase remote must expose
exactly three modules (see `rspack.config.mjs` → `exposes`):

| Exposed module | File           | Consumed by core as                               |
| -------------- | -------------- | ------------------------------------------------- |
| `./routes`     | `routes/`      | React Router routes rendered inside the console   |
| `./sidebar`    | `sidebar/`     | Sidebar menu entry (title, icon, subitems)        |
| `./provide`    | `src/provide/` | Optional extension points (e.g. `StudentDetail`)  |

Routes and sidebar items carry `requiredPermissions` (roles from `@tumaet/prompt-shared-state`);
core filters them per user.

## File map

```text
routes/index.tsx        Routes the phase contributes (Overview, Participants, Settings)
sidebar/index.tsx       Sidebar entry: title, icon, subitems with permissions
src/
  provide/              Extension points; StudentDetail renders inside core's student view
  OverviewPage.tsx      Landing page of the phase
  example_component/
    network/            Axios instance for the phase server + React Query queries
    pages/              ParticipantsPage (shared participations table), SettingsPage
  bootstrap.tsx         Standalone entry (running the component on its own, rare)
  index.js              Async boundary required by Module Federation
rspack.config.mjs       COMPONENT_NAME + COMPONENT_DEV_PORT constants, MF exposes/shared
Dockerfile              Production build served by nginx (used by docker-compose)
```

## What the code demonstrates

- **Talking to your phase server** — `src/example_component/network/exampleServerConfig.ts`
  creates an Axios instance with the JWT interceptor; `queries/getExampleInfo.ts` shows the
  React Query pattern against `/example-service/api/course_phase/:coursePhaseID/...`.
- **Reusing shared UI** — pages use `@tumaet/prompt-ui-components`
  (`ManagementPageHeader`, `ErrorPage`, `LoadingPage`, Cards) and the shared
  `CoursePhaseParticipationsTable` instead of custom widgets.
- **Role-gated navigation** — `requiredPermissions` on routes and sidebar subitems.

## Run it locally

```bash
make db          # Postgres + Keycloak
make server      # core server
make clients     # all micro-frontends, this one on port 3001
# or standalone: cd clients/example_component && yarn dev
```

Open <http://localhost:3000> and add a phase of type `example_component` to a course.

## Creating your own phase from this component

Prefer `make new-phase` (see above). To do it manually:

1. `cp -R clients/example_component clients/<name>_component` and rename the inner
   `src/example_component` directory.
2. Update `COMPONENT_NAME` and `COMPONENT_DEV_PORT` in `rspack.config.mjs`, and `"name"` in
   `package.json`.
3. Register the workspace in `clients/package.json` and `clients/lerna.json`.
4. Register the remote + phase mappings in core and the services in docker-compose — the full
   checklist lives in `docs/contributor/new_course_phase.md`.

Keep this component generic: it is the source for new phases and is planned to move into a
standalone GitHub template repository (see `template-repository/` at the repo root).
