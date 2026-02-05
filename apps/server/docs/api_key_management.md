# API Key Management

> Secure API key generation, revocation, and lifecycle management

## Overview

The API Key Management system provides secure authentication and authorization for the Whatspire WhatsApp Service. API keys are cryptographically generated, hashed for storage, and support role-based access control with comprehensive audit logging.

## Features

- **Secure Generation** - Cryptographically random 32-byte keys (base64-encoded)
- **Role-Based Access** - Read, Write, and Admin permission levels
- **Immediate Revocation** - Deactivate compromised keys instantly
- **Audit Trail** - Complete history of key creation, usage, and revocation
- **Pagination & Filtering** - Efficient key listing with role/status filters
- **Masked Display** - Keys shown as `abcd1234...xyz9` for security

## Architecture

### Key Generation Flow

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ POST /api/apikeys
       │ { role: "read", description: "..." }
       ▼
┌─────────────────────────────────────────┐
│         APIKeyUseCase                   │
│  1. Generate 32-byte random key         │
│  2. Hash with SHA-256                   │
│  3. Store hash in database              │
│  4. Return plain key (ONLY ONCE)        │
└──────┬──────────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────────┐
│      APIKeyRepository (GORM)            │
│  - Stores: ID, KeyHash, Role, Metadata  │
│  - Indexes: key_hash, is_active         │
└─────────────────────────────────────────┘
```

### Authentication Flow

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ GET /api/sessions
       │ Header: X-API-Key: <plain_key>
       ▼
┌─────────────────────────────────────────┐
│      APIKeyMiddleware                   │
│  1. Extract key from header             │
│  2. Hash the key                        │
│  3. Lookup in database by hash          │
│  4. Check if active (not revoked)       │
│  5. Verify role permissions             │
│  6. Update last_used_at                 │
└──────┬──────────────────────────────────┘
       │ ✅ Authorized
       ▼
┌─────────────────────────────────────────┐
│         Handler Logic                   │
└─────────────────────────────────────────┘
```

## API Endpoints

### Create API Key

**POST** `/api/apikeys`

Creates a new API key with the specified role. The plain-text key is returned only once and cannot be retrieved later.

**Request:**

```json
{
  "role": "read",
  "description": "Production read-only key"
}
```

**Response:**

```json
{
  "success": true,
  "data": {
    "api_key": {
      "id": "key_c0d97bc3-559a-415c-9474-1f9519cfff39",
      "masked_key": "abcd1234...xyz9",
      "role": "read",
      "description": "Production read-only key",
      "created_at": "2026-02-04T10:30:00Z",
      "last_used_at": null,
      "is_active": true,
      "revoked_at": null,
      "revoked_by": null,
      "revocation_reason": null
    },
    "plain_key": "abcd1234efgh5678ijkl9012mnop3456qrst7890"
  }
}
```

**Roles:**

- `read` - Can view sessions, messages, contacts, groups
- `write` - Can send messages, create sessions, manage contacts
- `admin` - Full access including API key management

**Security Note:** Store the `plain_key` securely. It cannot be retrieved after this response.

---

### Revoke API Key

**DELETE** `/api/apikeys/:id`

Immediately revokes an API key, preventing further authentication. This action cannot be undone.

**Request:**

```json
{
  "reason": "Security audit - key rotation"
}
```

**Response:**

```json
{
  "success": true,
  "data": {
    "id": "key_c0d97bc3-559a-415c-9474-1f9519cfff39",
    "revoked_at": "2026-02-04T11:00:00Z",
    "revoked_by": "system"
  }
}
```

**Effects:**

- Key is immediately deactivated (`is_active = false`)
- All subsequent authentication attempts fail with `401 Unauthorized`
- Revocation is logged in audit trail
- Key remains in database for audit purposes

---

### List API Keys

**GET** `/api/apikeys`

Retrieves a paginated list of API keys with optional filtering.

**Query Parameters:**

