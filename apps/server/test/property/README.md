# Property-Based Tests

## Known Issues

### Session Persistence Test Failures

**Issue**: `TestSessionPersistenceRoundTrip_Property2` fails with `SESSION_EXISTS` errors.

**Root Cause**: The property-based test framework (gopter) shrinks failing test cases to minimal examples, which can produce simple IDs like "9", "5", "7" that collide with sessions from previous test runs. The SQLite test database persists between runs, causing conflicts.

**Impact**: Low - Unit and integration tests pass. This affects only property-based edge case testing.

**Workaround**: Run tests with a fresh database or skip property tests:

```bash
# Skip property tests
go test -short ./test/unit/... ./test/integration/...

# Or clean database before running
rm -f test.db
go test -short ./test/property/...
```

**TODO**: Fix by either:

1. Using transaction-based test isolation with rollback
2. Generating truly unique IDs that don't shrink to colliding values
3. Cleaning database before each property test run
