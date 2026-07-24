# Quickstart

Get a local PROMPT 2.0 stack running in a few minutes.

## 1. Environment setup

```bash
cp .env.template .env
cp .env.dev.template .env.dev
```

`.env` holds Docker hostnames; `.env.dev` overrides them with localhost values for local
development. For Keycloak/auth specifics, see the `keycloak-local-setup` skill or
`docs/contributor/setup.md`.

## 2. Install client dependencies

```bash
cd clients && yarn install && cd ..
```

## 3. Start the servers (Docker)

```bash
docker compose up server-core server-assessment server-team-allocation
```

This also brings up the required databases and Keycloak. Add or drop `server-*` services as
needed (see `docker-compose.yml` for the full list).

## 4. Start the clients (yarn)

Run the core shell plus whichever component you're working on, each in its own terminal:

```bash
cd clients/core && yarn dev                   # app shell, port 3000
cd clients/assessment_component && yarn dev   # port 3007
cd clients/team_allocation_component && yarn dev  # port 3008
```

Or start everything at once: `make clients` (or `cd clients && yarn dev`).

## 5. Open the app

Visit <http://localhost:3000> and log in with a local Keycloak user.

## Useful commands

```bash
make lint      # lint everything
make test      # run tests
make db-down   # stop database and Keycloak
```
