---
sidebar_position: 3
---

# Setup Guide

Welcome to the **Prompt** setup guide! In this document, you will learn how to configure and run the development environment of the **Prompt** application. For a dockerized demo or production deployment, see the [Production Setup](/admin/productionSetup) guide instead.

## Overview

**Prompt** is composed of:

- A **Golang** backend (using [Gin](https://gin-gonic.com/), [SQLC](https://docs.sqlc.dev/), and [PostgreSQL](https://www.postgresql.org/)). It consists of the core server plus one service per course phase, each with its own database.
- A **TypeScript/React** client that runs in the browser and is structured as a core frontend dynamically loading multiple microfrontends. Each microfrontend typically represents one course phase.
- **Keycloak** for authentication and **SeaweedFS** (S3-compatible object storage) for uploaded files.

Almost every command below is a Makefile target. Run `make help` to list all of them.

## Prerequisites

:::warning
Read the Contributor Guidelines

![notPass](https://media3.giphy.com/media/v1.Y2lkPTc5MGI3NjExeHE5d3drMjQ2b2hidGdlYzM3azcyanhvZnZpNWF6bGl4cGdidGhvdCZlcD12MV9pbnRlcm5hbF9naWZfYnlfaWQmY3Q9Zw/xULW8MYvpNOfMXfDH2/giphy.gif)

Before you continue, please make sure to read the [contributor guidelines](./index.md).

:::

Before you can build and run **Prompt**, you must install and configure the following dependencies on your machine:

1. **Docker**
   - Install [Docker](https://docs.docker.com/get-started/get-docker/) including `docker compose`.
   - The databases, Keycloak, and the object storage all run as containers, so you do not need a local PostgreSQL installation.

2. **Golang**
   - Install [Go](https://go.dev/doc/install).
   - The modules require **Go 1.27** or newer (see the `go` directive in `servers/core/go.mod`).

3. **golang-migrate**
   - Install [golang-migrate](https://github.com/golang-migrate/migrate).
   - The servers invoke the `migrate` binary on startup to apply pending migrations, so it must be on your `PATH`.

4. **sqlc**
   - Install [sqlc](https://docs.sqlc.dev/en/latest/overview/install.html).
   - Only needed when you change SQL queries or migrations and regenerate the typed code with `make sqlc`.

5. **Node.js**
   - Install [Node.js](https://nodejs.org/en) **24.21.0** or newer (the version used by CI and the Docker images).
   - Node.js is required to compile and run the React client application.

6. **Yarn**
   - We use **Yarn 4** (pinned via `packageManager` in `clients/package.json`) to manage front-end dependencies.
   - First, install Corepack:
     - **macOS (Homebrew users)**: Homebrew strips Corepack from Node.js, so install it separately:

       ```bash
       brew install corepack
       ```

     - **Other platforms**: Corepack is included with Node.js 16.9+. If missing, install via npm:

       ```bash
       npm install -g corepack
       ```

   - Enable Corepack by running:

     ```bash
     corepack enable
     ```

   - This ensures you can run Yarn.

7. **make**
   - Ships with macOS (Xcode command line tools) and most Linux distributions.
   - On Windows, use WSL. The Makefile targets are thin wrappers, so you can also run the underlying commands directly.

## Development Environment

### 1. Clone the Repository

Clone (or download) the Prompt repository to your local machine:

```bash
git clone https://github.com/prompt-edu/prompt.git
cd prompt
```

### 2. Create Your Environment Files

Two templates live in the repository root:

- **`.env.template` → `.env`**: the base configuration. It is read by `docker compose` for variable substitution and by the Makefile, which exports the values to the servers it starts. Hostnames point at the Docker service names (for example `db`, `seaweedfs-s3`).
- **`.env.dev.template` → `.env.dev`**: localhost overrides for running the Go servers *outside* Docker. `make server` loads `.env` first and then `.env.dev`, so these values win.

```bash
cp .env.template .env
cp .env.dev.template .env.dev
```

Both `.env` and `.env.dev` are gitignored — never commit them. The defaults work out of the box for local development; the only value you have to fill in yourself is `KEYCLOAK_CLIENT_SECRET` (step 4).

The [Adjust Environment Variables](/admin/productionSetup#31-adjust-environment-variables) section of the production setup guide explains what the individual variables do.

:::note
The blocks below are copies of the templates for reference. The files in the repository are the source of truth.
:::

<details>
<summary>Copy of <code>.env.template</code></summary>

```bash
# ============================================================================
# PROMPT2 Environment Configuration
# ============================================================================
# This file contains all environment variables for the PROMPT2 application
# Copy this file to .env and adjust values for your environment

# ============================================================================
# GENERAL APPLICATION SETTINGS
# ============================================================================

# Application environment (development, staging, production)
ENVIRONMENT=development

# Core application host URL (without protocol for development, with for production)
CORE_HOST=http://localhost:3000

# University/Organization branding
CHAIR_NAME_SHORT=Applied Education Technologies
CHAIR_NAME_LONG=TUM Research Group for Applied Education Technologies

# Server configuration
SERVER_ADDRESS=0.0.0.0:8080

# ============================================================================
# AUDIT LOG (optional; off by default)
# ============================================================================
# Enable the centralized audit log. When unset/false, nothing is captured.
AUDIT_ENABLED=
# Retention window in days. Unset => entries are never pruned (kept forever).
AUDIT_RETENTION_DAYS=
# Read by core: the shared secrets it accepts, one per reporting service.
# Format: service1:key1,service2:key2 (two values per service allowed for rotation).
# The service names are the ones the services report on their info endpoint:
# assessment, certificate, example-service, interview, presentation,
# self-team-allocation, team-allocation, intro-course.
# Example: assessment:<key>,interview:<key>,team-allocation:<key>
# Externally deployed phases (intro-course among them) have no container here.
# List their keys anyway so they can report to this core.
AUDIT_INGEST_KEYS=
# Read by a phase process: its own ingest key. Under Docker Compose it is set
# per container from the AUDIT_INGEST_KEY_<SERVICE> variables below. In host
# mode the Makefile exports this one value to every server it starts, so
# `make servers` can only report as a single service; use it with
# `make server-<name>` when reporting from one phase at a time.
AUDIT_INGEST_KEY=
# Per-service ingest keys handed to the phase containers by Docker Compose.
# Each value must also appear in AUDIT_INGEST_KEYS above, under that service's
# reported name (for example AUDIT_INGEST_KEY_EXAMPLE_SERVER belongs to
# example-service).
AUDIT_INGEST_KEY_ASSESSMENT=
AUDIT_INGEST_KEY_CERTIFICATE=
AUDIT_INGEST_KEY_EXAMPLE_SERVER=
AUDIT_INGEST_KEY_INTERVIEW=
AUDIT_INGEST_KEY_PRESENTATION=
AUDIT_INGEST_KEY_SELF_TEAM_ALLOCATION=
AUDIT_INGEST_KEY_TEAM_ALLOCATION=

# ============================================================================
# CORE DATABASE CONFIGURATION
# ============================================================================
# Main application database settings

DB_CORE_HOST=db
DB_CORE_PORT=5432
DB_CORE_NAME=prompt
DB_CORE_USER=prompt-postgres
DB_CORE_PASSWORD=prompt-postgres

# ============================================================================
# TEAM ALLOCATION DATABASE CONFIGURATION
# ============================================================================
# Database for team allocation functionality

DB_HOST_TEAM_ALLOCATION=db-team-allocation
DB_PORT_TEAM_ALLOCATION=5432
DB_TEAM_ALLOCATION_NAME=prompt
DB_TEAM_ALLOCATION_USER=prompt-postgres
DB_TEAM_ALLOCATION_PASSWORD=prompt-postgres

# ============================================================================
# SELF TEAM ALLOCATION DATABASE CONFIGURATION
# ============================================================================
# Database for self team allocation functionality

DB_HOST_SELF_TEAM_ALLOCATION=db-self-team-allocation
DB_PORT_SELF_TEAM_ALLOCATION=5432
DB_SELF_TEAM_ALLOCATION_NAME=prompt
DB_SELF_TEAM_ALLOCATION_USER=prompt-postgres
DB_SELF_TEAM_ALLOCATION_PASSWORD=prompt-postgres

# ============================================================================
# ASSESSMENT DATABASE CONFIGURATION
# ============================================================================
# Database for assessment functionality

DB_HOST_ASSESSMENT=db-assessment
DB_PORT_ASSESSMENT=5432
DB_ASSESSMENT_NAME=prompt
DB_ASSESSMENT_USER=prompt-postgres
DB_ASSESSMENT_PASSWORD=prompt-postgres

# ============================================================================
# CERTIFICATE DATABASE CONFIGURATION
# ============================================================================
# Database for certificate functionality

DB_HOST_CERTIFICATE=db-certificate
DB_PORT_CERTIFICATE=5432
DB_CERTIFICATE_NAME=prompt
DB_CERTIFICATE_USER=prompt-postgres
DB_CERTIFICATE_PASSWORD=prompt-postgres

# ============================================================================
# INFRASTRUCTURE SETUP DATABASE CONFIGURATION
# ============================================================================
# Database for infrastructure setup functionality

DB_INFRASTRUCTURE_SETUP_HOST=db-infrastructure-setup
DB_INFRASTRUCTURE_SETUP_PORT=5432
DB_INFRASTRUCTURE_SETUP_NAME=prompt
DB_INFRASTRUCTURE_SETUP_USER=prompt-postgres
DB_INFRASTRUCTURE_SETUP_PASSWORD=prompt-postgres

# AES-256-GCM encryption key for provider credentials in the infrastructure setup service.
# Must be a base64-encoded 32-byte key. Generate with: openssl rand -base64 32
# Never rotate without re-encrypting existing provider_config rows.
# The value below is a placeholder so the service starts on a fresh checkout. It is
# committed, so it is public: the service refuses it unless DEBUG=true, and any
# deployment holding real provider credentials must set a key of its own.
ENCRYPTION_KEY=bG9jYWwtZGV2LWtleS1ub3QtYS1yZWFsLXNlY3JldCE=

# ============================================================================
# EXAMPLE SERVER DATABASE CONFIGURATION
# ============================================================================
# Database for example server functionality

DB_EXAMPLE_NAME=prompt
DB_EXAMPLE_PASSWORD=prompt-postgres
DB_EXAMPLE_USER=prompt-postgres
DB_HOST_EXAMPLE_SERVER=db-example-server
DB_PORT_EXAMPLE_SERVER=5432

# ============================================================================
# INTERVIEW DATABASE CONFIGURATION
# ============================================================================
# Database for interview scheduling functionality

DB_HOST_INTERVIEW=db-interview
DB_PORT_INTERVIEW=5432
DB_INTERVIEW_NAME=prompt
DB_INTERVIEW_USER=prompt-postgres
DB_INTERVIEW_PASSWORD=prompt-postgres

# ============================================================================
# PRESENTATION DATABASE CONFIGURATION
# ============================================================================
# Database for presentation scheduling, materials, and feedback

DB_HOST_PRESENTATION=db-presentation
DB_PORT_PRESENTATION=5432
DB_PRESENTATION_NAME=prompt
DB_PRESENTATION_USER=prompt-postgres
DB_PRESENTATION_PASSWORD=prompt-postgres

# ============================================================================
# KEYCLOAK DATABASE CONFIGURATION
# ============================================================================
# Database for Keycloak authentication service

KEYCLOAK_DB_HOST=keycloak-db
KEYCLOAK_DB_PORT=5432
KEYCLOAK_DB_NAME=keycloak
KEYCLOAK_DB_USER=keycloak
KEYCLOAK_DB_PASSWORD=keycloak

# ============================================================================
# KEYCLOAK AUTHENTICATION SETTINGS
# ============================================================================

# Keycloak server configuration
KEYCLOAK_HOST=http://localhost:8081
KEYCLOAK_REALM_NAME=prompt
KEYCLOAK_CLIENT_ID=prompt-server
KEYCLOAK_CLIENT_SECRET= # FIXME: Set your Keycloak client secret here
KEYCLOAK_ID_OF_CLIENT=a584ca61-fa83-4e95-98b6-c5f3157ae4b4
KEYCLOAK_AUTHORIZED_PARTY=prompt-client

# Keycloak admin credentials (development only)
KEYCLOAK_ADMIN=admin
KEYCLOAK_ADMIN_PASSWORD=admin

# ============================================================================
# EMAIL/SMTP CONFIGURATION
# ============================================================================
# Email server settings for notifications

SENDER_EMAIL=prompt@ase.cit.tum.de
SENDER_NAME=AET Mailing Service
SMTP_HOST=postfix
SMTP_PORT=25
SMTP_USERNAME=
SMTP_PASSWORD=

# ============================================================================
# DOCKER IMAGE TAGS
# ============================================================================
# Version tags for Docker images (used in production)

# Reported by locally started servers as their version on /info.
SERVER_IMAGE_TAG=local

SERVER_CORE_IMAGE_TAG=main
SERVER_TEAM_ALLOCATION_IMAGE_TAG=main
SERVER_SELF_TEAM_ALLOCATION_IMAGE_TAG=main
SERVER_ASSESSMENT_IMAGE_TAG=main
SERVER_INTERVIEW_IMAGE_TAG=main
SERVER_CERTIFICATE_IMAGE_TAG=main
SERVER_PRESENTATION_IMAGE_TAG=main
SERVER_INFRASTRUCTURE_SETUP_IMAGE_TAG=main

CORE_IMAGE_TAG=main
EXAMPLE_IMAGE_TAG=main
INTERVIEW_IMAGE_TAG=main
MATCHING_IMAGE_TAG=main
ASSESSMENT_IMAGE_TAG=main
DEVOPS_CHALLENGE_IMAGE_TAG=main
TEAM_ALLOCATION_IMAGE_TAG=main
SELF_TEAM_ALLOCATION_IMAGE_TAG=main
CERTIFICATE_IMAGE_TAG=main
PRESENTATION_IMAGE_TAG=main
INFRASTRUCTURE_SETUP_IMAGE_TAG=main

# ============================================================================
# SSL/TLS CONFIGURATION (Production)
# ============================================================================
# SSL settings for production deployments

ACME_EMAIL=your@email.de
SSL_MODE=disable

# ============================================================================
# MICRO-FRONTEND HOST URLS
# ============================================================================
# Runtime URLs for micro-frontends and status checks.
# In local Docker dev these point to localhost; in production they match CORE_HOST.

INTRO_COURSE_HOST=http://localhost:8082
DEVOPS_CHALLENGE_HOST=http://localhost:9010
TEAM_ALLOCATION_HOST=http://localhost:8083
SELF_TEAM_ALLOCATION_HOST=http://localhost:8084
ASSESSMENT_HOST=http://localhost:8085
EXAMPLE_HOST=http://localhost:8086
INTERVIEW_HOST=http://localhost:8087
CERTIFICATE_HOST=http://localhost:8088
PRESENTATION_HOST=http://localhost:8089
INFRASTRUCTURE_SETUP_HOST=http://localhost:8091

CORE_API_HOST=http://localhost:8080

# ============================================================================
# DEVELOPMENT SETTINGS
# ============================================================================
# Settings specific to development environment

# Enable debug mode for development
DEBUG=true

# Server core host for inter-service communication
SERVER_CORE_HOST=http://server-core:8080

# Git metadata defaults for local development
GITHUB_SHA=dev
GITHUB_REF=dev

# ============================================================================
# SENTRY CONFIGURATION
# ============================================================================
# Sentry DSN (Data Source Name) for error tracking and monitoring
# These are only used when SENTRY_ENABLED=true

SENTRY_ENABLED=false

SENTRY_DSN_ASSESSMENT=
SENTRY_DSN_CERTIFICATE=
SENTRY_DSN_CLIENT=
SENTRY_DSN_CORE=
SENTRY_DSN_EXAMPLE_SERVER=
SENTRY_DSN_INFRASTRUCTURE_SETUP=
SENTRY_DSN_INTERVIEW=
SENTRY_DSN_PRESENTATION=
SENTRY_DSN_SELF_TEAM_ALLOCATION=
SENTRY_DSN_TEAM_ALLOCATION=

# ============================================================================
# FILE STORAGE CONFIGURATION (S3-Compatible)
# ============================================================================
# File storage backend settings (works with SeaweedFS S3 gateway, AWS S3, MinIO, etc.)

S3_BUCKET=prompt-files
# S3 region (use 'us-east-1' for SeaweedFS)
S3_REGION=us-east-1
# Internal S3 endpoint for server-to-server communication
S3_ENDPOINT=http://seaweedfs-s3:8333
# Public S3 endpoint for presigned URLs (used by clients)
# Development: leave empty or use http://localhost:8334
# Production: https://s3.${CORE_HOST} (e.g., https://s3.prompt.aet.cit.tum.de)
S3_PUBLIC_ENDPOINT=
S3_ACCESS_KEY=admin
S3_SECRET_KEY=admin123
S3_FORCE_PATH_STYLE=true
# Presigned URL TTLs in seconds
S3_PRESIGN_UPLOAD_TTL_SECONDS=60
S3_PRESIGN_DOWNLOAD_TTL_SECONDS=30
MAX_FILE_UPLOAD_SIZE_MB=50
# Allowed file types for core's document uploads
# (comma-separated MIME types, leave empty for all)
ALLOWED_FILE_TYPES=application/pdf,image/jpeg,image/png,image/gif,application/zip,text/plain
# Allowed file types for presentation materials. Separate from the core list so slide decks
# can be permitted without also widening core's document uploads. This is the outer gate: each
# requested upload type additionally accepts only the formats that fit it, so narrowing this
# list can disable an upload type a lecturer selected. Leave it at the default to allow every
# type the presentation phase offers.
PRESENTATION_ALLOWED_FILE_TYPES=application/pdf,application/vnd.ms-powerpoint,application/vnd.openxmlformats-officedocument.presentationml.presentation,application/vnd.oasis.opendocument.presentation,application/msword,application/vnd.openxmlformats-officedocument.wordprocessingml.document,application/vnd.oasis.opendocument.text,image/png,image/jpeg,application/zip,application/x-zip-compressed,video/mp4

# ============================================================================
# CORE DATABASE ALIASES
# ============================================================================
# The core server reads generic DB_HOST/DB_PORT/etc. env vars,
# so we alias the DB_CORE_* vars for backward compatibility.

DB_HOST=${DB_CORE_HOST}
DB_PORT=${DB_CORE_PORT}
DB_NAME=${DB_CORE_NAME}
DB_USER=${DB_CORE_USER}
DB_PASSWORD=${DB_CORE_PASSWORD}

DB_HOST_INFRASTRUCTURE_SETUP=${DB_INFRASTRUCTURE_SETUP_HOST}
DB_PORT_INFRASTRUCTURE_SETUP=${DB_INFRASTRUCTURE_SETUP_PORT}
```

</details>

<details>
<summary>Copy of <code>.env.dev.template</code></summary>

```bash
# Development Environment Configuration
# Copy this file to .env.dev and fill in your values
#
# This file contains overrides for local development (running Go outside Docker)
# Usage: make server (automatically loads .env then .env.dev)

# Database - use localhost instead of Docker service names
DB_CORE_HOST=localhost
DB_HOST=localhost

# Team Allocation
DB_HOST_TEAM_ALLOCATION=localhost
DB_PORT_TEAM_ALLOCATION=5434
# Self Team Allocation
DB_HOST_SELF_TEAM_ALLOCATION=localhost
DB_PORT_SELF_TEAM_ALLOCATION=5436
# Assessment
DB_HOST_ASSESSMENT=localhost
DB_PORT_ASSESSMENT=5435
# Certificate
DB_HOST_CERTIFICATE=localhost
DB_PORT_CERTIFICATE=5439
# Example
DB_HOST_EXAMPLE_SERVER=localhost
DB_PORT_EXAMPLE_SERVER=5437
# Interview
DB_HOST_INTERVIEW=localhost
DB_PORT_INTERVIEW=5438
# Presentation
DB_HOST_PRESENTATION=localhost
DB_PORT_PRESENTATION=5440
# Infrastructure Setup
DB_HOST_INFRASTRUCTURE_SETUP=localhost
DB_PORT_INFRASTRUCTURE_SETUP=5441

# AES-256-GCM key for infrastructure setup provider credentials.
# Base64-encoded 32 bytes. Generate with: openssl rand -base64 32
# Local development placeholder: public, so the service only accepts it with DEBUG=true.
ENCRYPTION_KEY=bG9jYWwtZGV2LWtleS1ub3QtYS1yZWFsLXNlY3JldCE=

# Keycloak - use localhost
KEYCLOAK_HOST=http://localhost:8081

# Keycloak Client Secret - paste your secret here after generating it in Keycloak
# Go to: Keycloak Admin Console → Clients → prompt-server → Credentials → Regenerate
KEYCLOAK_CLIENT_SECRET=

# S3 - use localhost instead of Docker service name (seaweedfs-s3:8333 is Docker-only)
S3_ENDPOINT=http://localhost:8334
S3_PUBLIC_ENDPOINT=http://localhost:8334
S3_REGION=us-east-1
S3_BUCKET=prompt-files
S3_ACCESS_KEY=admin
S3_SECRET_KEY=admin123
S3_FORCE_PATH_STYLE=true

# Inter-service communication - use localhost instead of Docker service name
SERVER_CORE_HOST=http://localhost:8080

# Debug mode
DEBUG=true
```

</details>

### 3. Start the Infrastructure Containers

Prompt requires its databases and a Keycloak instance to run. Start them with:

```bash
make db
```

Each service uses a **separate** PostgreSQL database, and `make db` starts all of them together with Keycloak on port `8081` (Keycloak brings up its own database container as a dependency). Stop them again with `make db-down`.

The host ports are `5432` (core), `5434` (team allocation), `5435` (assessment), `5436` (self team allocation), `5437` (example), `5438` (interview), `5439` (certificate), `5440` (presentation), and `5441` (infrastructure setup) — the same values `.env.dev` already points at.

File uploads (application documents, certificates, presentation materials) are stored through the SeaweedFS S3 gateway. It is **required** — start it before the servers in step 5:

```bash
docker compose up -d seaweedfs-volume seaweedfs-s3
```

Without it the core and presentation servers exit during startup instead of coming up degraded: both check that their bucket exists while building the S3 adapter, and abort on the connection error with `Failed to initialize prompt file storage` (core) or `Unable to initialize presentation material storage` (presentation).

### 4. Configure Keycloak (only on initial setup)

The `prompt` realm is **imported automatically** from `keycloakConfig.json` when the container starts, so you do not create it by hand. Once Keycloak is running:

1. Navigate to [http://localhost:8081](http://localhost:8081) and log in to the **Keycloak Administrative Console** with:
   - **Username**: `admin`
   - **Password**: `admin`
2. The realm already contains seeded test users for every role (`admin`, `lecturer`, `course-lecturer`, `course-editor`, `student`), each with the `university_login` and `matriculation_number` attributes set. The [Keycloak Local Development Guide](./keycloak-dev.md) lists the credentials and explains how to add users or roles yourself.
3. **Generate a Client Secret**:
   - Go to **Clients** > **prompt-server** > **Credentials**.
   - Click **Regenerate** to get a new secret and copy it.
4. Paste the secret into `KEYCLOAK_CLIENT_SECRET`:
   - in **`.env.dev`** for servers you start with `make server` / `make servers`,
   - in **`.env`** for servers you run through `docker compose`.

   `.env.template` ships the line as `KEYCLOAK_CLIENT_SECRET= # FIXME: Set your Keycloak client secret here`. Delete the trailing comment instead of pasting in front of it: the Makefile `include`s `.env`, and the space before the `#` stays in the value, so Keycloak rejects the secret with `invalid_client`.

:::warning
`.env.dev` is loaded *after* `.env`, and an **empty** assignment still overrides. If you set the secret only in `.env` while `.env.dev` keeps `KEYCLOAK_CLIENT_SECRET=`, the Makefile-started servers come up without a secret and log `Failed to initialize keycloak`; everything that talks to Keycloak (course creation, role management) then fails.
:::

:::note
`keycloakConfig.json` ships with a placeholder secret, so this step is needed for every fresh Keycloak database. To start over, run `docker compose down -v` and `rm -rf keycloak_postgres_data`; the realm is re-imported on the next start.
:::

### 5. Start the Servers

Start the core server (port `8080`):

```bash
make server
```

The target loads `.env` and `.env.dev`, downloads Go dependencies on demand, and applies pending migrations on startup. Watch the log output: failures to reach PostgreSQL or Keycloak show up right there.

Phase services have their own targets — `make server-assessment`, `make server-interview`, `make server-team-allocation`, `make server-self-team-allocation`, `make server-example`, `make server-certificate`, `make server-presentation`, `make server-infrastructure-setup` — and `make servers` starts the core server together with all of them.

### 6. Start the Clients

In a separate terminal, launch **all** microfrontends at once:

```bash
make clients
```

This installs the workspace dependencies (`yarn install`) and starts every microfrontend in parallel on its own rspack dev server. Then open [http://localhost:3000](http://localhost:3000).

| Microfrontend | Port |
| --- | --- |
| `core` | 3000 |
| `example_component` | 3001 |
| `interview_component` | 3002 |
| `matching_component` | 3003 |
| `assessment_component` | 3007 |
| `team_allocation_component` | 3008 |
| `self_team_allocation_component` | 3009 |
| `certificate_component` | 3010 |
| `presentation_component` | 3011 |
| `infrastructure_setup_component` | 3012 |

To run only a subset, use one of the per-client targets — `make client-core`, `make client-assessment`, `make client-certificate`, `make client-interview`, `make client-matching`, `make client-presentation`. The remaining components (example, team allocation, self team allocation, infrastructure setup) have no target; run `yarn dev` inside the corresponding folder:

```bash
cd clients/core && yarn dev
```

Phases developed in their own repositories (for example the intro course and the GitHub challenge) are loaded as external remotes and served by those projects.

### 7. Load the Demo Data (optional)

The databases start empty. Once the servers have run once and created their schemas, load a fully populated demo course into every service database with:

```bash
make seed
```

The command can be re-run and resets the rows it owns. See [Database Seeding](./guide/seeding.md) for what it creates.

### 8. Verify Your Setup

```bash
make lint
make test
```

`make test` runs the client unit tests and every Go test suite (Go tests start throwaway databases via `testcontainers-go`, so Docker must be running). The end-to-end Playwright suite is separate and documented in `e2e/README.md`.

---

You should now have Keycloak (on `localhost:8081`), your PostgreSQL databases, the Go backend, and the React microfrontends running. Happy coding!

## Optional: IDE Configuration

- You can use any IDE or text editor for **Go** and **React** development. Popular choices include:
  - [Visual Studio Code](https://code.visualstudio.com/)
  - [GoLand](https://www.jetbrains.com/go/) for deeper Go integration

## Summary

With Docker, Go, Node.js, and Yarn installed, you can:

1. Compile and run the Golang backend.
2. Build and run the React frontend.
3. Develop new features or set up a demo environment.

---

**Happy coding with Prompt!**
