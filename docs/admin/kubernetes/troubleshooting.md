---
title: "Troubleshooting"
sidebar_position: 10
---

# Troubleshooting

## Certificate stuck (not `READY=True`)

```bash
kubectl -n prompt describe certificate prompt-tls-apex
kubectl -n prompt get challenge
```

Common causes:

- DNS for the host does not yet resolve to the Gateway address.
- Inbound port 80 is blocked, so the HTTP-01 challenge cannot be reached.
- The `http` listener is not programmed (check `kubectl get gateway`).
- Let's Encrypt rate limit hit - use the staging server while testing
  (`global.gateway.certManager.server`).

## Pods crash-looping at startup

- **`wait-for-db` init container hangs:** the database is not reachable. Check the CNPG cluster
  (`kubectl get cluster`) or the external host/credentials.
- **Backend exits during migration:** inspect logs; a "dirty" schema needs a manual
  `migrate force` (see [Database Operations](./database.md#migrations)).

## Database connection failures

```bash
kubectl -n prompt get cluster
kubectl -n prompt get secret prompt-db-core -o jsonpath='{.data.DB_NAME}' | base64 -d
```

Verify the pooler/cluster Service exists and the per-phase Secret points at it.

## Presigned upload/download fails in the browser

- **CORS error:** confirm the `SecurityPolicy` exists (Envoy) or CORS is configured on an
  external bucket, allowing the apex origin.
- **403 SignatureDoesNotMatch:** the `Host` header was rewritten. The S3 `HTTPRoute` must
  preserve it; do not add a hostname rewrite filter.
- **Wrong endpoint:** check `S3_PUBLIC_ENDPOINT` in the app ConfigMap matches the reachable
  `s3.<host>` (or external endpoint).

## A phase micro-frontend does not load

The core shell loads each phase via Module Federation from its public path. Confirm the phase is
enabled, its client Deployment is running, and its `HTTPRoute` resolves
(`curl -I https://<host>/<phase>/`). Check the browser console for a failed `remoteEntry.js`.

## Gateway not programmed

```bash
kubectl -n prompt get gateway prompt-gateway -o yaml | grep -A20 status:
```

Ensure Envoy Gateway is installed and the `GatewayClass` named in
`global.gateway.gatewayClassName` exists.

## Pending PVCs

```bash
kubectl -n prompt get pvc
```

A `Pending` PVC means no `StorageClass` can satisfy it. Set `global.postgresql.storageClass`
(and `infrastructure.seaweedfs.*`) to a class your cluster provides.
