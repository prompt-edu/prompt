---
title: "Configuration Reference"
sidebar_position: 4
---

# Configuration Reference

All keys live under `global` (shared with every subchart) or `infrastructure` (consumed only by
the infrastructure subchart). Per-phase blocks (`core`, `assessment`, ...) toggle and tune each
phase.

## Required values

The chart has no working set of defaults, so always install with a values file. These have no
usable default and the render fails with an explicit message if they are missing:

| Key | Required when |
| --- | --- |
| `global.host` | always (the `prompt.example.com` default resolves nowhere) |
| `global.appSecrets.data.S3_SECRET_KEY` | always, unless `global.appSecrets.existingSecret` is set |
| `global.appSecrets.data.S3_ACCESS_KEY` | `objectStorage.mode: external` (the default is the bundled SeaweedFS identity) |
| `global.postgresql.external.host` / `.user` / `.password` | `postgresql.mode: external` |
| `global.objectStorage.external.endpoint` / `.publicEndpoint` | `objectStorage.mode: external` |

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
| `global.podSecurity.runAsUser` / `runAsGroup` | `65532` | UID/GID the Go backends run as. The images build on distroless static (root variant), so the chart supplies the ID. |
| `global.dbWaitImage` | `postgres:17-alpine` | Image for the DB-wait init container. Must contain `psql`. |
| `global.externalRemotes.exampleServerHost` | `""` | Example server API base. Empty means the apex host. |
| `global.externalRemotes.introCourseHost` | `""` | Intro-course deployment URL. Not served by this chart. |
| `global.externalRemotes.devopsChallengeHost` | `""` | DevOps-challenge deployment URL. Not served by this chart. |

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
| `sslMode` | `""` | libpq `sslmode`. Empty resolves to `disable` in-cluster and `require` external. `verify-ca`/`verify-full` are unsupported: the chart configures no CA. |
| `external.host` / `port` / `user` / `password` | `""` | External DB connection. Host, user and password are required in external mode; one account is used for every logical database, matching the Compose setup. |
| `databases.<phase>` | see values | Per-phase `dbName`, `owner`, `hostVar`, `portVar`. `owner` applies to in-cluster mode only. |

### Keycloak (`global.keycloak`)

| Key | Default | Description |
| --- | --- | --- |
| `mode` | `external` | `external` or `in-cluster`. |
| `realm` | `prompt` | Realm name. |
| `clientId` / `idOfClient` / `authorizedParty` | ... | OIDC client identifiers. |
| `externalHost` | `https://auth.example.com` | Existing Keycloak URL (external mode). |

`keycloak.mode: in-cluster` requires `postgresql.mode: in-cluster`. The bundled Keycloak uses the
CloudNativePG cluster and its generated role secret, neither of which exists in external mode, so
the chart rejects that combination at render time rather than failing after apply. To run Keycloak
in the cluster against a managed database, provision the Keycloak database yourself and point an
externally managed Keycloak at it.

Importing the realm is a separate, manual step. See [Keycloak](./keycloak.md).

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
| `provider` | `envoygateway` | `envoygateway` or `none`. Gates every Envoy-Gateway-specific resource: the request rate limit **and** the S3 CORS policy. |
| `rateLimiting.enabled` | `true` | Request rate limit on `/api` routes. Independent of CORS. |
| `rateLimiting.average` | `300` | Requests per second, per Envoy instance. |

Setting `provider: none` also removes the CORS policy on the S3 route, which browsers need for
presigned uploads and downloads. On a non-Envoy controller, reproduce that CORS configuration with
your controller's own mechanism.

### Infrastructure (`infrastructure`)

| Key | Default | Description |
| --- | --- | --- |
| `networkPolicy.enabled` | `true` | Blocks cross-namespace access to the unauthenticated SeaweedFS master, volume and filer. Needs a CNI that enforces NetworkPolicy. |
| `backup.enabled` | `false` | CNPG WAL + scheduled backup. Uses deprecated in-tree Barman support, see [Database](./database.md). |
| `seaweedfs.*` | see values | Image and per-component PVC sizes. |

## Components this chart does not deploy

`clients/core` compiles its module-federation remote URLs at build time as same-origin paths, so
`/intro-course-developer` and `/github-challenge` must be served under the apex host. They come
from separate repositories (`prompt-intro-course`, `prompt-github-challenge`), and the Compose
deployment does not serve them either. If you run them, add your own route:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: intro-course-developer
spec:
  parentRefs:
    - name: prompt-gateway
      sectionName: apex
  hostnames: ["prompt.example.com"]
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /intro-course-developer
      filters:
        - type: URLRewrite
          urlRewrite:
            path:
              type: ReplacePrefixMatch
              replacePrefixMatch: /
      backendRefs:
        - name: intro-course-client   # a Service in this namespace
          port: 80
```

For an off-cluster deployment use Envoy Gateway's `Backend` resource with an FQDN endpoint. Do not
reach for an `ExternalName` Service: Gateway API treats those backends as implementation-specific
and says implementations should not support them. Another controller needs its own external-FQDN
mechanism.

Setting `global.externalRemotes.introCourseHost` / `devopsChallengeHost` only fills the matching
`window.env` API base URLs; it cannot move the remote entry URLs.

## Per-phase blocks

Each phase (`core`, `assessment`, `interview`, `team-allocation`, `self-team-allocation`,
`certificate`, `matching`, `example`) has an `enabled` flag and `server` / `client` sub-blocks
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
