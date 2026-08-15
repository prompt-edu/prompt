# Testing (all code)

- Run `make lint` and `make test` before considering a change done.
- **Go:** `cd servers/<service> && go test ./...`. Tests use `testcontainers-go` for DB isolation;
  seed data from `database_dumps/*.sql`; pattern `*_test.go` with `testutils.SetupTestDB()`.
- **End-to-end:** Playwright, in `e2e/`, one shard at a time. Run `make test-e2e-shard SHARD=<name>`
  (with no argument it prints the available shards), or `make test-e2e-shard PATHS="<pattern>"` for a
  narrower slice. Do **not** run `make test-e2e`: CI never runs the whole suite in one container
  either, it fans `e2e/shards.json` out one shard per service. See `e2e/README.md` for the stack it
  boots and the seeded users, and use the `e2e-testing` skill for Page Object Model, config, and
  flaky-test strategies.
- Add or update tests for behavior you change; don't mark work complete with failing tests.
