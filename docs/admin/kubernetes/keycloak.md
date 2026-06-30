---
title: "Keycloak"
sidebar_position: 8
---

# Keycloak

PROMPT authenticates against Keycloak. The realm, clients, and roles are the same regardless of
how Keycloak is hosted - see [Keycloak (production)](/admin/keycloak-prod) for the realm/client/role
model. This page covers only the Kubernetes hosting choice.

## External (default)

Point the platform at an existing Keycloak:

```yaml
global:
  keycloak:
    mode: external
    externalHost: https://auth.tum.de
    realm: prompt
    clientId: prompt-server
    idOfClient: "<uuid-of-prompt-server-client>"
    authorizedParty: prompt-client
```

Supply the client secret via the application secret (`KEYCLOAK_CLIENT_SECRET`). No Keycloak
workloads are created in the cluster.

## In-cluster (Keycloak Operator)

```yaml
global:
  keycloak:
    mode: in-cluster
    realm: prompt
```

This requires the [Keycloak Operator](./prerequisites.md#keycloak-operator-only-when-keycloakmode-in-cluster)
and creates:

- a `keycloak` database on the shared CloudNativePG cluster,
- a `Keycloak` custom resource (1 instance, TLS terminated at the gateway),
- an `HTTPRoute` on the `auth.<host>` listener.

### Realm import

To import the platform realm, place `keycloakConfig.json` in a ConfigMap and reference it:

```bash
kubectl -n prompt create configmap prompt-realm --from-file=realm.json=keycloakConfig.json
```

```yaml
infrastructure:
  realmConfigMap: prompt-realm
```

The operator applies the realm via a `KeycloakRealmImport`. After import, set
`global.keycloak.idOfClient` to the `prompt-server` client UUID and provide its secret.
