---
title: "Operations"
sidebar_position: 9
---

# Operations

## Upgrades & rollbacks

Bump `global.image.tag` (or override per component) and upgrade:

```bash
helm upgrade prompt charts/prompt -n prompt -f values-prod.yaml
```

- Database passwords are preserved across upgrades (Helm `lookup`).
- Each service re-runs its startup migrations safely.
- `helm rollback` reverts Deployments and config, but **not** data: a schema migration applied by
  a newer release is not undone by rolling back the release. Restore the database from a backup if
  you need to revert schema changes.
- Upgrade operators (CloudNativePG, Keycloak, Envoy Gateway, cert-manager) following their own
  release notes; they are independent of the chart.

## Scaling & HA

```yaml
global:
  postgresql:
    instances: 3            # 1 primary + 2 standbys
    pooler:
      instances: 2
core:
  server:
    replicas: 3
```

CloudNativePG manages Postgres failover and a PodDisruptionBudget. Increase per-component
`replicas` for stateless frontends and backends.

## Enabling/disabling phases

Toggle a phase with its `enabled` flag:

```yaml
matching:
  enabled: false
```

Disabling removes the phase's Deployments, Services, HTTPRoutes, and `Database` resource. Because
the database uses `databaseReclaimPolicy: retain` and PVCs persist, **no data is deleted**. The
per-phase role and credentials Secret remain so the data stays accessible if you re-enable later.

## Observability

- **Sentry:** set `global.appConfig.sentryEnabled: "true"` and provide the per-service DSNs
  (`SENTRY_DSN_*`) in the application secret.
- **Logs:** `kubectl -n prompt logs deploy/prompt-core-server` (and other components).
- **Health:** backends expose TCP readiness/liveness on their service port; frontends expose
  `GET /`. CloudNativePG and the Gateway report status on their own resources.

## Secrets management

The chart generates and persists **database** credentials. The shared **application** secret
(SMTP, Sentry DSNs, Keycloak client secret, S3 keys) is either rendered from
`global.appSecrets.data` or taken from an existing Secret:

```yaml
global:
  appSecrets:
    existingSecret: prompt-app-secrets
```

When `existingSecret` is set the chart renders no application Secret, and every component reads
the named Secret instead. This is the integration point for External Secrets Operator or
sealed-secrets - have them produce a Secret with the expected keys (`KEYCLOAK_CLIENT_SECRET`,
`SMTP_*`, `SENDER_*`, `S3_ACCESS_KEY`, `S3_SECRET_KEY`, `SENTRY_DSN_*`).

## Uninstall / teardown

```bash
helm uninstall prompt -n prompt
```

This removes the chart's workloads and routes. The following **persist by design** and must be
removed manually if you want a full purge:

- PersistentVolumeClaims (Postgres, SeaweedFS)
- CloudNativePG `Database` resources (retain policy)
- Generated credential Secrets

```bash
kubectl -n prompt delete pvc --all
kubectl -n prompt delete secret -l app.kubernetes.io/part-of=prompt
```
