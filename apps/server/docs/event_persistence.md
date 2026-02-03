# Event Persistence Documentation

## Overview

The event persistence feature provides comprehensive event storage, querying, and replay capabilities for debugging and auditing purposes. All WhatsApp events (messages, reactions, presence updates, etc.) can be stored in the database and queried later.

## Architecture

### Components

1. **Event Entity** (`internal/domain/entity/event.go`)
   - Core domain model representing an event
   - Contains: ID, Type, SessionID, Data (JSON), Timestamp

2. **Event Repository** (`internal/infrastructure/persistence/event_repository.go`)
   - Handles CRUD operations for events
   - Supports both SQLite and PostgreSQL
   - Optimized with composite indexes

3. **Event Use Case** (`internal/application/usecase/event_usecase.go`)
   - Business logic for event operations
   - Methods: QueryEvents, GetEventByID, ReplayEvents, DeleteOldEvents

4. **Event Handlers** (`internal/presentation/http/handler_events.go`)
   - HTTP endpoints for event operations
   - Role-based authorization

5. **Event Cleanup Job** (`internal/infrastructure/jobs/event_cleanup.go`)
   - Automated retention policy enforcement
   - Scheduled daily cleanup

## Configuration

### Environment Variables

```bash
# Enable event persistence
WHATSAPP_EVENTS_ENABLED=true

# Retention period (0 = keep forever)
WHATSAPP_EVENTS_RETENTION_DAYS=30

# Daily cleanup time (UTC, HH:MM format)
WHATSAPP_EVENTS_CLEANUP_TIME=02:00

# Cleanup check interval
WHATSAPP_EVENTS_CLEANUP_INTERVAL=1h
```

### Configuration Structure

```go
type EventsConfig struct {
    Enabled         bool          // Enable event storage
    RetentionDays   int           // Days to retain events (0 = forever)
    CleanupTime     string        // Daily cleanup time in UTC (HH:MM)
    CleanupInterval time.Duration // Cleanup check interval
}
```

## Database Schema

### Events Table

```sql
CREATE TABLE events (
    id VARCHAR(255) PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    session_id VARCHAR(255) NOT NULL,
    data JSONB,
    timestamp TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Composite indexes for optimal query performance
CREATE INDEX idx_events_session_timestamp ON events(session_id, timestamp DESC);
CREATE INDEX idx_events_type_timestamp ON events(type, timestamp DESC);
CREATE INDEX idx_events_session_type ON events(session_id, type);
```

### Performance

- **GetByID**: ~0.016ms (primary key lookup)
- **List with filters**: ~1.04ms (indexed queries)
- **Target**: <100ms p95 latency ✅

## API Endpoints

### 1. Query Events

**Endpoint**: `GET /api/events`

**Authorization**: Read role required

**Query Parameters**:

- `session_id` (optional): Filter by session ID
- `event_type` (optional): Filter by event type
- `since` (optional): Start timestamp (RFC3339)
- `until` (optional): End timestamp (RFC3339)
- `limit` (optional): Max results (1-1000, default: 100)
- `offset` (optional): Pagination offset

**Example Request**:

```bash
curl -H "X-API-Key: your-read-key" \
  "http://localhost:8080/api/events?session_id=session123&limit=50"
```

**Example Response**:

```json
{
  "success": true,
  "data": {
    "events": [
      {
        "id": "evt_123",
        "type": "message.received",
        "session_id": "session123",
        "data": {...},
        "timestamp": "2026-02-03T10:30:00Z"
      }
    ],
    "total": 150,
    "limit": 50,
    "offset": 0
  }
}
```

### 2. Get Event by ID

**Endpoint**: `GET /api/events/:id`

**Authorization**: Read role required

**Example Request**:

```bash
curl -H "X-API-Key: your-read-key" \
  "http://localhost:8080/api/events/evt_123"
```

**Example Response**:

```json
{
  "success": true,
  "data": {
    "id": "evt_123",
    "type": "message.received",
    "session_id": "session123",
    "data": {...},
    "timestamp": "2026-02-03T10:30:00Z"
  }
}
```

### 3. Replay Events

**Endpoint**: `POST /api/events/replay`

