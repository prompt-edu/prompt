---
title: Audit Log
sidebar_position: 80
---

# Audit Log (developer guide)

Audit logging is provided by the shared `audit` package in `prompt-sdk`. Core and every phase use the
**same API**; only the sink (how an entry is persisted) differs — core writes to its own database,
phases ship entries to core over HTTP.

## You usually don't have to do anything

Once `audit.Middleware` is registered on a service (already done for core), **every mutating request
is captured automatically** — `POST`/`PUT`/`PATCH`/`DELETE` that succeed (2xx) or are denied (core
aborts with 403, the SDK auth middleware with 401). The actor, timestamp, route, and outcome are
filled in for you. Read requests and validation errors are ignored. An explicit `audit.Record` whose
request later fails is recorded with the real status and an `error` outcome, rather than a premature
success.

## Naming an action (recommended for core)

Auto-derived labels are readable but generic ("Created copy"). Attach a precise, one-line label to a
route with `audit.Describe`:

```go
router.POST("/publish", audit.Describe("Published grades"), publishHandler)
```

## Rich or background events

For high-stakes actions that need the specific entity or a change summary, or for background/async
work, emit an explicit event with `audit.Record`. This also **suppresses the automatic backstop
entry** for the request, so there is exactly one rich entry:

```go
audit.Record(c, audit.Event{
    Action:     "Published grades",
    EntityType: "grade",
    EntityID:   gradeID,
    EntityName: "team Alpha",       // snapshotted, human-readable subject
    Metadata:   map[string]any{"change": "grade updated"},
})
```

Keep `Metadata` to identifiers and short change summaries — **never raw sensitive payloads** (grade
values, note contents).

In PROMPT every action traces to a human, so **background work carries the initiating human's
actor** (capture it from the request and pass it into the job), never a synthetic "system" actor.

### Atomic events (core)

When an audit entry must not be lost even if something fails, use the transactional core helper so
the change and its audit row commit together:

```go
auditLog.RecordTx(ctx, queries.WithTx(tx), audit.Event{Action: "Deleted course", ...})
```

## Suppressing noise

- Silence a whole noisy route or group: `api.Group("/sync", audit.Skip())`.
- Silence a single request from inside a handler: `audit.Suppress(c)`.

An explicit `audit.Record` still writes even when the request is suppressed (you asked for that event
on purpose).

## Enabling a new phase service

In the phase's `main.go`:

```go
router.Use(audit.Middleware(audit.NewCoreSink(coreURL, "myservice")))
```

and advertise the capability in the service info endpoint:

```go
Capabilities: map[string]bool{
    promptTypes.CapabilityAuditLog: audit.Enabled(),
},
```

Then set `AUDIT_ENABLED` and `AUDIT_INGEST_KEY` for the service (see the admin guide). Under Docker
Compose the service's `environment:` block forwards `AUDIT_ENABLED` and maps its own
`AUDIT_INGEST_KEY_<SERVICE>` variable onto `AUDIT_INGEST_KEY`, so no two phases share a key; that
key also has to be listed in core's `AUDIT_INGEST_KEYS` under the service name passed to
`NewCoreSink`. When the toggle or key is missing, the middleware is a no-op, so wiring it in is
always safe.
