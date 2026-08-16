---
title: "Installation"
sidebar_position: 3
---

# Installation

Two reference configurations are described below: a self-contained quick start and a
production install. Both use the same chart; they differ only in values.

Always install with a values file. Several settings have no usable default and the chart fails the
render with an explicit message rather than installing something broken. See
[Required values](./configuration.md#required-values).

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

**The release is not usable yet.** Because Keycloak runs in-cluster here, you still have to import
the platform realm and feed the resulting client identity back into the release. Until then every
login fails with `invalid_client`. Follow
[Keycloak → Realm import](./keycloak.md#realm-import), which also covers replacing the placeholder
client secret shipped in `keycloakConfig.json`.

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
