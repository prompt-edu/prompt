---
sidebar_position: 10
title: Infrastructure Setup Phase
description: How the infrastructure setup phase provisions external resources per team or student.
---

# Infrastructure Setup Phase

## Overview

The **Infrastructure Setup Phase** automates the creation of external resources (GitLab groups,
Slack channels, Outline collections, Rancher projects, Keycloak groups) per team or per student. It
replaces the manual, error-prone setup that large courses would otherwise repeat every semester.

Teams come from a preceding Team Allocation phase through the phase configurator. The service then
creates and configures external resources from instructor-defined name templates and permission
mappings.

---

## Architecture

### Backend: `servers/infrastructure_setup/` (Go, Gin, port 8091)

```text
servers/infrastructure_setup/
├── main.go                            # routes, provider registry, startup recovery
├── sqlc.yaml
├── db/
│   ├── migration/
│   │   ├── 0001_schema.up.sql         # enums and tables
│   │   ├── 0002_partial_status.up.sql # adds the 'partial' resource status
│   │   └── 0003_resource_config_identity.up.sql
│   ├── query/                         # sqlc sources
│   └── sqlc/                          # generated, committed
├── database_dumps/base.sql            # schema for testcontainers-based tests
├── encryption/aes.go                  # AES-256-GCM credential encryption
├── provider/
│   ├── interface.go                   # Provider interface and shared types
│   ├── gitlab/ slack/ outline/ rancher/ keycloak/
├── providerconfig/                    # credentials CRUD, validation, provider metadata
├── resourceconfig/                    # what to provision, per phase
├── phaseconfig/                       # the phase's own settings (semester tag)
├── execution/                         # instance lifecycle: trigger, worker, templates
└── copy/                              # PhaseCopyHandler and the SDK config endpoint
```

### Frontend: `clients/infrastructure_setup_component/` (React, TypeScript, port 3011)

- **SetupConfigPage** — the phase's semester tag.
- **ProvidersPage** — configure and validate credentials per provider. A provider without
  credentials is marked *credentials required*.
- **ResourceConfigPage** — CRUD for resource configs: provider, resource type, scope, name
  template, permission mapping.
- **ExecutionPage** — trigger provisioning and monitor instances. Polls every 3s while any
  instance is `pending` or `in_progress`.

---

## Providers

| Provider | Resource type | Credentials | Notes |
|---|---|---|---|
| GitLab | `group` | `base_url`, `private_token`, optional `parent_group_id` | Members are added through the group **invitations** endpoint, which works with a non-admin PAT and covers users who have not signed in yet. With a parent configured, only that parent's subgroups are searched. |
| Slack | `channel` | `bot_token` | Creates private channels. Requests are form-encoded, which is what the Web API documents for its read methods. |
| Outline | `collection` | `api_key`, optional `base_url` | `collections.list` is paginated, so an existing collection is never duplicated. |
| Rancher | `project` | `rancher_url`, `access_key`, `secret_key`, `cluster_id` | Users are resolved through the principals search endpoint and confirmed against the requested address; the returned principal ID is used as-is. |
| Keycloak | `group` | `keycloak_url`, `realm`, `client_id`, `client_secret` | The service account needs `manage-groups` and `view-users`, not `realm-admin`. |