**Authorization**: Admin role required

**Request Body**:

```json
{
  "session_id": "session123",
  "event_type": "message.received",
  "since": "2026-02-03T00:00:00Z",
  "until": "2026-02-03T23:59:59Z",
  "dry_run": true
}
```

**Note**: At least one filter (session_id, event_type, since, or until) is required.

**Example Request**:

```bash
curl -X POST \
  -H "X-API-Key: your-admin-key" \
  -H "Content-Type: application/json" \
  -d '{"session_id":"session123","dry_run":true}' \
  "http://localhost:8080/api/events/replay"
```

**Example Response (Dry Run)**:

```json
{
  "success": true,
  "data": {
    "success": true,
    "events_found": 42,
    "events_replayed": 0,
    "dry_run": true,
    "message": "Dry run: would replay 42 events"
  }
}
```

**Example Response (Actual Replay)**:

```json
{
  "success": true,
  "data": {
    "success": true,
    "events_found": 42,
    "events_replayed": 42,
    "events_failed": 0,
    "dry_run": false,
    "message": "Replayed 42 events successfully"
  }
}
```

## Event Types

The following event types are persisted:

- `message.received` - Incoming message
- `message.sent` - Outgoing message
- `message.delivered` - Message delivery confirmation
- `message.read` - Message read receipt
- `message.reaction` - Reaction to a message
- `presence.update` - User presence change
- `session.connected` - Session connected
- `session.disconnected` - Session disconnected
- `session.qr` - QR code generated
- `group.joined` - Joined a group
- `group.left` - Left a group
- `contact.update` - Contact information updated

## Retention Policy

### Automatic Cleanup

The cleanup job runs automatically based on configuration:

1. **Check Interval**: Runs every `cleanup_interval` (default: 1 hour)
2. **Scheduled Time**: Only executes at or after `cleanup_time` (default: 02:00 UTC)
3. **Daily Execution**: Runs once per day at the scheduled time
4. **Retention Period**: Deletes events older than `retention_days`

### Manual Cleanup

You can also trigger cleanup manually via the use case:

```go
deleted, err := eventUseCase.DeleteOldEvents(ctx, 30) // Delete events older than 30 days
```

### Disable Retention

Set `retention_days` to `0` to keep events forever:

```bash
WHATSAPP_EVENTS_RETENTION_DAYS=0
```

## Use Cases

### 1. Debugging Message Delivery Issues

Query all events for a specific session to trace message flow:

```bash
curl -H "X-API-Key: your-read-key" \
  "http://localhost:8080/api/events?session_id=session123&event_type=message.sent&limit=100"
```

### 2. Auditing User Activity

Query all events within a time range:

```bash
curl -H "X-API-Key: your-read-key" \
  "http://localhost:8080/api/events?since=2026-02-01T00:00:00Z&until=2026-02-03T23:59:59Z"
```

### 3. Replaying Events for Testing

Replay events to test webhook or WebSocket delivery:

```bash
# Dry run first to see what would be replayed
curl -X POST \
  -H "X-API-Key: your-admin-key" \
  -H "Content-Type: application/json" \
  -d '{"session_id":"session123","dry_run":true}' \
  "http://localhost:8080/api/events/replay"

# Actual replay
curl -X POST \
  -H "X-API-Key: your-admin-key" \
  -H "Content-Type: application/json" \
  -d '{"session_id":"session123","dry_run":false}' \
  "http://localhost:8080/api/events/replay"
```

### 4. Investigating Reaction Issues

Query all reaction events:

```bash
curl -H "X-API-Key: your-read-key" \
  "http://localhost:8080/api/events?event_type=message.reaction"
```

## Performance Considerations

### Indexes

The events table uses composite indexes for optimal query performance:

1. `(session_id, timestamp DESC)` - Session-based queries
2. `(type, timestamp DESC)` - Type-based queries
3. `(session_id, type)` - Combined session and type queries

### Query Limits

- Maximum limit per query: 1000 events
- Default limit: 100 events
- Use pagination (offset) for large result sets

### Replay Safety

- Maximum events per replay: 1000 (safety limit)
- Use filters to narrow down replay scope
- Always use dry_run first to verify

## Monitoring

