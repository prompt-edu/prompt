---
title: "Installation"
sidebar_position: 3
---

# Installation

Two reference configurations are described below: a self-contained quick start and a
production install. Both use the same chart; they differ only in values.

## Quick start (self-contained)

Everything runs in-cluster (CloudNativePG, SeaweedFS, Keycloak). Use this for evaluation or a
demo, not for production (database backups and object storage share the cluster's fate).

`values-quickstart.yaml`:

```yaml
global:
  host: prompt.example.com
  postgresql:
    mode: in-cluster
    instances: 1
    pooler:
      enabled: false
  objectStorage:
    mode: in-cluster
  keycloak:
    mode: in-cluster
  appSecrets:
    data:
      S3_SECRET_KEY: "change-me"
  gateway:
    certManager:
      email: admin@example.com
```

```bash
helm install prompt charts/prompt -n prompt --create-namespace -f values-quickstart.yaml
```

Then follow the printed notes: point DNS at the Gateway, wait for certificates, and watch the
pods come up.

## Production install (recommended)

In-cluster CloudNativePG, but **external** object storage (so database backups live off-cluster)
and an **external** Keycloak (most institutions already run one).

`values-prod.yaml`:

```yaml
global:
  host: prompt.aet.tum.de
  postgresql:
    mode: in-cluster
    instances: 3
    storageClass: encrypted-ssd
  objectStorage:
    mode: external
    bucket: prompt-files
    external:
      endpoint: https://s3.eu-central-1.amazonaws.com
      publicEndpoint: https://s3.eu-central-1.amazonaws.com
  keycloak:
    mode: external
    externalHost: https://auth.tum.de
    realm: prompt
    clientId: prompt-server
    idOfClient: "<uuid>"
  appSecrets:
    existingSecret: prompt-app-secrets   # provision out of band (see Configuration)
infrastructure:
  backup:
    enabled: true
    destinationPath: s3://prompt-backups/pg
    retention: 30d
```

```bash
helm install prompt charts/prompt -n prompt --create-namespace -f values-prod.yaml
```

## Verifying the install

```bash
kubectl -n prompt get pods
kubectl -n prompt get gateway,httproute
kubectl -n prompt get certificate            # all should reach READY=True
kubectl -n prompt get cluster                # CloudNativePG cluster health
```

See [Troubleshooting](./troubleshooting.md) if certificates or pods do not become ready.