All providers are **idempotent**: an existing resource with the same name is adopted rather than
recreated. Because of that, PROMPT cannot know whether it owns a resource, which is why it never
deletes one (see [Delete semantics](#delete-semantics)).

A provider declares which resource kinds it supports; the API rejects any other value, and the
resource config dialog offers only the supported kinds.

---

## Database schema

```sql
CREATE TYPE provider_type AS ENUM ('gitlab', 'slack', 'outline', 'rancher', 'keycloak');
CREATE TYPE resource_scope AS ENUM ('per_team', 'per_student');
CREATE TYPE resource_status AS ENUM ('pending', 'in_progress', 'created', 'failed', 'partial');

-- The phase's own settings.
CREATE TABLE course_phase_config (
    course_phase_id uuid PRIMARY KEY,
    semester_tag    text NOT NULL DEFAULT ''
);

-- Encrypted provider credentials, one row per (phase, provider type).
CREATE TABLE provider_config (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    course_phase_id uuid NOT NULL,
    provider_type   provider_type NOT NULL,
    credentials     bytea NOT NULL DEFAULT ''::bytea,  -- 12-byte nonce || ciphertext
    CONSTRAINT uq_provider_config_phase_type UNIQUE (course_phase_id, provider_type)
);

-- What to create, with name templates and permission mappings.
CREATE TABLE resource_config (
    id                    uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    course_phase_id       uuid NOT NULL,
    provider_type         provider_type NOT NULL,
    resource_type         text NOT NULL,
    scope                 resource_scope NOT NULL,
    name_template         text NOT NULL,
    permission_mapping    jsonb NOT NULL DEFAULT '{}',
    resource_extra_config jsonb NOT NULL DEFAULT '{}',
    created_at            timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_resource_config_provider
        FOREIGN KEY (course_phase_id, provider_type)
        REFERENCES provider_config (course_phase_id, provider_type) ON DELETE CASCADE,
    CONSTRAINT uq_resource_config_identity
        UNIQUE (course_phase_id, provider_type, resource_type, scope, name_template)
);

-- One row per provisioned resource, with its lifecycle status.
CREATE TABLE resource_instance (
    id                      uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    resource_config_id      uuid NOT NULL REFERENCES resource_config(id) ON DELETE CASCADE,
    course_phase_id         uuid NOT NULL,
    team_id                 uuid,
    course_participation_id uuid,
    status                  resource_status NOT NULL DEFAULT 'pending',
    external_id             text,
    external_url            text,
    error_message           text,
    created_at              timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at              timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- No duplicate non-failed instances per (config, team) and (config, student).
CREATE UNIQUE INDEX uq_resource_instance_team
    ON resource_instance (resource_config_id, team_id)
    WHERE team_id IS NOT NULL AND status != 'failed';

CREATE UNIQUE INDEX uq_resource_instance_student
    ON resource_instance (resource_config_id, course_participation_id)
    WHERE course_participation_id IS NOT NULL AND status != 'failed';
```

### Instance statuses

| Status | Meaning |
|---|---|
| `pending` | Queued, not yet picked up. |
| `in_progress` | Claimed by a worker. |
| `created` | The resource exists and every member was granted access. |
| `partial` | The resource exists, but at least one member could not be added. `error_message` lists them, `external_id`/`external_url` are kept. Retryable. |
| `failed` | The resource could not be created. `error_message` explains why. Retryable. |

---

## Credential encryption

Credentials are stored as AES-256-GCM encrypted `bytea`, formatted as `12-byte nonce || ciphertext`
with a fresh random nonce per call. The key comes from `ENCRYPTION_KEY`, a base64-encoded 32-byte
value, and is validated at startup so a misconfigured deployment fails immediately rather than on
the first credential write.

```bash
openssl rand -base64 32
```

The API never returns credentials. `GET /provider-configs` exposes only the provider type and a
`configured` boolean. Rotating the key requires re-encrypting every `provider_config.credentials`
row by hand.

---

## Name templates

| Placeholder | Description |
|---|---|
| `{{teamName}}` | Team name (`per_team` scope) |
| `{{semesterTag}}` | The phase's semester tag (e.g. `ss25`) |
| `{{studentName}}` | Full name (`per_student` scope) |
| `{{studentFirstName}}` | First name (`per_student` scope) |
| `{{studentLastName}}` | Last name (`per_student` scope) |
| `{{studentEmail}}` | Email address (`per_student` scope) |
| `{{studentLogin}}` | University login (`per_student` scope) |

`{{.TeamName}}`, `{{.StudentLogin}}`, `{{.Semester}}` and `{{.SemesterTag}}` are accepted as
aliases. Anything else is **rejected**: when the resource config is saved, and again before the
worker calls a provider, so an unresolved `{{...}}` can never end up in the name of a real
resource.

Values are sanitized generically and then again per provider (a Slack channel name is slugified and
truncated to 80 characters, a GitLab path is turned into a slug). A name that sanitizes to an empty
identifier is an error rather than a request.

---

## Execution

`POST .../execute`:

1. Loads the phase's resource configs and refuses to run if any of their providers has no
   credentials (the state a copied phase starts in).
2. Resolves targets from core, once per scope. This is HTTP, so it happens before the transaction.
3. Opens a transaction, takes a per-phase advisory lock, checks that no non-terminal instance
   exists, and inserts the pending rows. The lock makes the check and the insert atomic: a second
   trigger arriving at the same time gets **409** instead of creating a duplicate run.
4. Commits, then starts the background worker.

The worker claims all pending instances of the phase in a single
`UPDATE ... FOR UPDATE SKIP LOCKED`, so two workers can never process the same row. It then
processes up to 5 instances concurrently. Each one resolves its name, calls
`provider.CreateResource`, and is recorded as `created`, `partial` or `failed`. Provider calls are
retried up to 3 times with exponential backoff (1s, 2s, 4s) plus jitter.

**Startup recovery:** anything left `in_progress` by a crash is reset to `pending` at boot.

**Retry** (`POST .../instances/:instanceID/retry`) accepts `failed` and `partial` instances. An
unknown instance is a 404; one that is `created` or already queued is a 409. Since providers are
idempotent, retrying a `partial` instance heals it once the missing users exist upstream.

---

## Delete semantics

`DELETE .../instances/:instanceID` removes **only** the PROMPT row. The external resource is never
touched, and the confirmation dialog says so.

This is deliberate. Providers adopt existing resources by name, so PROMPT cannot tell whether a
group it points at was created for this course or already belonged to someone else; deleting could
remove a shared resource. Deleting a resource config or a provider config cascades to the instance
rows in SQL, which would bypass any per-instance cleanup anyway. `external_id` and `external_url`
are kept on the instance so the resource can be found and removed by hand.

---

## Phase copy

Copying a phase carries over, in one transaction:

- `course_phase_config`, including the semester tag. Without this row the config endpoint reports
  nothing as configured at all.
- `provider_config` rows **with empty credentials**. The instructor must re-enter the secrets.
- `resource_config` rows. Copying the same source twice does not duplicate them.

`resource_instance` rows are not copied; provisioning is triggered again on the new phase.

Until the credentials are re-entered, the copied phase reports `providerConfig: false`, its
providers show *credentials required*, resource configs cannot be created against them, and
triggering execution is refused.

---

## API reference

Phase-scoped routes live under `/infrastructure-setup/api/course_phase/:coursePhaseID` and require
`PromptAdmin` or `CourseLecturer`. These routes carry external credentials, so `CourseEditor` is
deliberately **not** granted access.

| Method | Path | Description |
|---|---|---|
| GET | `/setup-config` | Read the phase settings (semester tag) |
| PUT | `/setup-config` | Update the phase settings |
| GET | `/provider-configs` | List providers with a `configured` flag; credentials are never returned |
| PUT | `/provider-configs` | Create or replace credentials for one provider |
| DELETE | `/provider-configs/:providerType` | Remove credentials; resource configs and instances cascade |
| POST | `/provider-configs/:providerType/validate` | Test the stored credentials against the provider |
| GET | `/provider-configs/:providerType/fields` | Credential fields the provider needs |
| GET | `/provider-configs/:providerType/resource-types` | Resource kinds the provider supports |
| GET | `/resource-configs` | List resource configs |
| POST | `/resource-configs` | Create a resource config |
| GET | `/resource-configs/:resourceConfigID` | Read a resource config |
| PUT | `/resource-configs/:resourceConfigID` | Update a resource config |
| DELETE | `/resource-configs/:resourceConfigID` | Delete a resource config |
| GET | `/instances` | List instances and their status |
| POST | `/execute` | Create pending instances and start provisioning (202, or 409 while a run is active) |
| POST | `/instances/:instanceID/retry` | Re-queue a failed or partial instance |
| DELETE | `/instances/:instanceID` | Delete the PROMPT row; the external resource is untouched |
| GET | `/config` | Readiness flags consumed by core |

Two routes are not phase-scoped:

| Method | Path | Description |
|---|---|---|
| POST | `/infrastructure-setup/api/copy` | Phase copy handler |
| GET | `/infrastructure-setup/api/info` | Public service info consumed by the management console |

---

## Environment variables

| Variable | Required | Description |
|---|---|---|
| `ENCRYPTION_KEY` | Yes | Base64-encoded 32-byte AES key. Validated at startup. |
| `DB_INFRASTRUCTURE_SETUP_HOST` / `DB_HOST_INFRASTRUCTURE_SETUP` | Yes | PostgreSQL host (the compose file maps the first onto the second) |
| `DB_PORT_INFRASTRUCTURE_SETUP` | Yes | PostgreSQL port (5440 on the host in local dev) |
| `DB_NAME`, `DB_USER`, `DB_PASSWORD` | Yes | Database credentials |
| `KEYCLOAK_HOST`, `KEYCLOAK_REALM_NAME` | Yes | Auth middleware configuration |
| `CORE_HOST` | Yes | Client origin, used for CORS |
| `SERVER_CORE_HOST` | Yes | Core service URL for inter-service calls |
| `INFRASTRUCTURE_SETUP_HOST` | Yes (client) | Where the micro-frontend sends its requests |
| `SERVER_ADDRESS` | No | Bind address (default `localhost:8091`) |
| `SENTRY_DSN_INFRASTRUCTURE_SETUP`, `SENTRY_ENABLED` | No | Error reporting |
| `DEBUG` | No | Debug logging |

---

## Tests

```bash
make test-infrastructure-setup
```

DB-backed tests use `testcontainers-go` with `database_dumps/base.sql`, which must be kept in step
with the migrations.

- **encryption** — round-trip, wrong key, nonce uniqueness, corrupted ciphertext.
- **execution/template** — every placeholder and alias, sanitization, and rejection of unknown
  placeholders.
- **execution/worker** — created, partial (with the external ID preserved), retry to exhaustion,
  success after a transient failure, a vanished target, an unresolvable template, and startup
  recovery. Two workers racing on one phase call the provider exactly once.
- **execution/service** — instances per team and per student, 409 for a second trigger, two
  concurrent triggers on separate connections creating one run, and retry returning 404/409.
- **execution/target_resolver** — malformed upstream payloads are skipped, not fatal.
- **provider/**\* — an `httptest` server per provider covering create, idempotency, member handling
  and each provider-specific fix.
- **providerconfig / resourceconfig** — credentials are encrypted at rest and never returned,
  non-string and unknown credential fields are rejected, and resource types are validated against
  the provider.
- **copy / phaseconfig** — a copied phase reports its semester tag and resource configs but not its
  providers, and copying twice does not duplicate anything.

End-to-end coverage lives in `e2e/tests/infrastructure-setup/`.

---

## Infrastructure

Compose (`docker-compose.yml`) adds `server-infrastructure-setup` (host port 8091),
`client-infrastructure-setup` and `db-infrastructure-setup` (host port 5440). The server waits for
both its database and Keycloak to become healthy, since it performs OIDC discovery at startup.

Core registers the phase in `servers/core/coursePhaseType/initializeTypes.go` as
`Infrastructure Setup`, declaring the team and team-allocation inputs as required, with a base URL
of `http://localhost:8091/infrastructure-setup/api` in dev and `{CORE_HOST}/infrastructure-setup/api`
in production.
