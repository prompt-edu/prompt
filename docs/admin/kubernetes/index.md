---
title: "Overview & Architecture"
sidebar_position: 1
description: Deploy PROMPT 2.0 on Kubernetes with the official Helm chart.
---

# Kubernetes Deployment

PROMPT 2.0 ships a reusable Helm chart (`charts/prompt`) for running the platform on Kubernetes.
It is the recommended path for institutions that already operate a cluster.
For the Docker Compose path see [Production Setup](/admin/productionSetup).

## What the chart deploys

The chart is an umbrella composed of one subchart per course phase plus a shared `infrastructure` subchart.

```text
charts/prompt/                 # umbrella
  charts/
    prompt-common/             # library chart: shared Deployment/Service/HTTPRoute templates
    core/                      # server-core + client-core (platform shell)
    assessment/ interview/ ... # one subchart per phase (server + micro-frontend)
    matching/ template/        # client-only phases
    infrastructure/            # CNPG, SeaweedFS, Gateway + TLS, optional Keycloak
```

Each phase contributes a Go backend Deployment and a micro-frontend Deployment, both fronted by
Gateway API `HTTPRoute`s. The frontends inject their runtime config (`env.js`) at container
start, so no rebuild is needed per environment.

## Architecture

```text
                       Internet
                          |
                 Envoy Gateway (Gateway API)
        apex listener      s3. listener      auth. listener
        (TLS)              (TLS)             (TLS, in-cluster KC)
          |                  |                  |
   HTTPRoutes           SeaweedFS S3       Keycloak (operator)
   /  -> client-core
   /api -> server-core
   /assessment -> client-assessment
   /assessment/api -> server-assessment
   ...                       |                  |
          \-----------------\|/-----------------/
                     CloudNativePG cluster
              (one logical database per phase)
```

- **Routing:** Gateway API. A single apex host serves the app and all phase paths; object
  storage and (optional) Keycloak get their own subdomain listeners.
- **Database:** one CloudNativePG cluster hosting a logical database per phase, fronted by a
  PgBouncer pooler.
- **TLS:** cert-manager issues a separate certificate per listener via Let's Encrypt HTTP-01.
- **Auth:** enforced inside the Go services (prompt-sdk); the gateway does not validate tokens.

## How it maps from Docker Compose

| Compose | Kubernetes |
| --- | --- |
| Traefik + labels | Envoy Gateway + `HTTPRoute` |
| One Postgres container per service | One CNPG cluster, one database per phase |
| Traefik ACME | cert-manager + Let's Encrypt |
| SeaweedFS services | SeaweedFS StatefulSets + S3 `HTTPRoute` |
| External/bundled Keycloak | `keycloak.mode: external` (default) or in-cluster operator |
| `.env` variables | shared ConfigMap + Secret (`envFrom`) + per-phase DB Secret |

## Pluggability

Every stateful dependency is switchable between an in-cluster bundle (self-contained installs)
and an external managed service (production):

- `global.postgresql.mode`: `in-cluster` (default) or `external`
- `global.objectStorage.mode`: `in-cluster` (default) or `external`
- `global.keycloak.mode`: `external` (default) or `in-cluster`

Continue with [Prerequisites](./prerequisites.md).
