# PROMPT 2.0 End-to-End Tests

Black-box e2e tests that boot the **core server + core client + Keycloak +
Postgres + SeaweedFS** in Docker and drive them like a real user, with
[Playwright](https://playwright.dev). They catch full-stack regressions (auth
flow, routing, API contract, data rendering) that the Go unit tests can't.

The suite covers two layers with one framework:

- **Browser** (`tests/**/*.spec.ts`) - real Chromium against the core client.
- **API** (`tests/api/*.api.spec.ts`) - direct HTTP calls to the core server.

---

## Running it

Two ways to run, both driven by `make` from the **repo root**.

### 1. Full run in Docker (canonical — same locally and in CI)

Everything (app stack + the Playwright runner) runs in containers. This is what
CI runs and what you should trust for a green/red verdict. Only Docker required.

```bash
make test-e2e        # build the stack, run all specs, exit non-zero on failure
make test-e2e-down   # stop the stack and remove volumes (also done automatically)
```

The HTML report lands in `e2e/playwright-report/` (open `index.html`, or
`npx playwright show-report e2e/playwright-report`). On CI it's uploaded as the
`playwright-report` artifact.

### 2. Interactive mode (watch / debug)

The **Playwright UI** runs inside the stack and is served to your browser. You
get the full time-travel experience - per-step DOM snapshots, network, console,
watch mode (edits to `tests/` and `src/` re-run live) - with the same in-network
config as the canonical run, so auth works identically. Only Docker required.

```bash
make test-e2e-ui     # builds + boots the stack, then serves the Playwright UI
```

Then open **http://127.0.0.1:8123** in your browser (use `127.0.0.1`, not
`localhost` - some Docker backends don't publish the port over IPv6). Press the
▶ buttons to run specs, click any step to time-travel, and edit specs in your
editor to re-run on save. Stop with `Ctrl-C` (the stack is torn down
automatically).

> Why in Docker and not on the host? The app, server, and Keycloak address each
> other by Docker service name, and the token issuer is tied to those names.
> Running the test browser inside the network keeps the real login flow working
> without host-networking/DNS workarounds.

Prefer a CLI subset or a one-off headless run? Use the canonical runner with
extra args, e.g.:

```bash
# run a single spec file, headless, in the containerized runner
docker compose -f docker-compose.e2e.yml --env-file e2e/.env.e2e run --rm \
  e2e-runner npx playwright test tests/courses
docker compose -f docker-compose.e2e.yml --env-file e2e/.env.e2e down -v
```

---

## How it fits together

```
docker-compose.e2e.yml
 ├── db            ephemeral Postgres, seeded from seed/e2e_seed.sql on init
 ├── keycloak(+db) realm imported from keycloak/realm.json (seeded users)
 ├── seaweedfs-*   S3 storage the core server depends on
 ├── server-core   built from ../servers/core (runs migrate up on startup)
 ├── client-core   built from ../clients/core (nginx)
 └── e2e-runner    Playwright container; waits for health, runs this suite
```

Everything is ephemeral (no host DB volumes), so each run starts from a clean,
deterministically seeded database. The databases and SeaweedFS stay internal to
the compose network; only client / API / Keycloak ports are published.

**Networking:** in the containerized run, the runner's browser, the app's
runtime config, the core server, and Keycloak all address each other by **Docker
service name** (`client-core`, `server-core`, `keycloak`). This matters because
Chromium hard-codes `localhost` → loopback and ignores `/etc/hosts`, so a
`localhost`+host-gateway remap does **not** work for an in-container browser. With
service names, the realm's redirect URIs and the token issuer all line up.
Interactive mode (`make test-e2e-ui`) keeps this exact networking - the test
browser still runs inside the network; only the Playwright **UI** is served out
to your browser - so auth behaves identically to the canonical run.

## Authentication

The realm (`keycloak/realm.json`, derived from the repo's `keycloakConfig.json`
with redirect URIs / web origins repointed at the e2e service names) seeds users
whose username == password:

| role            | username / password | permission              |
| --------------- | ------------------- | ----------------------- |
| `admin`         | `admin`             | `PROMPT_Admin`          |
| `lecturer`      | `lecturer`          | `PROMPT_Lecturer`       |
| `course-lecturer` | `course-lecturer` | `PROMPT_Course_Lecturer`|
| `course-editor` | `course-editor`     | `PROMPT_Course_Editor`  |
| `student`       | `student`           | `PROMPT_Student`        |

`src/global-setup.ts` logs in each role **once** through the real Keycloak form
and saves the session to `.auth/<role>.json`. Tests reuse it, so the login flow
is exercised once and everything after is fast. A 24h Keycloak SSO cookie keeps
sessions valid even after the 5-min access token expires.

## Adding a browser test

