# Testing (all code)

- Run `make lint` and `make test` before considering a change done.
- **Go:** `cd servers/<service> && go test ./...`. Tests use `testcontainers-go` for DB isolation;
  seed data from `database_dumps/*.sql`; pattern `*_test.go` with `testutils.SetupTestDB()`.
- **End-to-end:** Playwright, in `e2e/`. Run with `make test-e2e` (or `make test-e2e-ui`); see
  `e2e/README.md` for the stack it boots, the seeded users, and how to add a test.
- Add or update tests for behavior you change; don't mark work complete with failing tests.
