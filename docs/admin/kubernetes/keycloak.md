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
and `postgresql.mode: in-cluster` (the bundled Keycloak uses the CloudNativePG cluster and its
generated role secret; the chart rejects the external combination at render time). It creates:

- a `keycloak` database on the shared CloudNativePG cluster,
- a `Keycloak` custom resource (1 instance, TLS terminated at the gateway),
- an `HTTPRoute` on the `auth.<host>` listener.

### Realm import

The chart does **not** import a realm. `KeycloakRealmImport.spec.realm` takes a full inline
`RealmRepresentation`, which cannot be assembled from chart values, so this is a one-off manual
step. Until it is done nobody can log in: the realm exists with no `prompt-server` /
`prompt-client` clients and every login fails with `invalid_client`.

1. **Copy the realm file.** Start from `keycloakConfig.json` in the repository root. It is written
   for local development, so replace the redirect URIs and web origins of the `prompt-server` and
   `prompt-client` clients with your deployment host.

2. **Generate a real client secret.** The shipped file has the literal placeholder `"**********"`
   as the `prompt-server` secret. Applying it unchanged makes that predictable string your client
   secret. Generate one and set it in your copy:

   ```bash
   openssl rand -hex 32
   ```

3. **Apply it as a `KeycloakRealmImport`.** The chart prints the Keycloak CR name after install
   (`<release>-keycloak`):

   ```bash
   yq -n '.apiVersion="k8s.keycloak.org/v2alpha1" | .kind="KeycloakRealmImport"
     | .metadata.name="prompt-realm" | .spec.keycloakCRName="prompt-keycloak"
     | .spec.realm = load("keycloakConfig.json")' | kubectl -n prompt apply -f -
   ```

   Watch it complete with `kubectl -n prompt get keycloakrealmimport`.

4. **Feed the client identity back into the release.** Read the `prompt-server` client UUID from
   the Keycloak admin console, then upgrade:

   ```yaml
   global:
     keycloak:
       idOfClient: "<uuid-of-prompt-server-client>"
     appSecrets:
       data:
         KEYCLOAK_CLIENT_SECRET: "<the secret generated in step 2>"
   ```

   The backends read both at startup, so `helm upgrade` restarts them automatically.
