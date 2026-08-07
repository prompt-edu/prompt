---
title: Audit Log
sidebar_position: 60
---

# Audit Log (administration)

The audit log is a centralized, append-only record of important actions. It is **off by default** and
enabled per service by an administrator.

## Enabling audit logging

Audit logging is controlled by environment variables. Set them on the **core** service:

| Variable | Purpose |
| --- | --- |
| `AUDIT_ENABLED` | Set to `true` to turn the feature on. When unset, nothing is captured or exposed. |
| `AUDIT_RETENTION_DAYS` | Retention window in days. **Unset means entries are never pruned** (kept forever). |
| `AUDIT_INGEST_KEYS` | Per-service shared secrets that let phase services report events, formatted `service1:key1,service2:key2`. |

Enabling requires setting `AUDIT_ENABLED=true`. Missing configuration fails safe (feature stays off).

## Phase services (reporting from course phases)

Phase microservices report their events to core over an authenticated HTTP endpoint. There is **no
Keycloak on this path** — each phase authenticates with its own shared secret:

- On **core**, list every reporting service and its key in `AUDIT_INGEST_KEYS`
  (e.g. `interview:<key1>,assessment:<key2>`). Two values may be listed per service to rotate keys
  without downtime.
- On **each phase**, set `AUDIT_ENABLED=true` and `AUDIT_INGEST_KEY=<that service's key>`.

Because keys are per-service, a leaked key only affects one service, and the reported `source` is
trustworthy (derived from which key matched). Keep the ingest endpoint on the internal network.

A phase reports whether audit is enabled through its status endpoint (`GET /info`) as the
`audit.log` capability, alongside other phase capabilities.

## Rotation

To rotate a service's key, add the new key alongside the old one in `AUDIT_INGEST_KEYS`, deploy the
phase with the new `AUDIT_INGEST_KEY`, then remove the old value from core.

## Privacy / GDPR

The audit log is **exempt from privacy-deletion requests**. Entries are *not* removed when a
subject's data is deleted — they are retained as a compliance record (lawful basis:
accountability/legal obligation) until the retention window lapses. Set `AUDIT_RETENTION_DAYS`
according to your institution's policy.
