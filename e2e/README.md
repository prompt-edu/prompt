# PROMPT 2.0 End-to-End Tests

Black-box e2e tests that boot the **core server + core client + Keycloak +
Postgres + SeaweedFS** in Docker — plus the **self team allocation**,
**assessment**, and **interview phase modules** (Go service, own Postgres,
Module Federation remote each) — and drive them like a real user, with
[Playwright](https://playwright.dev). They catch full-stack regressions (auth
flow, routing, API contract, data rendering, remote loading) that the Go unit
tests can't.

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
 ├── client-core   built from ../clients/core (nginx, e2e conf with the
 │                 phase-module proxy locations — nginx/client-core.conf)
 ├── self team allocation phase module:
 │    ├── db-self-team-allocation      own ephemeral Postgres (empty; the
 │    │                                server migrates it on startup)
 │    ├── server-self-team-allocation  built from ../servers/self_team_allocation
 │    └── client-self-team-allocation  the Module Federation remote (nginx)
 ├── assessment phase module:
 │    ├── db-assessment      own ephemeral Postgres (empty; the server's
 │    │                      migrations also create the default schemas)
 │    ├── server-assessment  built from ../servers/assessment
 │    └── client-assessment  the Module Federation remote (nginx)
 ├── interview phase module:
 │    ├── db-interview      own ephemeral Postgres (empty; the server runs its
 │    │                     own migrations on startup)
 │    ├── server-interview  built from ../servers/interview
 │    └── client-interview  the Module Federation remote (nginx)
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

## Phase modules in the e2e stack

The self team allocation module is the blueprint for adding a course-phase
module (Go service + Module Federation remote) to the stack; the assessment
(see `tests/assessment/`) and interview (see `tests/interview/`) modules are
further implementations of it. To add another module, copy each of these steps:

**1. Compose services** (`docker-compose.e2e.yml`): a `db-<module>` Postgres
(ephemeral, `pg_isready` healthcheck), a `server-<module>` (build
`./servers/<module>`, `dockerfile: ../Dockerfile`, depends on its db +
keycloak), and a `client-<module>` (build `./clients/<module>_component` with
the shared `*client-build-args` anchors, wget healthcheck). Phase servers read
the generic `DB_USER`/`DB_PASSWORD`/`DB_NAME` plus service-specific
`DB_HOST_<MODULE>`/`DB_PORT_<MODULE>` env vars — map them explicitly from new
`DB_<MODULE>_*` entries in `.env.e2e`. No host ports; everything is reached
through the client-core proxy. Add both services to `client-core`'s and
`e2e-runner`'s `depends_on`.

**2. Routing** (`nginx/client-core.conf`): the core client is a **production**
build, so remotes are baked as relative URLs (`/<module>/remoteEntry.js`) and
the phase API is called on the browser origin. In prod Traefik routes these
prefixes; here client-core's nginx does, with the exact same semantics:

```nginx
location /<module>/api/ { proxy_pass http://server-<module>:8080; }  # prefix KEPT
location /<module>/    { proxy_pass http://client-<module>/; }       # prefix STRIPPED
```

Set the module's host var in `.env.e2e` to **empty** (`<MODULE>_HOST=`) so the
component's axios baseURL falls back to the browser origin (same-origin, no
CORS).

**3. Core seed** (`seed/e2e_seed.sql`): pre-insert the module's
`course_phase_type` row with a **fixed UUID** — core's startup init matches by
name and would otherwise create it with a random UUID, breaking seeded FK
references. Also mirror the type's provided/required DTO rows from
`servers/core/coursePhaseType/initializeTypes.go` (init skips them when the
type exists). Then add a `course_phase` on the seeded course and a
`course_phase_graph` edge. Careful: the graph has UNIQUE constraints on both
`from_` and `to_course_phase_id`, so phases form a **chain** — append your
phase to the tail (currently the self team allocation phase), don't branch
from Application. Add `course_phase_participation` rows for `student`/`student2`
(the student UI 404s into UnauthorizedPage without an own participation;
backend auth checks course-level enrollment via core's `is_student`).

**4. Phase DB seed:** usually **none** — the phase server migrates its empty
DB on startup and tests create their own data. If a module does need seeded
phase data, mount a dump into the db's `/docker-entrypoint-initdb.d/` with
`schema_migrations` pinned to the module's latest migration (same pattern as
the core seed below).

