---
paths:
  - "servers/**/*.go"
---

# Go Patterns

- **Prefer `prompt-sdk` over custom implementations** for auth, CORS, inter-service fetching,
  validation, and shared domain types — see `prompt-sdk.md` for the API catalog.
- **Construction is separate from route registration.** A module exposes
  `NewXService(queries, conn, deps...)` returning the service, and
  `RegisterRoutes(routerGroup, service, authMiddleware)` mounting it. No package-level
  `XServiceSingleton`, no `InitXModule` that does both — a test must be able to obtain a service
  without building a router. `main.go` constructs every service in one block, then registers every
  route group, and is the only place that knows the wiring order. Reference: `servers/presentation`,
  and `servers/example_server` as the scaffolding template.
- **Handlers are methods on the service they use** (`func (s *XService) getThing(c *gin.Context)`),
  registered as `service.getThing`. Reach state through the receiver, never a global. A separate
  `handler` struct earns its place only when it composes more than one collaborator, or when the
  service type lives in another package and cannot take methods (see `servers/core/auth`).
- Gin handlers: parse/validate input → call `service.go` business logic → return JSON. Keep
  handlers thin; business logic lives in `service.go`, validation in `validation.go`. Gin stops at
  `router.go`: domain methods take `context.Context` and typed parameters, not `*gin.Context`.
- DB access goes through generated sqlc methods (`sqlc.md`), never hand-written SQL strings in Go.
- Use `DeferDBRollback(tx, ctx)` for safe transaction rollback.
- Phase services expose config + copy endpoints via the SDK:
  `RegisterConfigEndpoint(...)`, `RegisterCopyEndpoint(...)` with `PhaseConfigHandler` /
  `PhaseCopyHandler` implementations (see the `config/` and `copy/` packages in `example_server`).
  Both are structural interfaces, so implement `HandlePhaseConfig` / `HandlePhaseCopy` on the
  service and pass the service itself; don't wrap it in a zero-value handler struct, which compiles
  and then dereferences nil at request time. The privacy endpoints take function types, so pass the
  method values (`service.DataExportHandler`) directly.
- Custom validators are registered via the SDK `utils` subpackage (`RegisterValidation`,
  `ValidateStruct`); built-ins include `matriculationNumber` and `universityLogin`.
- **Audit logging** is automatic via the shared `prompt-sdk/audit` middleware — mutating requests are
  captured with no per-route work. Name important actions with `audit.Describe("…")`, emit rich or
  background events with `audit.Record`/`RecordTx`, and silence noisy routes with `audit.Skip()`.
  Keep `Metadata` to identifiers/change-summaries, never raw sensitive values. See
  `docs/contributor/guide/audit-log.md`.
