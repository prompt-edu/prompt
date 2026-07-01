---
sidebar_position: 4
---

# API Stress & Fuzz Testing

The `api-stress/` suite drives **every** PROMPT API endpoint to find what is **slow,
degrades under load, errors, or has an access-control gap**, and actively tries to
break the servers via the API only. It produces a prioritized, bug-fixable report.
It is designed to be **structured, repeatable, and readable by both humans and AI
agents**.

> Two consecutive runs produce byte-identical findings - the suite is deterministic.

## When to use it

- Before a release, to catch regressions in latency, error handling, or authorization.
- When changing the auth/SDK middleware, course/phase model, or any handler that
  touches the database (the suite is good at finding unhandled `pgx.ErrNoRows`).
- To get a per-endpoint latency baseline and a list of endpoints that 5xx.

## Architecture

The suite is a thin pipeline around a single source of truth (the endpoint catalog):

```
catalog/partial_*.json   ── all 252 routes (method, path, roles, params), per service
        │                   merge_catalog.py → endpoints.json (generated each run)
        │
        ├─ lib/auth.py        mint + cache one Keycloak token per role
        ├─ lib/fixtures.py    seed 2 courses (A + B), phases, a student, kc roles
        │
        ├─ k6/scenario.js     smoke / load / spike / soak  (per-endpoint metrics)
        ├─ k6/exhaustion.js   connection flood + multi-MB / deeply-nested bodies
        ├─ fuzz/fuzz.py       inputs, auth matrix, cross-course IDOR, files, slowloris
        │
        └─ report/build_report.py  → report.md / report.html / findings.json
```

`run.sh` orchestrates: **preflight → seed fixtures → mint tokens → k6 scenarios →
fuzz → build report → teardown**.

### Why an isolated stack

The suite runs against its own Docker Compose project, **`prompt-stress`**, defined by
`api-stress/docker-compose.stress.yml` (an override of the repo `docker-compose.yml`)
plus `api-stress/stress.env`. It exists so the suite never collides with a PROMPT
stack you already have running:

- Every container is renamed `stress-*` and every host port is offset by **+10000**
  (core `18089`, team-allocation `18083`, self-team `18084`, assessment `18085`,
  template `18086`, interview `18087`, certificate `18088`, Keycloak `18081`, S3 `18334`).
- Postgres/SeaweedFS data live in named volumes (the worktree stays clean).
- Compose **concatenates** `ports`/`volumes` across files, so the override uses the
  `!override` YAML tag to *replace* them.
- Sentry is disabled; Keycloak's access-token lifespan is raised to 1h (a full run
  outlasts the 300s default).

## Running it

Prerequisites: Docker, [`k6`](https://k6.io/) (`brew install k6`), Python 3 with
`httpx` (`python3 -m venv api-stress/.venv && api-stress/.venv/bin/pip install -r api-stress/requirements.txt`).

```bash
# 1. local env file (gitignored; run.sh also auto-creates it)
cp api-stress/stress.env.example api-stress/stress.env

# 2. bring up the isolated stack
docker compose --env-file api-stress/stress.env \
  -f docker-compose.yml -f api-stress/docker-compose.stress.yml \
  -p prompt-stress up -d \
  db db-team-allocation db-self-team-allocation db-assessment db-template-server \
  db-interview db-certificate keycloak-db keycloak \
  seaweedfs-master seaweedfs-volume seaweedfs-filer seaweedfs-s3 \
  server-core server-team-allocation server-self-team-allocation server-assessment \
  server-template server-interview server-certificate

# 3. run the suite
api-stress/run.sh                    # full run, medium intensity
api-stress/run.sh --smoke-only       # quick reachability + latency baseline
api-stress/run.sh --no-exhaustion    # skip the host-stressing lane (good for repeats)
api-stress/run.sh --intensity brutal # all-out

# 4. teardown when done
docker compose -p prompt-stress down -v
```

Output lands in `api-stress/reports/<timestamp>/` (gitignored). Open `report.md` first.

## How findings are prioritized

| Priority | Meaning |
|----------|---------|
| **P0** | Auth bypass (no / forged / `alg=none` token accepted), cross-course IDOR/BOLA, or a service that went DOWN. |
| **P1** | 5xx on a *valid* request, load-induced degradation, oversized-body crash, slowloris. |
| **P2** | 5xx on malformed input / non-existent target (missing 404 handling), individually slow endpoints. |

Each finding carries the endpoint, method, observed vs expected, a `servers/<service>/`
"where to look" pointer, the evidence body, and a ready-to-paste `curl` repro.

## How the suite avoids false signal

These are deliberate design choices worth preserving when you extend the suite:

- **Non-destructive smoke.** In smoke, write/delete endpoints resolve their path ids
  to a *non-existent* id, so a full walk can never corrupt the shared fixtures
  (e.g. `DELETE /courses/{uuid}`). Reads use real fixture ids for a true latency
  baseline. Real write bodies are exercised by the fuzzer instead.
- **Meltdown-only abort.** k6 aborts a run only on a real 5xx/timeout meltdown
  (`fatal_rate`), never on 4xx - many catalog reads legitimately 404 on
  non-seeded sub-resources, and that must not look like the server falling over.
- **Privilege-escalation, not "wrong role".** The auth probe uses the *lowest*-
  privilege actor (a course student) against endpoints that exclude `CourseStudent`.
  It deliberately does **not** probe with admin/lecturer (global privileged roles
  legitimately allowed broadly), nor on `my-*`/`self` endpoints (caller-scoped) -
  those produce false positives.
- **IDOR via a second course.** Fixtures seed course **B** in addition to A so the
  fuzzer can prove cross-course reads with a token scoped only to A.

## Extending the suite

- **Add or fix a route:** edit the relevant `api-stress/catalog/partial_<service>.json`, then
  `python3 api-stress/catalog/merge_catalog.py` to regenerate `api-stress/catalog/endpoints.json`. The
  smoke run validates the catalog against the live server (a wrong path shows up as
  a 404).
- **Tune intensity:** the `INTENSITY` presets (`gentle`/`medium`/`brutal`) live at
  the top of `k6/scenario.js` and `k6/exhaustion.js`.
- **Add a fuzz axis:** add a method to `fuzz/fuzz.py` and call it from `run()`.
- **Change resolution of a path param:** edit `resolveParam` in `k6/lib/catalog.js`
  (k6) and `Fuzzer.resolve` in `fuzz/fuzz.py` (keep them in sync).

## Troubleshooting

- **Keycloak crash-loops on import** ("Duplicate resource error" /
  `UK_J3RWUVD56ONTGSUHOGM184WW2`): the repo pins Keycloak `26.6.3`, which fails to
  import the committed realm on a fresh DB. The override pins **`26.4.7`**, which
  imports cleanly. If you wiped the keycloak volume, recreate it:
  `docker compose -p prompt-stress up -d keycloak-db keycloak`.
- **Everything 401 mid-run / IDOR count drops to 0:** tokens expired. `run.sh` raises
  the realm `accessTokenLifespan` to 1h at preflight and re-mints before fuzzing; if
  you call the Python tools directly, mint fresh tokens first (`lib/auth.py`).
- **All servers return connection-refused (status 0) at once:** the Docker host
  itself went down, not the API. The all-out `exhaustion` lane (connection flood +
  large bodies), run back-to-back, can exhaust the local Docker VM. Restart it and
  use `--no-exhaustion` for repeat runs.
- **Port already allocated:** another PROMPT stack is using a `+10000` port. Adjust
  the published ports in `docker-compose.stress.yml` and the matching `services.json`.

See `api-stress/README.md` for the same content in repo form.
