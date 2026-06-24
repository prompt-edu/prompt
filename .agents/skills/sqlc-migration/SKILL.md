---
name: sqlc-migration
description: Add or change a database schema in a PROMPT 2.0 Go service using golang-migrate + sqlc — write the migration and queries, regenerate type-safe Go, and stay green on the migration CI check. Use when adding a table/column, writing SQL queries, or running sqlc.
argument-hint: <server name, e.g. core | assessment | interview>
---

Schema and queries for a Go microservice live under `servers/<service>/db/`. Migrations run via
`golang-migrate` (auto-applied on startup), and `sqlc` generates type-safe Go from the schema +
queries. Work in the target service directory.

## 1. Write the migration

- Migrations are in `servers/<service>/db/migration/`, numbered sequentially:
  `NNNN_short_description.up.sql` (zero-padded 4 digits, next number after the highest existing one).
- Down migrations are OPTIONAL and inconsistent across services — only add a matching
  `NNNN_short_description.down.sql` if that service already ships `.down.sql` files
  (`ls servers/<service>/db/migration/*.down.sql`). If it doesn't, stay up-only to match convention.
- Keep one logical change per migration; never edit an already-merged migration — add a new one.

## 2. Write queries

- Add/extend SQL in `servers/<service>/db/query/*.sql` with sqlc annotations, e.g.:
  ```sql
  -- name: GetTutors :many
  SELECT * FROM tutors WHERE course_phase_id = $1;
  ```
  Use `:one`, `:many`, `:exec`, `:execrows` as appropriate.

## 3. Regenerate Go

- From `servers/<service>/`: `sqlc generate` (config in `sqlc.yaml`; output lands in `db/sqlc/`,
  package `db`, pgx/v5). Or `make sqlc-<service>` from the repo root.
- Use generated methods in services:
  ```go
  import db "github.com/prompt-edu/prompt/servers/<service>/db/sqlc"
  tutors, err := queries.GetTutors(ctx, coursePhaseID)
  ```

## 4. Verify (matches CI)

- `cd servers/<service> && go build ./... && go test ./...` (tests use testcontainers-go).
- The `detect-migrations.yml` workflow flags migration changes — ensure new files follow the
  `NNNN_*.up.sql` naming and that `db/sqlc/` was regenerated and committed alongside the migration.
- Never commit a migration without its regenerated `db/sqlc/` changes.
