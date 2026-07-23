---
title: "TLS & DNS"
sidebar_position: 5
---

# TLS & DNS

## Per-listener certificates

The Gateway exposes one HTTPS listener per host: `apex`, `s3.` (in-cluster object storage), and
`auth.` (in-cluster Keycloak). Each listener references its own certificate Secret, and
cert-manager issues each certificate independently via Let's Encrypt HTTP-01.

This is deliberate: a single multi-SAN certificate would let a validation failure on one optional
subdomain block TLS for the entire site. With per-listener certificates, an unconfigured
`auth.` DNS record never affects the apex certificate.

## cert-manager

The chart annotates the Gateway with `cert-manager.io/cluster-issuer` and, when
`global.gateway.certManager.create` is true, creates a `ClusterIssuer` using the HTTP-01
`gatewayHTTPRoute` solver. cert-manager spins up a temporary `HTTPRoute` on the `http` listener
for each challenge.

To reuse an existing issuer instead:

```yaml
global:
  gateway:
    certManager:
      create: false
      issuerName: my-existing-clusterissuer
```

## DNS

Point A/AAAA records at the Gateway's external address for every enabled host. Find the address
with:

```bash
kubectl -n prompt get gateway prompt-gateway -o wide
```

HTTP-01 requires inbound port 80 for each host. Wildcard certificates are **not** used (HTTP-01
cannot issue them); each host is validated on its own.

## Verifying issuance

```bash
kubectl -n prompt get certificate
kubectl -n prompt describe certificate prompt-tls-apex
```

A certificate stuck in `False`/`Issuing` usually means the HTTP-01 challenge cannot reach the
cluster (DNS not propagated, port 80 blocked, or the `http` listener not programmed). See
[Troubleshooting](./troubleshooting.md).

## Bring your own certificate

Disable cert-manager and reference pre-created TLS Secrets named `prompt-tls-apex`,
`prompt-tls-s3`, `prompt-tls-auth` (whichever listeners are enabled):

```yaml
global:
  gateway:
    certManager:
      enabled: false
```