**5. Readiness** (`src/global-setup.ts`): phase server images are distroless
(no healthcheck), so poll `${BASE_URL}/<module>/api/info` with
`waitForServiceInfo(...)` — it asserts the `/info` JSON (`serviceName`,
`healthy`), which both gates on the server and proves the proxy wiring (a bare
status poll would accept the SPA fallback's 200).

**6. Smoke test** (`tests/<module>/mf-smoke.spec.ts`): open the seeded phase at
`/management/course/<courseId>/<phaseId>` and assert a heading rendered by the
**remote** — if `remoteEntry.js` fails to load, the shell renders a
LoadingError instead, so this one assertion covers the whole Module Federation
path. Journeys and API auth checks build on top (see
`tests/self-team-allocation/`, `tests/api/self-team-allocation.api.spec.ts`).

## Authentication

The realm (`keycloak/realm.json`, derived from the repo's `keycloakConfig.json`
with redirect URIs / web origins repointed at the e2e service names) seeds users
whose username == password:

| role            | username / password | permission              |
| --------------- | ------------------- | ----------------------- |
| `admin`         | `admin`             | `PROMPT_Admin`          |
| `lecturer`      | `lecturer`          | `PROMPT_Lecturer`       |
| `course-lecturer` | `course-lecturer` | `PROMPT_Course_Lecturer` + `ios2425-iPraktikum-Lecturer` |
| `course-editor` | `course-editor`     | `PROMPT_Course_Editor`  |
| `student`       | `student`           | `PROMPT_Student` + `ios2425-iPraktikum-Student` |
| `student2`      | `student2`          | `PROMPT_Student` + `ios2425-iPraktikum-Student` |

The `ios2425-iPraktikum-*` roles are **course-scoped** roles
(`<semesterTag>-<courseName>-<Role>`) for the seeded iPraktikum course — PROMPT
authorizes course/phase access against these, not the global `PROMPT_*` roles.
`student`/`student2` also exist as `student` table rows in the DB seed (matched
by the `matriculation_number`/`university_login` token claims) with
`course_participation` rows, so they are real course members end to end.
`student2` exists for two-user journeys (e.g. one student creates a team, the
other joins it).

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

It contains three read-only courses (`iPraktikum`, `iPraktikum-Test`,
`TestCourse`) plus **`iPraktikumFull`**, a course with a full linear phase graph
(Application → Interview → Matching → Team Allocation → Assessment), ~7 course
participations, and a funnel of phase participations. The seeded `student` user
maps to a DB `student` row (matriculation `00000005` / login `no42tum`) that
participates in every `iPraktikumFull` phase; `student2` (Selma) is also
enrolled for the assessment visibility checks. That is how a student gets
course access (roles are DB-derived from `matriculation_number` +
`university_login`, not from Keycloak). The `lecturer` and `course-lecturer`
users hold the `ios2425-iPraktikumFull-Lecturer` role and `course-editor` holds
`ios2425-iPraktikumFull-Editor` (added to `keycloak/realm.json`). See
`FULL_COURSE_PHASES`, `FULL_COURSE_STUDENT`, `FULL_COURSE_STUDENT2`, and
`FULL_COURSE_ROLES` in `src/data/constants.ts`.

Spec files run in **parallel workers** locally, so every spec file that mutates
phase state owns its own seeded phase. The assessment specs follow this rule
(release state and schema locking never cross files): the graph-tail Assessment
phase hosts the lecturer journey + smoke + API checks, and two **standalone**
Assessment phases (no graph edges — they route by URL but are filtered from the
course sidebar) host the visibility and self-evaluation journeys. A
participant-less Assessment phase on `TestCourse` is the negative-auth fixture.
See `ASSESSMENT_FIXTURE_PHASES` and `ASSESSMENT_FOREIGN_PHASE_ID` in
`src/data/constants.ts`. For the same reason the self team allocation
lecturer-overview spec owns a standalone phase
(`SELF_TEAM_ALLOCATION_OVERVIEW_PHASE_ID`) — the team it forms would otherwise
block team creation in the student journey. The assessment server needs **no
phase-DB seed**: its migrations create the default template schemas, and the
first `GET /config` on a phase binds it to them. Peer/tutor evaluation journeys
are not covered — they need team data from a team-allocation resolution the
e2e seed does not wire.

The `iPraktikumFull` Application phase is open (start 2020, end 2099, external
students allowed) and carries one required text question (`Motivation`) so the
application journey exercises the configurable form. `TestCourse` has a
**closed** Application phase (`CLOSED_APPLICATION_PHASE_ID`, deadline in 2020)
as the negative fixture for the public apply endpoints.

> The course name (`iPraktikumFull`) and `semester_tag` (`ios2425`) are
> **hyphen-free on purpose**: the course-list query parses course roles with
> `split_part(role, '-', N)`, so a hyphen in either would break course
> visibility for non-admins.

> The `Interview` / `Matching` / `Team Allocation` phase types are seeded by
> their canonical names (matching
> `servers/core/coursePhaseType/initializeTypes.go`), so core's startup
> initializer skips re-creating them. As a result they carry **no provided/
> required DTO metadata** — fine for phase-graph, participant-list, and
> role-access tests, but the inter-phase data-dependency graph is not exercised.
> (`Assessment` and `Self Team Allocation` DO mirror their DTO rows, per step 3
> of the module blueprint.) The `Interview` remote **is** now built into the e2e
> client and exercised through its own UI (see `tests/interview/`); the
> `Matching` / `Team Allocation` / … remotes are not, so tests for those should
> target core-level views (course config, phase graph, participant lists,
> role-based access), not the phase remotes' own UIs.

> Note: the repo's `servers/core/database_dumps/full_db.sql` is **not** usable
> as an e2e seed — it's a hand-maintained Go-test fixture whose schema is
> internally inconsistent (some tables migrated, others not; `schema_migrations`
> stuck at 9), so `migrate up` cannot run against it.

**Data-only changes** (adding courses, phases, participations, students, roles
without a schema migration) are made by editing the `INSERT` blocks in
`seed/e2e_seed.sql` directly and keeping `schema_migrations = 24`. Full
regeneration is only needed when a core migration changes the schema.

To regenerate after a **schema** change, apply the migrations to a throwaway
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
change in `keycloakConfig.json`, regenerate it and re-apply those URL edits —
plus the e2e-only additions: the course-scoped `ios2425-iPraktikum-Student` /
`ios2425-iPraktikum-Lecturer` client roles (assigned to `student`, `student2`,
and `course-lecturer`) and the `student2` user.

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
nginx/client-core.conf client-core nginx with the phase-module proxy locations
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
