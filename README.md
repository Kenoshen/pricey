## Running the Postgres contract tests

The contract test suite in `store_contract_test.go` validates that the `*Postgres` implementation is behaviourally equivalent to the `Store` interface. It requires a live Postgres database and is skipped automatically when the environment variable is not set.

```bash
TEST_POSTGRES_DSN="postgres://user:pass@localhost:5432/testdb?sslmode=disable" go test ./...
```

The schema is applied automatically at the start of the test run, so no manual migration step is needed. All tables are truncated between individual sub-tests to keep them independent.
