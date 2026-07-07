---
paths:
  - "servers/**/*.go"
---

# Go Patterns

- **Prefer `prompt-sdk` over custom implementations** for auth, CORS, inter-service fetching,
  validation, and shared domain types — see `prompt-sdk.md` for the API catalog.
- Gin handlers: parse/validate input → call `service.go` business logic → return JSON. Keep
  handlers thin; business logic lives in `service.go`, validation in `validation.go`.
- DB access goes through generated sqlc methods (`sqlc.md`), never hand-written SQL strings in Go.
- Use `DeferDBRollback(tx, ctx)` for safe transaction rollback.
- Phase services expose config + copy endpoints via the SDK:
  `RegisterConfigEndpoint(...)`, `RegisterCopyEndpoint(...)` with `PhaseConfigHandler` /
  `PhaseCopyHandler` implementations (see the `config/` and `copy/` packages in `template_server`).
- Custom validators are registered via the SDK `utils` subpackage (`RegisterValidation`,
  `ValidateStruct`); built-ins include `matriculationNumber` and `universityLogin`.
