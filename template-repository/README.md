# Course Phase Template Repository — Staging Area

This directory stages everything needed to create the standalone GitHub template repository
**`prompt-edu/prompt-course-phase-template`** for building custom PROMPT course phases outside
this monorepo (issue #1864).

Nothing in here is wired into the monorepo build — it only becomes active when the migration in
[`MIGRATION.md`](./MIGRATION.md) is executed. Until then, `clients/example_component` and
`servers/example_server` remain the single source of truth for the example phase; keep changes
to the example in those directories, not here.

## Contents

| Path                          | Becomes (in the new repo)                                |
| ----------------------------- | -------------------------------------------------------- |
| `MIGRATION.md`                | — (the runbook for performing the move)                   |
| `init.sh`                     | `init.sh` — renames `example` placeholders after "Use this template" |
| `repo/README.md`              | `README.md`                                               |
| `repo/gitignore`              | `.gitignore`                                              |
| `repo/env.template`           | `.env.template`                                           |
| `repo/docker-compose.yml`     | `docker-compose.yml`                                      |
| `repo/docker-compose.prod.yml`| `docker-compose.prod.yml`                                 |
| `repo/client.package.json`    | `client/package.json` (standalone deps, replaces the workspace one) |
| `repo/client.Dockerfile`      | `client/Dockerfile` (self-contained, no `prompt-clients-base`) |
| `repo/client.nginx.conf`      | `client/nginx.conf`                                       |
| `repo/server.Dockerfile`      | `server/Dockerfile` (copy of `servers/Dockerfile`)        |
| `repo/workflows/quality.yml`  | `.github/workflows/quality.yml`                           |
| `repo/workflows/build-and-push.yml` | `.github/workflows/build-and-push.yml`              |

The `client/` and `server/` directories themselves are NOT staged here — they are created at
migration time from `clients/example_component` and `servers/example_server` (with git history,
see the runbook).

Files that normally start with a dot are staged without it (`gitignore`, `env.template`) so they
stay inert in this repo.
