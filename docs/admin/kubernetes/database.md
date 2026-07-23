---
title: "Database Operations"
sidebar_position: 6
---

# Database Operations

## Model

In-cluster mode runs one CloudNativePG `Cluster` hosting one logical database per phase
(`prompt_core`, `prompt_assessment`, ...). Each database has its own owner role, declared on the
cluster (`spec.managed.roles`) with a persisted password. A PgBouncer `Pooler` fronts read-write
traffic so many service replicas do not exhaust `max_connections`.

Each phase's `Database` resource uses `databaseReclaimPolicy: retain`, so deleting it (or
disabling the phase) never drops data.

## Credentials

The chart generates each role's password and stores it in a Secret. On `helm upgrade` it reuses
the existing password (Helm `lookup`), so upgrades never rotate credentials and break
authentication. The application reads its database connection from a per-phase Secret
(`<release>-db-<phase>`) via `envFrom`.

## Migrations

Each Go service runs `golang-migrate` on startup against its own database. golang-migrate takes a
database advisory lock, so concurrent replica startups serialize safely. A generous
`startupProbe` prevents the liveness probe from killing a pod mid-migration.

If a pod is killed mid-migration the schema can be left "dirty". Recover by forcing the version:

```bash
# inspect
kubectl -n prompt exec deploy/prompt-core-server -- migrate -path ... -database "$DATABASE_URL" version
# force to the last good version, then let the pod re-run
migrate -path ... -database "$DATABASE_URL" force <version>
```

## Backups (WAL + scheduled base backups)

Enable continuous WAL archiving and scheduled base backups to object storage:

```yaml
infrastructure:
  backup:
    enabled: true
    schedule: "0 0 2 * * *"
    retention: 30d
    destinationPath: s3://prompt-backups/pg
```

The backup uses the S3 credentials from the shared application secret. **Use off-cluster object
storage** for backups: writing them to the in-cluster SeaweedFS is a correlated failure (if the
cluster dies, so do the backups).

### Point-in-time recovery (PITR)

Recover into a new cluster from the object store, targeting a timestamp:

```yaml
# new cluster manifest (illustrative)
spec:
  bootstrap:
    recovery:
      source: prompt-pg
      recoveryTarget:
        targetTime: "2026-06-22 14:30:00+00"
  externalClusters:
    - name: prompt-pg
      barmanObjectStore:
        destinationPath: s3://prompt-backups/pg
        s3Credentials: { ... }
```

Always recover into a new cluster name / new path to avoid overwriting the live WAL archive.

## Minor-version upgrades

CloudNativePG performs rolling minor-version upgrades when you bump the cluster image. Application
startup migrations are independent of this and re-run safely.

## External mode

With `global.postgresql.mode: external`, no CNPG resources are created. The chart still creates
the per-phase `<release>-db-<phase>` Secrets pointing at `global.postgresql.external.*`, so the
services remain mode-agnostic. Provision the databases and roles on the external server yourself.
