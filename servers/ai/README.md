# AI service

The AI service is the only component that talks to the model provider (Logos). Phase servers call
it with their service key, the user's Keycloak token and their own Logos key for the course phase;
this service stores no Logos keys. Every call is recorded in this service's own database before the
provider sees it. The audit records are browsed in `clients/ai_component`.

The routes and their roles, the request headers, the configuration and the key provisioning are
described in [`docs/contributor/architecture/ai-integration.md`](../../docs/contributor/architecture/ai-integration.md),
which is the one place to keep them current.

Run it with `make server-ai` (port 8092), test it with `make test-ai`. The Go tests start Postgres
and `aimock` through testcontainers, with the fixtures shared with e2e under `e2e/fixtures/aimock`.
