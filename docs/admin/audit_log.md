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

On a GitHub Actions deployment, these reach the container through the deploy workflow: set
`AUDIT_ENABLED` and `AUDIT_RETENTION_DAYS` as environment *variables* and `AUDIT_INGEST_KEYS` as an
environment *secret*. While the feature is off, the management console hides the audit log pages
instead of linking to endpoints that are not mounted.

## Phase services (reporting from course phases)

Phase microservices report their events to core over an authenticated HTTP endpoint. There is **no
Keycloak on this path** — each phase authenticates with its own shared secret:

- On **core**, list every reporting service and its key in `AUDIT_INGEST_KEYS`
  (e.g. `interview:<key1>,assessment:<key2>`). Two values may be listed per service to rotate keys
  without downtime.
- On **each phase**, set `AUDIT_ENABLED=true` and `AUDIT_INGEST_KEY=<that service's key>`.

Because keys are per-service, a leaked key only affects one service, and the reported `source` is
trustworthy (derived from which key matched).

**Transport security.** The shared secret travels in the `X-Audit-Token` header. Keep phase→core
traffic on the internal network (the default `SERVER_CORE_HOST=http://server-core:8080` stays inside
the container network, the same channel the user token already uses). If it must cross an untrusted
network, terminate it over HTTPS/TLS — otherwise the secret and event payload are sent in plaintext.
The SDK logs a warning at startup if the configured ingest URL is plaintext HTTP to a non-internal
host.

A phase reports whether audit is enabled through its status endpoint (`GET /info`) as the
`audit.log` capability, alongside other phase capabilities.

## Rotation

To rotate a service's key, add the new key alongside the old one in `AUDIT_INGEST_KEYS`, deploy the
phase with the new `AUDIT_INGEST_KEY`, then remove the old value from core.

## Privacy / GDPR

By design the audit log is **not** scrubbed by the privacy-deletion flow: entries (including the
snapshotted actor name/email and entity name) are retained until the retention window lapses, even
after the referenced subject's other data is deleted. The table is append-only, so entries cannot be
edited or anonymized in place after the fact.

This is a deliberate retention exception, not a blanket GDPR exemption. The GDPR right to erasure
(Art. 17) can be overridden only for a specific lawful basis - typically retention required by a
legal obligation (Art. 17(3)(b)) - and that basis, together with a defined retention schedule and a
process for handling erasure requests, is the operator's responsibility. Before enabling audit
logging in production:

- Confirm the lawful basis and retention period with your institution's legal/privacy office.
- Set `AUDIT_RETENTION_DAYS` to that period so entries expire on schedule (leaving it unset keeps
  personal data indefinitely, which is rarely defensible).
- Document how you handle erasure requests that touch audit data.

If your policy does not support retaining this data, keep `AUDIT_ENABLED` off.
