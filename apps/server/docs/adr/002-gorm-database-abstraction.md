# ADR-002: GORM for Database Abstraction

**Date**: 2026-02-03  
**Status**: Accepted  
**Deciders**: Development Team  
**Technical Story**: Production Deployment - Database Flexibility

---

## Context and Problem Statement

The application initially used SQLite for development simplicity, but production deployments require PostgreSQL for scalability and reliability. We needed a database abstraction layer that supports multiple databases while maintaining type safety and good performance. Which ORM or database library should we use?

## Decision Drivers

- Support for both SQLite (development) and PostgreSQL (production)
- Type-safe query building
- Migration support
- Good performance (queries < 100ms p95)
- Active maintenance and community support
- Minimal boilerplate code
- Support for complex queries and relationships

## Considered Options

- **Option 1**: GORM (Go ORM)
- **Option 2**: sqlx (SQL extension library)
- **Option 3**: database/sql (standard library only)
- **Option 4**: sqlc (SQL compiler)

## Decision Outcome

Chosen option: "**GORM**", because it provides the best balance of productivity, type safety, and multi-database support while maintaining good performance for our use case.

### Positive Consequences

- **Multi-Database Support**: Single codebase works with SQLite and PostgreSQL
- **Auto-Migrations**: Schema changes managed automatically
- **Type Safety**: Compile-time checking of queries
- **Productivity**: Less boilerplate than raw SQL
- **Relationships**: Easy handling of associations
- **Hooks**: Support for lifecycle callbacks
- **Performance**: Query performance meets requirements (<100ms p95)

### Negative Consequences

- **Magic**: Some implicit behavior can be surprising
- **Performance**: Slightly slower than raw SQL for complex queries
- **Learning Curve**: Need to understand GORM conventions
- **Debugging**: Generated SQL can be hard to inspect
- **Flexibility**: Some complex queries require raw SQL

## Pros and Cons of the Options

### Option 1: GORM

- Good, because supports multiple databases with same code
- Good, because auto-migrations reduce manual schema management
- Good, because type-safe query building
- Good, because active community and good documentation
- Good, because hooks for lifecycle events
- Bad, because some implicit behavior (soft deletes, auto-timestamps)
- Bad, because performance overhead for simple queries
- Bad, because can generate inefficient SQL if not careful

### Option 2: sqlx

- Good, because minimal overhead over database/sql
- Good, because explicit SQL queries
- Good, because good performance
- Bad, because no auto-migrations
- Bad, because manual schema management
- Bad, because more boilerplate code
- Bad, because database-specific SQL required

### Option 3: database/sql

- Good, because no dependencies
- Good, because maximum control
- Good, because best performance
- Bad, because massive boilerplate
- Bad, because no type safety
- Bad, because manual connection pooling
- Bad, because no migration support
- Bad, because database-specific code

### Option 4: sqlc

- Good, because generates type-safe Go from SQL
- Good, because explicit SQL queries
- Good, because good performance
- Bad, because requires separate SQL files
- Bad, because no auto-migrations
- Bad, because less flexible than GORM
- Bad, because smaller community

## Implementation Details

### Dialector Pattern

We use GORM's dialector pattern to support multiple databases:

```go
type DatabaseDialector interface {
    GetDialector() gorm.Dialector
    GetDSN() string
}

type SQLiteDialector struct { ... }
type PostgreSQLDialector struct { ... }
```

### Configuration

```yaml
database:
  driver: postgres # or sqlite
  dsn: "host=localhost user=postgres password=secret dbname=whatspire"
```

### Migration Strategy

We use GORM's AutoMigrate for schema management:

```go
db.AutoMigrate(
    &models.Session{},
    &models.Event{},
    &models.Presence{},
    // ...
)
```

### Performance Benchmarks

Actual performance results:

- Event GetByID: 0.016ms (target: <100ms) ✅
- Event List (100 records): 1.04ms (target: <100ms) ✅
- Session queries: <5ms average ✅

## Database-Specific Considerations

### SQLite

- **Use Case**: Development, testing, small deployments
- **Advantages**: Zero configuration, single file
- **Limitations**: No concurrent writes, limited scalability

### PostgreSQL

- **Use Case**: Production deployments
- **Advantages**: ACID compliance, concurrent writes, scalability
- **Configuration**: Connection pooling, prepared statements

## Migration Path

If GORM becomes a bottleneck:

1. **Optimize GORM queries**: Use preloading, select specific fields
2. **Add raw SQL**: For complex queries, use `db.Raw()`
3. **Hybrid approach**: GORM for CRUD, raw SQL for analytics
4. **Replace gradually**: Migrate hot paths to sqlx or database/sql

## Links

- [GORM Documentation](https://gorm.io/)
- Related: ADR-001 (Clean Architecture)
- Related: ADR-005 (Event Persistence Strategy)

---

## Notes

### Query Optimization

Best practices we follow:

- Use `Select()` to fetch only needed fields
- Use `Preload()` for relationships to avoid N+1 queries
- Add indexes for frequently queried fields
- Use `Find()` with limits for pagination
- Monitor query performance with benchmarks

### Testing Strategy

- **Unit Tests**: Use in-memory SQLite
- **Integration Tests**: Use PostgreSQL in Docker
- **Benchmarks**: Track query performance over time

### Known Issues

- **Soft Deletes**: GORM's default soft delete can be surprising
  - Solution: Explicitly use `Unscoped()` when needed
- **Auto-Timestamps**: CreatedAt/UpdatedAt added automatically
  - Solution: Documented in model definitions
- **Preloading**: Can generate inefficient queries
  - Solution: Use `Joins()` for simple relationships

### Future Considerations

- **Read Replicas**: GORM supports multiple databases
- **Sharding**: Can implement custom dialector
- **Caching**: Add Redis layer for hot data
- **Analytics**: Use separate read-optimized database