1. Create `tests/<area>/<name>.spec.ts`.
2. Import the auth-aware test and pick a role:

   ```ts
   import { test, expect } from '../../src/fixtures/auth'
   import { CoursesPage } from '../../src/pages/CoursesPage'

   test.use({ role: 'admin' }) // starts already logged in as admin

   test('admin sees courses', async ({ page }) => {
     const courses = new CoursesPage(page)
     await courses.goto()
     await courses.expectLoaded()
   })
   ```

3. Put selectors in a **Page Object** under `src/pages/` (don't scatter raw
   locators across specs). Prefer role/label/text selectors:
   `getByRole('heading', { name: 'Courses' })`.
4. When no stable selector exists, add a `data-testid` to the core client
   component (`clients/core/src/...`) and select it with `getByTestId(...)`.
   The client has few testids today - adding them as needed is expected.

## Adding an API test

1. Create `tests/api/<name>.api.spec.ts` (the `.api.spec.ts` suffix routes it to
   the `api` Playwright project - no browser).
2. Use the `apiAs(role)` fixture, which mints a real bearer token via Keycloak's
   password grant and returns an `APIRequestContext` pointed at the core API:

   ```ts
   import { test, expect } from '../../src/fixtures/api'

   test('GET /api/courses', async ({ apiAs }) => {
     const api = await apiAs('lecturer')
     const res = await api.get('/api/courses/')
     expect(res.ok()).toBeTruthy()
   })
   ```

## Test data

Assertions reference seeded rows via `src/data/constants.ts` (course/student IDs
and names). The seed lives in `seed/e2e_seed.sql`.

The seed is a **consistent dump of the current (v24) schema** with a small
deterministic data set. It records `schema_migrations = 24`, so the core
server's startup `migrate up` is a clean no-op and the data loads as-is.

> Note: the repo's `servers/core/database_dumps/full_db.sql` is **not** usable
> as an e2e seed — it's a hand-maintained Go-test fixture whose schema is
> internally inconsistent (some tables migrated, others not; `schema_migrations`
> stuck at 9), so `migrate up` cannot run against it.

To regenerate after a schema change, apply the migrations to a throwaway
Postgres, insert the data, and dump it:

```bash
docker run -d --name seedgen -e POSTGRES_USER=prompt-postgres \
  -e POSTGRES_PASSWORD=prompt-postgres -e POSTGRES_DB=prompt postgres:15.18-alpine
# wait for readiness, then:
for f in servers/core/db/migration/*.up.sql; do
  docker exec -i seedgen psql -v ON_ERROR_STOP=1 -U prompt-postgres -d prompt < "$f"
done
docker exec -i seedgen psql -U prompt-postgres -d prompt <<'SQL'
CREATE TABLE schema_migrations (version bigint PRIMARY KEY, dirty boolean NOT NULL);
INSERT INTO schema_migrations VALUES (24, false);
-- ...your INSERTs for courses / students / course_phase_type...
SQL
docker exec seedgen pg_dump --no-owner --no-privileges --inserts \
  -U prompt-postgres -d prompt > e2e/seed/e2e_seed.sql
docker rm -f seedgen
# bump `24` to the latest migration number, and update src/data/constants.ts
```

### Keeping the e2e realm in sync

`keycloak/realm.json` is a copy of the repo's `keycloakConfig.json` with the
`prompt-client` redirect URIs / web origins (and `prompt-server` redirect URIs)
repointed at the e2e ports (`4000` / `18090`). If users, roles, or clients
change in `keycloakConfig.json`, regenerate it and re-apply those URL edits.

## Conventions

- One user journey per spec; keep specs independent (no ordering dependence).
- Any data a test creates, it should clean up.
- Keep selectors resilient: role/label/testid over CSS/text-with-styling.
- Don't assert on incidental copy that changes often.

## Debugging

- **Report:** `npx playwright show-report e2e/playwright-report` (or download the
  `playwright-report` artifact from the CI run).
- **Traces/videos:** captured on failure under `test-results/` (`trace.zip`,
  `video.webm`). Open a trace with `npx playwright show-trace <path>`.
- **Interactive:** `make test-e2e-ui`, then open http://127.0.0.1:8123.
- **Stack logs:** `docker compose -f docker-compose.e2e.yml logs server-core`
  (or `client-core`, `keycloak`).

## Files

At the repo root:

```
docker-compose.e2e.yml   the e2e stack (canonical, service-name networking)
```

In `e2e/`:

```
playwright.config.ts   projects (api, chromium), globalSetup, reporters
Dockerfile             Playwright runner image (tag must match @playwright/test)
.env.e2e               compose env (test-only credentials)
keycloak/realm.json    dedicated e2e Keycloak realm (seeded users + roles)
seed/e2e_seed.sql      seeded core DB (consistent v24 schema + data)
src/
  env.ts               URLs / endpoints
  data/                roles + seeded-data constants
  fixtures/auth.ts     role option → storageState injection (browser)
  fixtures/api.ts      apiAs(role) → authed APIRequestContext (API)
  pages/               Page Objects
  global-setup.ts      logs in each role, writes .auth/<role>.json
tests/                 specs
```
