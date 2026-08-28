---
name: verify
description: Run PROMPT 2.0 and observe a change in the real UI or API. Boots the seeded e2e stack (core client + server + Keycloak + Postgres + SeaweedFS) without the Playwright suite and drives it from inside the compose network. Use when verifying a client or core-server change at its surface rather than through tests.
---

# Verify a change against a running PROMPT stack

The `docker-compose.e2e.yml` stack is the fastest handle: it boots the core client, the
core server, every phase service, Keycloak with an imported realm, and a fully seeded
database (`e2e/seed/e2e_seed.sql`) on ports that coexist with a dev stack — client 4000,
core API 18090, Keycloak 18081.

## Boot the app without running tests

The Makefile exports `.env` / `.env.dev`, whose host-mode overrides outrank
`--env-file e2e/.env.e2e`, so unset those keys (this is what `make test-e2e-*` does):

```bash
env -u DB_CORE_HOST -u DB_CORE_PORT -u DB_CORE_USER -u DB_CORE_PASSWORD -u DB_CORE_NAME \
    -u KEYCLOAK_HOST -u CORE_HOST -u SERVER_ADDRESS \
  docker compose -f docker-compose.e2e.yml --env-file e2e/.env.e2e up -d client-core server-core
```

`depends_on` pulls in the databases, Keycloak and SeaweedFS. A cold build is ~10 min;
afterwards startup is under a minute. Readiness:

```bash
curl -s -o /dev/null -w '%{http_code}\n' http://localhost:18090/api/hello   # core API
curl -s -o /dev/null -w '%{http_code}\n' http://localhost:4000              # core client
```

Tear down with `make test-e2e-down` (removes volumes, so the seed is fresh next boot).

## Drive it from inside the compose network, not from the host

The client's runtime config points the browser at `http://keycloak:8080`, a name only
Docker DNS resolves — a host browser dies on `ERR_NAME_NOT_RESOLVED` at the login
redirect, and a Keycloak token minted via `localhost:18081` is rejected by the server as
`{"error":"Invalid token"}` (issuer mismatch). Run driver scripts in the `e2e-runner`
container instead: it sits on the network and ships Playwright plus Chromium.

```bash
env -u DB_CORE_HOST -u DB_CORE_PORT -u DB_CORE_USER -u DB_CORE_PASSWORD -u DB_CORE_NAME \
    -u KEYCLOAK_HOST -u CORE_HOST -u SERVER_ADDRESS \
  docker compose -f docker-compose.e2e.yml --env-file e2e/.env.e2e \
  run --rm --no-deps -e SHOTS=/work/test-results \
  -v "$PWD/e2e/drive.mjs:/work/drive.mjs" e2e-runner node drive.mjs
```

Inside the container use `BASE_URL=http://client-core`, `http://server-core:8080` for the
core API and `http://keycloak:8080` for tokens. Screenshots written to
`/work/test-results` land in `e2e/test-results/` on the host. Delete the driver script
from the repo when done — `e2e/` is checked in.

Log in the way the suite does (`e2e/src/pages/LoginPage.ts`): go to `/management`, fill
`#username` / `#password`, click `#kc-login`, then wait for the `jwt_token` localStorage
key. Seeded accounts are in `e2e/src/data/roles.ts` — username equals password
(`lecturer`, `admin`, `student`, …). Seeded courses, phases and participants are in
`e2e/src/data/constants.ts`.

## Exercising real write paths

A lecturer token from the container drives any core API directly, which beats hand-editing
the seed when you need a state the fixtures don't have:

```js
const body = new URLSearchParams({ client_id: 'prompt-client', username: 'lecturer',
  password: 'lecturer', grant_type: 'password' })
const { access_token } = await (await fetch(
  'http://keycloak:8080/realms/prompt/protocol/openid-connect/token',
  { method: 'POST', body })).json()
```

## Gotchas

- `/management/**` and `/apply/:id/authenticated` sit behind `RequireAuth`, which asks
  Keycloak for a login when no session can be restored; the public pages never redirect,
  so enter through a management route when you need to log in.
- The console logs a `[prompt-shared-state] Missing or invalid env keys` warning and
  404s for optional phase assets on every page — pre-existing noise, not your change.
- The seed is a `pg_dump` file loaded by `initdb` with `ON_ERROR_STOP`: one duplicate key
  aborts it and the whole stack fails to boot with `container prompt-e2e-db exited (3)`.
  Validate a seed edit on its own before a full run:
  `docker run --rm -e POSTGRES_PASSWORD=x -e POSTGRES_DB=prompt \
   -v "$PWD/e2e/seed/e2e_seed.sql:/docker-entrypoint-initdb.d/e2e_seed.sql:ro" postgres:15.18-alpine`
- Reusing images with `SKIP_BUILD=1` keeps the *baked-in* copy of `e2e/src` and
  `e2e/tests`; the seed is a bind mount and updates without a rebuild. Edited constants
  therefore need a rebuild, or the browser navigates to stale ids.