- `page` (int, default: 1) - Page number
- `limit` (int, default: 50, max: 100) - Items per page
- `role` (string, optional) - Filter by role: `read`, `write`, `admin`
- `status` (string, optional) - Filter by status: `active`, `revoked`

**Example Request:**

```
GET /api/apikeys?page=1&limit=10&role=read&status=active
```

**Response:**

```json
{
  "success": true,
  "data": {
    "api_keys": [
      {
        "id": "key_c0d97bc3-559a-415c-9474-1f9519cfff39",
        "masked_key": "abcd1234...xyz9",
        "role": "read",
        "description": "Production read-only key",
        "created_at": "2026-02-04T10:30:00Z",
        "last_used_at": "2026-02-04T10:45:00Z",
        "is_active": true,
        "revoked_at": null,
        "revoked_by": null,
        "revocation_reason": null
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 25,
      "total_pages": 3
    }
  }
}
```

---

### Get API Key Details

**GET** `/api/apikeys/:id`

Retrieves detailed information about a specific API key including usage statistics.

**Response:**

```json
{
  "success": true,
  "data": {
    "api_key": {
      "id": "key_c0d97bc3-559a-415c-9474-1f9519cfff39",
      "masked_key": "abcd1234...xyz9",
      "role": "read",
      "description": "Production read-only key",
      "created_at": "2026-02-04T10:30:00Z",
      "last_used_at": "2026-02-04T10:45:00Z",
      "is_active": true,
      "revoked_at": null,
      "revoked_by": null,
      "revocation_reason": null
    },
    "usage_stats": {
      "total_requests": 1523,
      "last_7_days": 342
    }
  }
}
```

**Note:** Usage statistics are currently placeholder values. Full implementation requires querying the audit_logs table.

## Database Schema

### api_keys Table

```sql
CREATE TABLE api_keys (
    id TEXT PRIMARY KEY,
    key_hash TEXT NOT NULL UNIQUE,
    role TEXT NOT NULL,
    description TEXT,
    created_at DATETIME NOT NULL,
    last_used_at DATETIME,
    is_active BOOLEAN NOT NULL DEFAULT 1,
    revoked_at DATETIME,
    revoked_by TEXT,
    revocation_reason TEXT
);

CREATE INDEX idx_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX idx_api_keys_is_active ON api_keys(is_active);
CREATE INDEX idx_api_keys_created_at ON api_keys(created_at);
```

**Fields:**

- `id` - Unique identifier (format: `key_<timestamp>`)
- `key_hash` - SHA-256 hash of the plain-text key (64 hex chars)
- `role` - Permission level: `read`, `write`, `admin`
- `description` - Optional human-readable label
- `created_at` - Timestamp of key generation
- `last_used_at` - Last successful authentication timestamp
- `is_active` - Boolean flag (false when revoked)
- `revoked_at` - Timestamp of revocation
- `revoked_by` - User/system that revoked the key
- `revocation_reason` - Optional reason for revocation

## Security Considerations

### Key Storage

✅ **DO:**

- Store only SHA-256 hashes in the database
- Use cryptographically secure random generation (`crypto/rand`)
- Display keys only once during creation
- Mask keys in UI/logs (show first 8 + last 4 chars)

❌ **DON'T:**

- Store plain-text keys in database
- Log plain-text keys
- Send keys in URLs or query parameters
- Reuse revoked key IDs

### Key Transmission

- Always use HTTPS in production
- Send keys in `X-API-Key` header (not in URL)
- Implement rate limiting per API key
- Monitor for suspicious usage patterns

### Key Rotation

Best practices for key rotation:

1. **Generate new key** with same role
2. **Update client** with new key
3. **Verify new key** works correctly
4. **Revoke old key** with reason "Key rotation"
5. **Monitor audit logs** for any issues

Recommended rotation schedule:

- **Production keys**: Every 90 days
- **Development keys**: Every 180 days
- **Compromised keys**: Immediately

## Usage Examples

### cURL Examples

**Create API Key:**

