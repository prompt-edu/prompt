---
title: "Prerequisites"
sidebar_position: 2
---

# Prerequisites

The chart provisions PROMPT's own workloads but relies on cluster-level operators being present.
Operators install CRDs and are cluster-scoped, so they are documented prerequisites rather than
bundled per release.

## Cluster and tooling

- Kubernetes 1.29+
- Helm 3.14+ (or 4.x)
- A default (ideally encryption-at-rest) `StorageClass`
- DNS control over the apex host and any enabled subdomains

## Required operators

### Envoy Gateway + Gateway API CRDs (always)

```bash
helm install eg oci://docker.io/envoyproxy/gateway-helm \
  --version v1.3.0 -n envoy-gateway-system --create-namespace
```

This installs the Gateway API CRDs and an `eg` GatewayClass. To use a different controller, set
`global.gateway.gatewayClassName`; note that rate limiting is Envoy-specific (see
[Configuration](./configuration.md)).

### cert-manager with the Gateway API feature gate (for TLS)

```bash
helm install cert-manager jetstack/cert-manager \
  -n cert-manager --create-namespace \
  --set crds.enabled=true \
  --set "extraArgs={--feature-gates=ExperimentalGatewayAPISupport=true}"
```

The Gateway API feature gate is required for cert-manager to issue certificates from Gateway
listeners. Without it, no TLS certificates are created.

### CloudNativePG operator (when `postgresql.mode: in-cluster`)

```bash
kubectl apply --server-side -f \
  https://raw.githubusercontent.com/cloudnative-pg/cloudnative-pg/release-1.27/releases/cnpg-1.27.0.yaml
```

### Keycloak Operator (only when `keycloak.mode: in-cluster`)

```bash
kubectl apply -f \
  https://raw.githubusercontent.com/keycloak/keycloak-k8s-resources/26.0.0/kubernetes/keycloaks.k8s.keycloak.org-v1.yml
kubectl apply -f \
  https://raw.githubusercontent.com/keycloak/keycloak-k8s-resources/26.0.0/kubernetes/keycloakrealmimports.k8s.keycloak.org-v1.yml
kubectl apply -f \
  https://raw.githubusercontent.com/keycloak/keycloak-k8s-resources/26.0.0/kubernetes/kubernetes.yml
```

## DNS

Create A/AAAA records pointing at the Gateway's external address for every enabled host:

- `<host>` (always)
- `s3.<host>` (when object storage is in-cluster)
- `auth.<host>` (when Keycloak is in-cluster)

HTTP-01 validation requires port 80 to be reachable from the internet for each host.

Continue with [Installation](./installation.md).
