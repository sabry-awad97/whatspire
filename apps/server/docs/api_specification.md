# API Specification

> Whatspire WhatsApp Service REST API Reference

**Base URL**: `http://localhost:8080`  
**Content-Type**: `application/json`  
**Authentication**: `X-API-Key` header (when enabled)

---

## Authentication

When API key authentication is enabled, include the header:

```
X-API-Key: your-api-key-here
```

### Roles

| Role    | Permissions                                     |
| ------- | ----------------------------------------------- |
| `read`  | Read-only endpoints (contacts, chats, profiles) |
| `write` | Read + Send messages, reactions, receipts       |
| `admin` | Full access including session management        |

---

## Health Endpoints

### GET /health

Liveness probe - returns service status.

**Response** `200 OK`

```json
{
  "status": "healthy",
  "timestamp": "2026-02-03T13:30:00Z"
}
```

### GET /ready

Readiness probe - checks all dependencies.

**Response** `200 OK`

```json
{
  "status": "ready",
  "checks": {
    "whatsapp_client": "healthy",
    "event_publisher": "healthy"
  }
}
```

### GET /metrics

Prometheus metrics endpoint.

---

## Session Management (Admin Role)

### POST /api/internal/sessions/register

Register a new WhatsApp session.

**Request**

```json
{
  "session_id": "session-123",
  "name": "My WhatsApp"
}
```

**Response** `201 Created`

```json
{
  "id": "session-123",
  "status": "pending",
  "created_at": "2026-02-03T13:30:00Z"
}
```

### POST /api/internal/sessions/:id/reconnect

Reconnect a disconnected session.

**Response** `200 OK`

```json
{
  "id": "session-123",
  "status": "connected",
  "jid": "1234567890@s.whatsapp.net"
}
```

### POST /api/internal/sessions/:id/disconnect

Disconnect session (keeps credentials).

**Response** `200 OK`

```json
{
  "id": "session-123",
  "status": "disconnected"
}
```

### POST /api/internal/sessions/:id/unregister

Delete session and credentials.

**Response** `204 No Content`

### POST /api/internal/sessions/:id/history-sync

Configure history sync settings.

**Request**

```json
{
  "enabled": true,
  "full_sync": false,
  "since": "2026-01-01T00:00:00Z"
}
```

---

## Messages (Write Role)

### POST /api/messages

Send a WhatsApp message.

**Request - Text Message**

```json
{
  "session_id": "session-123",
  "to": "1234567890@s.whatsapp.net",
  "type": "text",
  "content": {
    "text": "Hello, World!"
  }
}
```

**Request - Image Message**

```json
{
  "session_id": "session-123",
  "to": "1234567890@s.whatsapp.net",
  "type": "image",
  "content": {
    "url": "https://example.com/image.jpg",
    "caption": "Check this out!"
  }
}
```

**Request - Document Message**

```json
{
  "session_id": "session-123",
  "to": "1234567890@s.whatsapp.net",
  "type": "document",
  "content": {
    "url": "https://example.com/document.pdf",
    "filename": "report.pdf"
  }
}
```

**Response** `202 Accepted`

```json
{
  "id": "msg-uuid-123",
  "session_id": "session-123",
  "to": "1234567890@s.whatsapp.net",
  "status": "queued",
  "timestamp": "2026-02-03T13:30:00Z"
}
```

### POST /api/messages/:messageId/reactions

Send a reaction to a message.

**Request**

```json
{
  "session_id": "session-123",
  "chat_jid": "1234567890@s.whatsapp.net",
  "emoji": "👍"
}
```

**Response** `200 OK`

```json
{
  "message_id": "msg-123",
  "emoji": "👍",
  "sent_at": "2026-02-03T13:30:00Z"
}
```

### DELETE /api/messages/:messageId/reactions

Remove a reaction from a message.

**Request**

```json
{
  "session_id": "session-123",
  "chat_jid": "1234567890@s.whatsapp.net"
}
```

**Response** `200 OK`

### POST /api/messages/receipts

Send read receipts.

**Request**

