---
title: "Object Storage"
sidebar_position: 7
---

# Object Storage

PROMPT stores uploaded files in S3-compatible object storage and hands browsers **presigned
URLs** to upload and download directly. Two endpoints matter:

- `S3_ENDPOINT` - internal, server-to-storage (set by the chart).
- `S3_PUBLIC_ENDPOINT` - baked into presigned URLs, so it must be browser-reachable with valid
  TLS.

## In-cluster (SeaweedFS, default)

The chart deploys SeaweedFS (master, volume, filer as StatefulSets; S3 gateway as a Deployment)
and exposes the S3 gateway on a dedicated `s3.<host>` listener. `S3_PUBLIC_ENDPOINT` is set to
`https://s3.<host>`.

Two details are essential for presigned URLs to work:

- **Host preservation.** The S3 `HTTPRoute` does not rewrite the `Host` header. Rewriting it
  would break the SigV4 signature.
- **CORS.** Browser uploads send a preflight `OPTIONS` request. The chart attaches an Envoy
  `SecurityPolicy` allowing the apex origin. With a non-Envoy gateway you must configure CORS
  yourself.

Storage sizes are tunable:

```yaml
infrastructure:
  seaweedfs:
    masterStorage: 1Gi
    volumeStorage: 50Gi
    filerStorage: 1Gi
```

## External (AWS S3, MinIO, Ceph RGW)

```yaml
global:
  objectStorage:
    mode: external
    bucket: prompt-files
    region: eu-central-1
    forcePathStyle: "true"   # false for AWS S3 virtual-hosted style
    external:
      endpoint: https://s3.eu-central-1.amazonaws.com
      publicEndpoint: https://s3.eu-central-1.amazonaws.com
```

Provide `S3_ACCESS_KEY` / `S3_SECRET_KEY` via the application secret. No in-cluster S3 resources
or `s3.` listener are created. Configure CORS on the bucket to allow the apex origin.

External object storage is also the recommended target for database backups
(see [Database Operations](./database.md#backups-wal--scheduled-base-backups)).
