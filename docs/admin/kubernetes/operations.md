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

- Database passwords are preserved across upgrades: the chart generates each in-cluster role
  password once and reads it back with Helm's `lookup`. This needs `get` on secrets in the release
  namespace for whoever runs `helm upgrade`. Without that permission Helm fails the render rather
  than silently regenerating. External installs supply their own credentials and use no `lookup`.
- Config and secret changes roll the pods. Every workload carries `checksum/appconfig` and
  `checksum/appsecrets` annotations (backends also `checksum/db`), so editing
  `global.appConfig`, `global.appSecrets` or the external database connection triggers a restart on
  `helm upgrade` instead of leaving the old values live in running containers. Two cases are not
  covered and need `kubectl rollout restart` or a reloader: `appSecrets.existingSecret` (the chart
  cannot see the content) and an in-cluster role password edited out of band.
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
the database uses `databaseReclaimPolicy: retain` and the CNPG `Cluster` is retained on uninstall,
**no data is deleted**. The per-phase role and credentials Secret remain so the data stays
accessible if you re-enable later.

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
`SMTP_*`, `SENDER_*`, `S3_ACCESS_KEY`, `S3_SECRET_KEY`, `SENTRY_DSN_*`). The chart cannot hash a
Secret it does not render, so rotating a value there does not restart pods on its own.

## Uninstall / teardown

```bash
helm uninstall prompt -n prompt
```

This removes the chart's workloads and routes. Data-bearing resources carry
`helm.sh/resource-policy: keep`, so uninstall leaves them behind:

- the CloudNativePG `Cluster` (it owns the Postgres PVCs through `ownerReferences`, so deleting it
  deletes the databases)
- the generated per-role credential Secrets
- SeaweedFS PVCs, which come from `volumeClaimTemplates` and are not part of the release at all

Kept resources are **orphaned** from the release. A later `helm install` with the same name in the
same namespace may be accepted, because those objects still carry their
`meta.helm.sh/release-name` and `meta.helm.sh/release-namespace` annotations, but that is a
necessary condition and not a guarantee: if the annotations were changed or the objects edited by
hand, the install fails with an ownership conflict. Do not use `helm install --replace` to work
around it. Either resolve the conflict on the specific object, or purge and start clean.

For a genuinely clean teardown, including the data:

```bash
helm uninstall prompt -n prompt
kubectl -n prompt delete cluster.postgresql.cnpg.io prompt-pg
kubectl -n prompt delete pvc --all
kubectl -n prompt delete secret -l app.kubernetes.io/part-of=prompt
```
