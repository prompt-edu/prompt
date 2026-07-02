# Testing (all code)

- Run `make lint` and `make test` before considering a change done.
- **Go:** `cd servers/<service> && go test ./...`. Tests use `testcontainers-go` for DB isolation;
  seed data from `database_dumps/*.sql`; pattern `*_test.go` with `testutils.SetupTestDB()`.
- **End-to-end:** Playwright. Use the `e2e-testing` skill for Page Object Model, config, and
  flaky-test strategies.
- Add or update tests for behavior you change; don't mark work complete with failing tests.
