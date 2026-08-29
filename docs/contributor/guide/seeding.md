---
sidebar_position: 3
---

# Database Seeding

A fresh `make db` gives you eight **empty** databases. `make seed` fills them with
one realistic course so you can click through PROMPT without creating anything by
hand.

```bash
make db          # every service database + Keycloak
make servers     # the servers own the schemas: their startup migrations create them
make seed        # load the demo course into all eight databases
```

`make seed` is re-runnable. It is **authoritative** for the rows it owns: it deletes
them and inserts them again, so a second run reconciles whatever you changed rather
than failing on a duplicate key. Everything inside the seeded courses is therefore
disposable - do not keep work there.

## What you get

**`iPraktikumDemo`** (semester `ios2526`), a course with data in every phase:

| Phase | In the graph | Seeded with |
| --- | --- | --- |
| Application | yes (initial) | text, multi-select and file-upload questions, answers and scores for 8 applicants |
| Interview | yes | two questions, four slots, six assignments, six scored reviews |
| Matching | yes | the pass/fail funnel |
| Team Allocation | yes | 3 skills, 2 teams, survey responses, a published allocation |
| Assessment | yes | a rubric with 4 competencies, 6 graded participants, a self-evaluation round |
| Presentation | yes | a team presentation schedule and feedback categories |
| Certificate | yes | a released Typst template |
| Self Team Allocation | no | 2 student-formed teams |
| Example | no | the module's single row |

The last two sit **off the graph**. `course_phase_graph` is UNIQUE on both of its
endpoint columns, so a course's phases form a strict chain that cannot branch - and
Team Allocation and Self Team Allocation are alternatives no real course runs
together. They still carry full data; they are just not part of the sequential
student flow.

The seed also contains the fixture courses the Playwright suite asserts against
(`iPraktikum`, `iPraktikum-Test`, `TestCourse`, `iPraktikumFull`). Leave those alone
unless you are changing an e2e test.

## Logging in

The Keycloak users are seeded by `keycloakConfig.json` (username == password):

| User | Sees the demo course as |
| --- | --- |
| `admin` | PromptAdmin (sees every course) |
| `lecturer`, `course-lecturer` | Course Lecturer |
| `course-editor` | Course Editor |
| `student` (Stan), `student2` (Selma) | enrolled students |

Staff access comes from the `ios2526-iPraktikumDemo-Lecturer` / `-Editor` client
roles. Student access does **not**: it is derived in the database by matching the
token's `university_login` and `matriculation_number` against the `student` row.

:::warning Existing Keycloak installs
Keycloak imports the realm only into an empty database. If your Keycloak already
exists, the demo course roles will be missing until you recreate it:

```bash
make db-down
rm -rf keycloak_postgres_data
make db
```
:::

## How it works

`scripts/seed.sh` applies one data-only SQL file per service from `seed/`.

Data only, not a dump: the servers own the schema. Core shells out to `migrate` at
startup and the phase services call `sdkUtils.RunMigrations`, so the seed must run
**after** them. The script waits per database until `schema_migrations` reaches the
highest numbered migration in that service's `db/migration` directory and is not
dirty. There is no version file to keep in sync - the expected version is read from
the source tree at run time.

Core gets one extra gate: the seed resolves course phase types **by name**, because
`servers/core/coursePhaseType/initializeTypes.go` creates them with random UUIDs on
every fresh database. The script therefore also waits for all eight type names the
seed uses. Two consequences worth knowing:

- Never insert a `course_phase_type` row yourself. Core skips a type's provided and
  required DTO descriptors whenever the type row already exists, so pinning the type
  would silently cost you the inter-phase data dependency metadata. The one exception
  is `example_component`: no service creates it, so `seed/core.sql` inserts it. That
  is an ownership gap rather than a pattern to copy - its `base_url` is seed-owned
  configuration nothing reconciles, so it points at core's gateway path and does not
  reach a host-run example server on 8086.
- `base_url` comes out right per environment (`http://localhost:8087/...` locally,
  `{CORE_HOST}/...` in Docker) instead of being frozen into the seed.

Each file runs in a single transaction with `ON_ERROR_STOP`, so a mistake rolls back
instead of half-seeding. The readiness checks all run before the first delete, but
there is no transaction *across* databases: if the run fails midway, some databases
hold the new seed and others the previous one, and the next run repairs it.

`make seed` runs the script inside a `postgres:15.19-alpine` container on the compose
network, so you need no local `psql` and local development takes the same code path
as CI. The script itself is plain POSIX `sh` and runs on the host too if you have
`psql`:

```bash
./scripts/seed.sh                       # every service
./scripts/seed.sh core assessment       # just these
SEED_CORE_PORT=55432 ./scripts/seed.sh core
```

`core` is applied first whatever order you list the targets in, because every other
file references ids only `seed/core.sql` creates. Seeding a phase database on its own
therefore aborts unless core already holds the demo course - otherwise the phase rows
would point at course phases that do not exist, and no foreign key would say so.

Every connection setting is per service: `SEED_<SERVICE>_HOST`, `_PORT`, `_NAME`,
`_USER` and `_PASSWORD`. Each falls back to `SEED_DB_NAME`, `SEED_DB_USER` and
`SEED_DB_PASSWORD` (localhost and the default port for host and port), so the split
only matters once a database uses credentials of its own - which is how
`docker-compose.yml` passes each service its `DB_<SERVICE>_*` values.

## Changing the seed

Edit the `INSERT` blocks. There is nothing to regenerate: a migration only forces a
seed change when it renames or drops a column the seed writes, and then `psql` fails
loudly on the next run.

Every service has its own database, and the ids linking them (`course_phase_id`,
`course_participation_id`, `student_id`) are bare UUID values that no foreign key
enforces. A typo there shows up as an empty screen, not an error, so:

```bash
make seed-check
```

verifies that every core-owned id the phase seeds reference is one `seed/core.sql`
actually creates. `seed/README.md` lists the pinned id ranges.

If you add a phase to the demo course, remember the graph is a chain: append to the
tail or leave the phase off the graph.

## The e2e suite

The Playwright stack loads the same files. `docker-compose.e2e.yml` runs a one-shot
`seed` service after the servers start, and `e2e-runner` waits for it to finish. See
[`e2e/README.md`](https://github.com/prompt-edu/prompt/blob/main/e2e/README.md) for
the fixture data the specs rely on.