```bash
curl -X POST http://localhost:8080/api/apikeys \
  -H "Content-Type: application/json" \
  -H "X-API-Key: <admin_key>" \
  -d '{
    "role": "read",
    "description": "Production read-only key"
  }'
```

**List Active Keys:**

```bash
curl -X GET "http://localhost:8080/api/apikeys?status=active" \
  -H "X-API-Key: <admin_key>"
```

**Revoke Key:**

```bash
curl -X DELETE http://localhost:8080/api/apikeys/key_c0d97bc3-559a-415c-9474-1f9519cfff39 \
  -H "Content-Type: application/json" \
  -H "X-API-Key: <admin_key>" \
  -d '{
    "reason": "Security audit - key rotation"
  }'
```

### Go Client Example

```go
package main

import (
    "context"
    "fmt"
    "whatspire/internal/application/usecase"
    "whatspire/internal/infrastructure/persistence"
)

func main() {
    // Initialize repository and use case
    repo := persistence.NewAPIKeyRepository(db)
    auditLogger := persistence.NewAuditLogger(db)
    apiKeyUC := usecase.NewAPIKeyUseCase(repo, auditLogger)

    // Create API key
    plainKey, apiKey, err := apiKeyUC.CreateAPIKey(
        context.Background(),
        "read",
        strPtr("Production key"),
        "admin@example.com",
    )
    if err != nil {
        panic(err)
    }

    fmt.Printf("API Key ID: %s\n", apiKey.ID)
    fmt.Printf("Plain Key (save this!): %s\n", plainKey)

    // List API keys
    keys, total, err := apiKeyUC.ListAPIKeys(
        context.Background(),
        1,    // page
        50,   // limit
        nil,  // role filter
        nil,  // status filter
    )
    if err != nil {
        panic(err)
    }

    fmt.Printf("Total keys: %d\n", total)
    for _, key := range keys {
        fmt.Printf("- %s (%s): %s\n", key.ID, key.Role, key.Description)
    }

    // Revoke API key
    reason := "Key rotation"
    revokedKey, err := apiKeyUC.RevokeAPIKey(
        context.Background(),
        apiKey.ID,
        "admin@example.com",
        &reason,
    )
    if err != nil {
        panic(err)
    }

    fmt.Printf("Revoked: %s at %s\n", revokedKey.ID, revokedKey.RevokedAt)
}

func strPtr(s string) *string {
    return &s
}
```

## Error Handling

### Common Error Codes

| Code                | HTTP Status | Description                                    |
| ------------------- | ----------- | ---------------------------------------------- |
| `VALIDATION_FAILED` | 400         | Invalid role or missing required fields        |
| `ALREADY_REVOKED`   | 400         | Attempting to revoke an already-revoked key    |
| `NOT_FOUND`         | 404         | API key ID not found                           |
| `DUPLICATE`         | 409         | Key hash already exists (extremely rare)       |
| `UNAUTHORIZED`      | 401         | Missing or invalid API key                     |
| `FORBIDDEN`         | 403         | Insufficient permissions (admin role required) |
| `INTERNAL_ERROR`    | 500         | Server error during key generation/storage     |

### Error Response Format

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_FAILED",
    "message": "invalid role: must be read, write, or admin",
    "details": {
      "role": "must be one of the allowed values: read write admin"
    }
  }
}
```

## Audit Logging

All API key operations are logged to the audit trail:

### Events Logged

1. **APIKeyCreated**
   - API key ID
   - Role
   - Description
   - Created by (user/system)
   - Timestamp

2. **APIKeyUsed**
   - API key ID
   - Endpoint accessed
   - HTTP method
   - IP address
   - Timestamp

3. **APIKeyRevoked**
   - API key ID
   - Revoked by (user/system)
   - Revocation reason
   - Timestamp

### Querying Audit Logs

```sql
-- Find all operations for a specific key
SELECT * FROM audit_logs
WHERE api_key_id = 'key_c0d97bc3-559a-415c-9474-1f9519cfff39'
ORDER BY created_at DESC;

