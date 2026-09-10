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

## `helm install` fails before anything is created

The chart rejects configurations that cannot work, rather than installing something broken. The
error names the offending value:

- `... is required` - see [Required values](./configuration.md#required-values).
- `global.keycloak.mode=in-cluster requires global.postgresql.mode=in-cluster` - the bundled
  Keycloak needs the CNPG cluster. Use an external Keycloak, or run PostgreSQL in-cluster.

## Pods crash-looping at startup

- **`wait-for-db` init container hangs:** the init container runs `psql` against the phase's own
  database, so it also waits out a database or role that does not exist yet. On a fresh install
  that is normal for a minute while CloudNativePG reconciles the `Database` resources. If it never
  clears, check `kubectl get cluster,database` or the external host and credentials, and read the
  init container's log: `kubectl logs <pod> -c wait-for-db`.
- **Backend exits during migration:** inspect logs; a "dirty" schema needs a manual
  `migrate force` (see [Database Operations](./database.md#migrations)).
- **`CreateContainerConfigError: container has runAsNonRoot and image will run as root`:** a
  backend image was overridden with one that has no non-root user, or
  `global.podSecurity.runAsUser` was cleared. The chart supplies UID 65532 because the Go images
  build on distroless static.
- **Frontend pod logs `bind() to 0.0.0.0:80 failed (13: Permission denied)`:** something is
  dropping capabilities from the frontend containers. The stock nginx base needs
  `NET_BIND_SERVICE`, `SETUID`/`SETGID` and `DAC_OVERRIDE`; do not enforce the `restricted` Pod
  Security Standard on this namespace (see [Prerequisites](./prerequisites.md)).
- **Pods rejected by PodSecurity admission:** the chart targets `baseline`, not `restricted`.

## A config or secret change had no effect

Workload pod templates carry `checksum/appconfig`, `checksum/appsecrets` and (backends)
`checksum/db` annotations, so `helm upgrade` restarts the pods when those values change. Two cases
are not covered: `global.appSecrets.existingSecret` (the chart cannot see the content) and an
in-cluster role password edited by hand. Restart manually:

```bash
kubectl -n prompt rollout restart deployment -l app.kubernetes.io/part-of=prompt
```

## Database connection failures

```bash
kubectl -n prompt get cluster
kubectl -n prompt get secret prompt-db-core -o jsonpath='{.data.DB_NAME}' | base64 -d
```

Verify the pooler/cluster Service exists and the per-phase Secret points at it.

## Presigned upload/download fails in the browser

- **CORS error:** confirm the `SecurityPolicy` exists (Envoy) or CORS is configured on an
  external bucket, allowing the apex origin. The policy is gated on
  `global.gateway.provider: envoygateway`, not on rate limiting, so turning the rate limit off does
  not remove it.
- **403 SignatureDoesNotMatch:** the `Host` header was rewritten. The S3 `HTTPRoute` must
  preserve it; do not add a hostname rewrite filter.
- **Wrong endpoint:** check `S3_PUBLIC_ENDPOINT` in the app ConfigMap matches the reachable
  `s3.<host>` (or external endpoint).

## A phase micro-frontend does not load

The core shell loads each phase via Module Federation from its public path. Confirm the phase is
enabled, its client Deployment is running, and its `HTTPRoute` resolves
(`curl -I https://<host>/<phase>/`). Check the browser console for a failed `remoteEntry.js`.

If the console shows a `ScriptExternalLoadError` for `/intro-course-developer` or
`/github-challenge`, that is expected: those remotes come from separate deployments this chart does
not install, and their paths are compiled into the core bundle. See
[Components this chart does not deploy](./configuration.md#components-this-chart-does-not-deploy).

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