### Cleanup Job Status

The cleanup job logs its activity:

```
✅ Event cleanup job started (retention: 30 days, interval: 1h, cleanup time: 02:00)
🧹 Starting event cleanup (retention: 30 days)...
✅ Event cleanup completed: deleted 1234 events (duration: 45ms)
```

### Event Persistence Status

Check if event persistence is enabled:

```
✅ Event persistence enabled
```

## Troubleshooting

### Events Not Being Persisted

1. Check if event persistence is enabled:

   ```bash
   echo $WHATSAPP_EVENTS_ENABLED
   ```

2. Check database connection and migrations:

   ```bash
   # Check if events table exists
   psql -d your_database -c "\dt events"
   ```

3. Check application logs for errors:
   ```
   ⚠️  Failed to persist event: <error>
   ```

### Cleanup Not Running

1. Verify retention days is not 0:

   ```bash
   echo $WHATSAPP_EVENTS_RETENTION_DAYS
   ```

2. Check cleanup time configuration:

   ```bash
   echo $WHATSAPP_EVENTS_CLEANUP_TIME
   ```

3. Check application logs:
   ```
   ✅ Event cleanup job started (retention: 30 days, interval: 1h, cleanup time: 02:00)
   ```

### Query Performance Issues

1. Verify indexes exist:

   ```sql
   SELECT indexname FROM pg_indexes WHERE tablename = 'events';
   ```

2. Use EXPLAIN to analyze query plans:

   ```sql
   EXPLAIN ANALYZE SELECT * FROM events WHERE session_id = 'session123' ORDER BY timestamp DESC LIMIT 100;
   ```

3. Consider reducing retention period to limit table size

### Replay Failures

1. Check event publisher health:

   ```bash
   curl http://localhost:8080/ready
   ```

2. Verify webhook/WebSocket configuration

3. Check for partial failures in response:
   ```json
   {
     "events_replayed": 40,
     "events_failed": 2,
     "message": "Replayed 40 events successfully, 2 failed"
   }
   ```

## Security

### Role-Based Access Control

- **Read Role**: Can query and view events
- **Admin Role**: Can replay events (potentially triggers actions)

### API Key Configuration

```bash
# Read-only key
WHATSAPP_API_KEYS_MAP='[{"key":"read-key","role":"read"}]'

# Admin key
WHATSAPP_API_KEYS_MAP='[{"key":"admin-key","role":"admin"}]'
```

### Data Privacy

- Event data contains message content and user information
- Implement appropriate retention policies
- Consider encryption at rest for sensitive data
- Audit access to event endpoints

## Migration Guide

### Enabling Event Persistence

1. Update configuration:

   ```bash
   export WHATSAPP_EVENTS_ENABLED=true
   export WHATSAPP_EVENTS_RETENTION_DAYS=30
   ```

2. Restart application (migrations run automatically)

3. Verify events table created:
   ```sql
   SELECT COUNT(*) FROM events;
   ```

### Disabling Event Persistence

1. Update configuration:

   ```bash
   export WHATSAPP_EVENTS_ENABLED=false
   ```

2. Restart application

3. Optionally drop events table:
   ```sql
   DROP TABLE IF EXISTS events;
   ```

## Best Practices

1. **Set Appropriate Retention**: Balance storage costs with debugging needs (30-90 days typical)

2. **Use Filters**: Always filter queries by session_id or event_type for better performance

3. **Dry Run First**: Always test replay with dry_run=true before actual replay

4. **Monitor Storage**: Track events table size and adjust retention as needed

5. **Index Maintenance**: Regularly analyze and vacuum the events table (PostgreSQL)

6. **Backup Strategy**: Include events table in backup procedures if long-term retention needed

7. **Access Control**: Restrict admin role to trusted users only

## Future Enhancements

Potential improvements for future versions:

- Event aggregation and analytics
- Event search by message content
- Event export to external systems
- Real-time event streaming
- Event-based alerting
- Compression for old events
- Partitioning for large tables

## References

- [Database Migrations Guide](database_migrations.md)
- [Deployment Guide](deployment_guide.md)
- [API Specification](api_specification.md)
- [Configuration Guide](configuration.md)