```json
{
  "session_id": "session-123",
  "chat_jid": "1234567890@s.whatsapp.net",
  "message_ids": ["msg-1", "msg-2", "msg-3"]
}
```

**Response** `200 OK`

---

## Presence (Write Role)

### POST /api/presence

Send presence update (typing indicator).

**Request**

```json
{
  "session_id": "session-123",
  "chat_jid": "1234567890@s.whatsapp.net",
  "state": "typing"
}
```

**States**: `typing`, `paused`, `online`, `offline`

**Response** `200 OK`

---

## Contacts (Read Role)

### GET /api/contacts/check

Check if phone number is on WhatsApp.

**Query Parameters**

- `session_id` (required): Session to use
- `phone` (required): Phone number (E.164 format)

**Response** `200 OK`

```json
{
  "phone": "+1234567890",
  "on_whatsapp": true,
  "jid": "1234567890@s.whatsapp.net"
}
```

### GET /api/contacts/:jid/profile

Get user profile information.

**Query Parameters**

- `session_id` (required): Session to use

**Response** `200 OK`

```json
{
  "jid": "1234567890@s.whatsapp.net",
  "name": "John Doe",
  "push_name": "Johnny",
  "status": "Available",
  "picture_url": "https://..."
}
```

### GET /api/sessions/:id/contacts

List all contacts for a session.

**Response** `200 OK`

```json
{
  "contacts": [
    {
      "jid": "1234567890@s.whatsapp.net",
      "name": "John Doe",
      "push_name": "Johnny"
    }
  ]
}
```

### GET /api/sessions/:id/chats

List all chats for a session.

**Response** `200 OK`

```json
{
  "chats": [
    {
      "jid": "1234567890@s.whatsapp.net",
      "name": "John Doe",
      "last_message_at": "2026-02-03T13:00:00Z",
      "unread_count": 5
    }
  ]
}
```

---

## Groups (Write Role)

### POST /api/sessions/:id/groups/sync

Sync all groups from WhatsApp.

**Response** `200 OK`

```json
{
  "groups": [
    {
      "jid": "123456789-1234567890@g.us",
      "name": "Family Group",
      "owner_jid": "1234567890@s.whatsapp.net",
      "participant_count": 15
    }
  ]
}
```

---

## WebSocket Endpoints

### WS /ws/qr/:sessionId

QR code authentication stream.

**Events Received**

```json
{"type": "qr", "data": "base64-encoded-qr-image"}
{"type": "authenticated", "data": "1234567890@s.whatsapp.net"}
{"type": "error", "message": "QR timeout"}
{"type": "timeout"}
```

### WS /ws/events

Real-time event stream.

**Event Types**

```json
{"type": "message.received", "payload": {...}}
{"type": "message.sent", "payload": {...}}
{"type": "message.delivered", "payload": {...}}
{"type": "message.read", "payload": {...}}
{"type": "message.reaction", "payload": {...}}
{"type": "presence.update", "payload": {...}}
{"type": "session.connected", "payload": {...}}
{"type": "session.disconnected", "payload": {...}}
```

---

## Error Responses

All errors follow this format:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable message",
    "details": {}
  }
}
```

### Error Codes

| Code                    | HTTP Status | Description                   |
| ----------------------- | ----------- | ----------------------------- |
| `SESSION_NOT_FOUND`     | 404         | Session doesn't exist         |
| `SESSION_NOT_CONNECTED` | 400         | Session is disconnected       |
| `INVALID_REQUEST`       | 400         | Malformed request body        |
| `VALIDATION_ERROR`      | 400         | Field validation failed       |
| `UNAUTHORIZED`          | 401         | Missing or invalid API key    |
| `FORBIDDEN`             | 403         | Insufficient role permissions |
| `RATE_LIMITED`          | 429         | Too many requests             |
| `INTERNAL_ERROR`        | 500         | Server error                  |

---

## Rate Limiting

Default limits:

- **10 requests/second** per IP
- **20 burst** capacity

Rate limit headers:

```
X-RateLimit-Limit: 10
X-RateLimit-Remaining: 8
X-RateLimit-Reset: 1706965800
```
