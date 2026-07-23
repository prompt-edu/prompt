---
title: "Configuration Reference"
sidebar_position: 4
---

# Configuration Reference

All keys live under `global` (shared with every subchart) or `infrastructure` (consumed only by
the infrastructure subchart). Per-phase blocks (`core`, `assessment`, ...) toggle and tune each
phase.

## Global

| Key | Default | Description |
| --- | --- | --- |
| `global.host` | `prompt.example.com` | Bare apex hostname (no scheme). |
| `global.scheme` | `https` | URL scheme used to build public URLs. |
| `global.image.registry` | `ghcr.io/prompt-edu/prompt` | Image registry/prefix. |
| `global.image.tag` | `latest` | Chart-wide image tag (the release version). |
| `global.image.pullSecrets` | `[]` | Image pull secrets. |
| `global.environment` | `production` | Surfaced to the frontend. |
| `global.chairNameShort` / `chairNameLong` | `AET` / ... | Branding shown in the UI. |

### PostgreSQL (`global.postgresql`)

| Key | Default | Description |
| --- | --- | --- |
| `mode` | `in-cluster` | `in-cluster` (CloudNativePG) or `external`. |
| `clusterName` | `prompt-pg` | CNPG cluster name. |
| `instances` | `3` | Replicas (1 primary + N-1 standbys). |
| `storageSize` | `10Gi` | PVC size per instance. |
| `storageClass` | `""` | Use an encrypted class in production. |
| `maxConnections` | `200` | Postgres `max_connections`. |
| `pooler.enabled` | `true` | Deploy a PgBouncer pooler. |
| `external.host` / `port` / `user` / `password` | `""` | External DB connection (external mode). |
| `databases.<phase>` | see values | Per-phase `dbName`, `owner`, `hostVar`, `portVar`. |

### Keycloak (`global.keycloak`)

| Key | Default | Description |
| --- | --- | --- |
| `mode` | `external` | `external` or `in-cluster`. |
| `realm` | `prompt` | Realm name. |
| `clientId` / `idOfClient` / `authorizedParty` | ... | OIDC client identifiers. |
| `externalHost` | `https://auth.example.com` | Existing Keycloak URL (external mode). |

### Object storage (`global.objectStorage`)

| Key | Default | Description |
| --- | --- | --- |
| `mode` | `in-cluster` | `in-cluster` (SeaweedFS) or `external`. |
| `bucket` / `region` | `prompt-files` / `us-east-1` | Bucket and region. |
| `forcePathStyle` | `"true"` | Required for SeaweedFS / MinIO. |
| `external.endpoint` / `publicEndpoint` | `""` | Internal and browser-facing S3 URLs. |

### Gateway and TLS (`global.gateway`)

| Key | Default | Description |
| --- | --- | --- |
| `gatewayClassName` | `eg` | GatewayClass (Envoy Gateway). |
| `certManager.enabled` | `true` | Issue per-listener TLS via cert-manager. |
| `certManager.issuerName` | `prompt-letsencrypt` | ClusterIssuer name. |
| `certManager.create` | `true` | Create the ClusterIssuer (set false to reuse one). |
| `certManager.email` | `admin@example.com` | ACME account email. |
| `rateLimiting.provider` | `envoygateway` | `envoygateway` or `none`. |
| `rateLimiting.average` / `burst` | `300` / `100` | Request rate limit on `/api` routes. |

## Per-phase blocks

Each phase (`core`, `assessment`, `interview`, `team-allocation`, `self-team-allocation`,
`certificate`, `matching`, `template`) has an `enabled` flag and `server` / `client` sub-blocks
(`image`, `replicas`, `port`, `apiPath`/`path`, `dbKey`, `rateLimited`, `stripPrefix`). Disable a
phase entirely:

```yaml
interview:
  enabled: false
```

This removes the phase's workloads, routes, and database. The logical database's credentials are
retained (see [Operations](./operations.md)).

## Image versioning

`global.image.tag` pins every component to one release. Override a single component with its
`tag` key, for example `assessment.server.tag: v1.2.3`.

## Secrets

See [Secrets management](./operations.md#secrets-management) for chart-generated vs.
`existingSecret`. Database credentials are always generated and persisted by the chart; the
shared application secret (SMTP, Sentry DSNs, Keycloak client secret, S3 keys) is provided via
`global.appSecrets.data` or `global.appSecrets.existingSecret`.