-- Find recent revocations
SELECT * FROM audit_logs
WHERE event_type = 'api_key_revoked'
  AND created_at >= datetime('now', '-7 days')
ORDER BY created_at DESC;

-- Find keys with high usage
SELECT api_key_id, COUNT(*) as usage_count
FROM audit_logs
WHERE event_type = 'api_key_used'
  AND created_at >= datetime('now', '-7 days')
GROUP BY api_key_id
ORDER BY usage_count DESC
LIMIT 10;
```

## Troubleshooting

### Key Not Working After Creation

**Symptom:** `401 Unauthorized` immediately after creating key

**Causes:**

1. Key not copied correctly (whitespace/truncation)
2. Wrong header name (should be `X-API-Key`)
3. Database not updated (check logs)

**Solution:**

```bash
# Verify key exists in database
sqlite3 data/application.db "SELECT id, role, is_active FROM api_keys WHERE id = 'key_xxx';"

# Check if key is active
# is_active should be 1 (true)
```

### Revoked Key Still Working

**Symptom:** Revoked key continues to authenticate

**Causes:**

1. Middleware not checking `is_active` flag
2. Caching issue (if caching is implemented)
3. Database update failed

**Solution:**

```bash
# Verify key is revoked in database
sqlite3 data/application.db "SELECT id, is_active, revoked_at FROM api_keys WHERE id = 'key_xxx';"

# is_active should be 0 (false)
# revoked_at should have a timestamp

# Restart server to clear any in-memory state
```

### High Authentication Latency

**Symptom:** Slow API responses (>100ms)

**Causes:**

1. Missing database index on `key_hash`
2. Large number of keys (>10,000)
3. Slow disk I/O

**Solution:**

```bash
# Verify indexes exist
sqlite3 data/application.db ".indexes api_keys"

# Should show:
# idx_api_keys_key_hash
# idx_api_keys_is_active
# idx_api_keys_created_at

# Check query performance
sqlite3 data/application.db "EXPLAIN QUERY PLAN SELECT * FROM api_keys WHERE key_hash = 'xxx';"

# Should use index: SEARCH api_keys USING INDEX idx_api_keys_key_hash
```

## Performance Considerations

### Key Lookup Performance

- **Target:** <50ms for key authentication
- **Actual:** ~1-5ms with proper indexes
- **Bottleneck:** Database query on `key_hash`

**Optimization:**

- Ensure `idx_api_keys_key_hash` index exists
- Use connection pooling (GORM default)
- Consider Redis caching for high-traffic scenarios

### Pagination Performance

- **Target:** <100ms for listing 50 keys
- **Actual:** ~10-20ms with indexes
- **Bottleneck:** COUNT query for total

**Optimization:**

- Use `idx_api_keys_created_at` for sorting
- Cache total count for 60 seconds
- Implement cursor-based pagination for large datasets

## Future Enhancements

### Planned Features

1. **Key Expiration**
   - Auto-revoke keys after N days
   - Configurable per-key TTL
   - Email notifications before expiration

2. **Usage Quotas**
   - Rate limits per key
   - Request count limits
   - Bandwidth limits

3. **IP Whitelisting**
   - Restrict keys to specific IP ranges
   - CIDR notation support
   - Geo-blocking

4. **Key Scopes**
   - Fine-grained permissions beyond roles
   - Resource-level access control
   - Endpoint-specific permissions

5. **Webhook Notifications**
   - Alert on key creation
   - Alert on suspicious usage
   - Alert on revocation

## Related Documentation

- [API Specification](./api_specification.md) - Complete REST API reference
- [Configuration](./configuration.md) - Environment variables
- [Deployment Guide](./deployment_guide.md) - Production setup
- [Troubleshooting](./troubleshooting.md) - Common issues

## Support

For issues or questions:

1. Check [Troubleshooting](./troubleshooting.md)
2. Review audit logs for errors
3. Enable debug logging: `WHATSAPP_LOG_LEVEL=debug`
4. Contact support with API key ID (never send plain key)
