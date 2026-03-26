---
sidebar_position: 5
---

# 📁 File Storage (SeaweedFS)

PROMPT uses [SeaweedFS](https://github.com/seaweedfs/seaweedfs) as an S3-compatible object store for user-uploaded files (e.g., application attachments). The core server communicates with SeaweedFS via the standard AWS S3 SDK and exposes presigned URLs to clients.

---

## Architecture Overview

SeaweedFS is deployed as four containers:

| Container          | Role                                                           | Internal Port |
| ------------------ | -------------------------------------------------------------- | ------------- |
| `seaweedfs-master` | Cluster coordination and volume assignment                     | 9333          |
| `seaweedfs-volume` | Actual blob storage on disk                                    | 8080          |
| `seaweedfs-filer`  | File-system abstraction over volumes (required for S3 gateway) | 8888          |
| `seaweedfs-s3`     | S3-compatible API gateway                                      | 8333          |

All four services run on the shared Docker network. Only the **S3 gateway** is exposed externally (via Traefik in production, via port mapping in development).

---

## Network-Level Access

### Production

```
Browser ─── HTTPS ──▶ Traefik (s3.<CORE_HOST>:443) ──▶ seaweedfs-s3:8333
Core Server ── HTTP ──▶ seaweedfs-s3:8333  (Docker-internal, prompt-network)
```

- **Master, Volume, Filer** are internal-only (`traefik.enable=false`, only `expose` — no host port bindings).
- **S3 gateway** is the single entry point. Traefik terminates TLS and routes `Host(s3.<CORE_HOST>)` to port 8333.
- The core server reaches the S3 gateway via the internal Docker network endpoint `http://seaweedfs-s3:8333` (`S3_ENDPOINT`).
- Browsers reach the S3 gateway via the public endpoint `https://s3.<CORE_HOST>` (`S3_PUBLIC_ENDPOINT`), which is used to generate presigned URLs.

### Local Development

```
Browser ─── HTTP ──▶ localhost:8334 ──▶ seaweedfs-s3:8333
Core Server ── HTTP ──▶ seaweedfs-s3:8333  (Docker-internal, default compose network)
```

Port 8334 on the host is mapped to the S3 gateway's port 8333.

---

## Access Control

SeaweedFS itself performs **credential-based access control** via an S3 config file generated when the `seaweedfs-s3` container starts. The file is rendered from the configured S3 credentials and passed to `weed s3` via `-config`, so only that generated identity can interact with the S3 API.

### Presigned URL Flow

All client-facing file access uses **presigned URLs**, which embed a time-limited cryptographic signature. The flow is:

1. **Upload**: The client requests a presigned PUT URL from the core server (`POST /apply/:coursePhaseID/files/presign`). The server generates a presigned URL using the S3 credentials and returns it. The client uploads the file directly to SeaweedFS using this URL, then calls `/files/complete` so the server verifies the file exists and records its metadata in the database.

2. **Download**: When a file needs to be accessed, the core server generates a presigned GET URL (default TTL: 30 seconds, configurable via `S3_PRESIGN_DOWNLOAD_TTL_SECONDS`). The browser fetches the file directly from SeaweedFS using this URL.

3. **Expiry**: Presigned URLs expire after their TTL. Without a valid signature, SeaweedFS rejects the request. This means **no file is accessible without the core server explicitly granting a time-limited URL**.

### Application-Level Authorization (Core Server)

The core server enforces additional authorization before generating presigned URLs:

- **Upload** is only allowed when the target course phase is an **open application phase** (`ensureOpenApplicationPhase()` check).
- **Delete** requires authentication, ownership verification (file must have been uploaded by the requesting user), and course-phase membership.
- **Authenticated routes** (`/apply/authenticated/...`) require a valid Keycloak bearer token.
- **External routes** (`/apply/...`) are unauthenticated but scoped to open application phases only — these are used for the public application form where applicants are not yet registered.

### Storage Key Isolation

Files are stored with keys scoped to their course phase: `course-phase/<coursePhaseID>/<uuid>-<filename>`. The `CreateFileFromStorageKey` service method validates that the storage key prefix matches the target course phase, preventing cross-phase file references.

---

## GitHub Environment Configuration

To deploy SeaweedFS via the CI/CD pipeline, the following variables and secrets must be set in the GitHub repository settings for each environment (`prompt-dev-vm`, `prompt-prod-vm`):

### Variables (`vars.*`)

| Variable                  | Example Value                          | Description                                                                 |
| ------------------------- | -------------------------------------- | --------------------------------------------------------------------------- |
| `S3_BUCKET`               | `prompt-files`                         | Bucket name (auto-created at startup if missing)                            |
| `S3_REGION`               | `us-east-1`                            | Region (use `us-east-1` for SeaweedFS)                                      |
| `S3_ENDPOINT`             | `http://seaweedfs-s3:8333`             | Internal Docker endpoint for the S3 gateway                                 |
| `S3_PUBLIC_ENDPOINT`      | `https://s3.prompt.ase.cit.tum.de`     | Public endpoint for presigned URLs (must resolve to S3 gateway via Traefik) |
| `S3_FORCE_PATH_STYLE`     | `true`                                 | Must be `true` for SeaweedFS/MinIO                                          |
| `S3_PRESIGN_UPLOAD_TTL_SECONDS`   | `60`                                   | Presigned upload URL TTL in seconds                                          |
| `S3_PRESIGN_DOWNLOAD_TTL_SECONDS` | `30`                                   | Presigned download URL TTL in seconds                                        |
| `S3_PRESIGN_TTL_SECONDS`          | (optional legacy)                      | Legacy fallback TTL used if the specific upload/download TTLs are not set    |
| `MAX_FILE_UPLOAD_SIZE_MB` | `50`                                   | Maximum upload size in MB                                                   |
| `ALLOWED_FILE_TYPES`      | `application/pdf,image/jpeg,image/png` | Comma-separated MIME types (empty = allow all)                              |

### Secrets (`secrets.*`)

| Secret          | Description                                                                                |
| --------------- | ------------------------------------------------------------------------------------------ |
| `S3_ACCESS_KEY` | S3 API access key (used by both core server and SeaweedFS gateway, e.g., `prompt-s3-user`) |
| `S3_SECRET_KEY` | S3 API secret key (used by both core server and SeaweedFS gateway) — **must be strong**    |

:::tip
These credentials are used by **both** the core server's S3 client and the SeaweedFS S3 gateway container. The gateway authenticates incoming requests with these credentials, and the core server uses them to sign requests.
:::

### DNS Requirement

In production, `s3.<CORE_HOST>` must resolve to the same VM where Traefik is running, so that:

- Traefik can obtain a TLS certificate via Let's Encrypt
- Presigned URLs generated with `S3_PUBLIC_ENDPOINT` are reachable by clients

---

## Production Docker Compose Notes

The `docker-compose.prod.yml` is **correctly configured** for SeaweedFS deployment:

- All four SeaweedFS containers are defined with health checks and restart policies.
- Master, Volume, and Filer are internal-only (no Traefik routing, only `expose`).
- The S3 gateway has Traefik labels for `Host(s3.${CORE_HOST})` with TLS.
- The core server receives all required S3 environment variables.
- SeaweedFS data directories are bind-mounted for persistence across container restarts.
- The S3 gateway generates its `s3.json` config on startup from `S3_ACCESS_KEY` / `S3_SECRET_KEY` and passes it to `weed s3` via `-config`.

:::note
The SeaweedFS containers are **not restarted** during application deployments — the deployment workflow's stop step only targets application containers, so file storage persists across deploys.
:::
