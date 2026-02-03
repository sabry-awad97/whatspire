# @whatspire/api

Type-safe API client for the Whatspire WhatsApp API with runtime schema validation.

## Features

- **Type-safe**: Full TypeScript support with type inference from Zod schemas
- **Runtime validation**: All requests and responses validated using `@whatspire/schema`
- **Automatic retries**: Exponential backoff for failed requests
- **Error handling**: Structured error responses with detailed information
- **WebSocket support**: Real-time QR code and event streaming
- **Configurable**: Flexible configuration for base URL, API keys, timeouts, and retry logic

## Installation

```bash
bun add @whatspire/api
```

## Quick Start

### API Client

```typescript
import { ApiClient } from "@whatspire/api";

// Create client instance
const client = new ApiClient({
  baseURL: "http://localhost:8080",
  apiKey: "your-api-key", // Optional
  timeout: 30000,
  retryConfig: {
    maxAttempts: 3,
    initialDelayMs: 1000,
  },
});

// Create a session
const session = await client.createSession({
  session_id: "my-session",
  name: "My WhatsApp",
});

// Send a message
const result = await client.sendMessage({
  session_id: "my-session",
  to: "1234567890@s.whatsapp.net",
  type: "text",
  content: {
    text: "Hello from Whatspire!",
  },
});

// List all sessions
const sessions = await client.listSessions();
```

### WebSocket Manager

```typescript
import { createQRWebSocket, createEventWebSocket } from "@whatspire/api";

// QR Code WebSocket for authentication
const qrWs = createQRWebSocket("my-session", {
  baseURL: "ws://localhost:8080",
});

qrWs.subscribe((event) => {
  if (event.type === "qr") {
    console.log("QR Code:", event.data);
  } else if (event.type === "authenticated") {
    console.log("Authenticated:", event.data);
  }
});

qrWs.connect();

// Event WebSocket for real-time updates
const eventWs = createEventWebSocket({
  baseURL: "ws://localhost:8080",
});

eventWs.subscribe((event) => {
  switch (event.type) {
    case "message.received":
      console.log("New message:", event.payload);
      break;
    case "presence.update":
      console.log("Presence update:", event.payload);
      break;
  }
});

eventWs.connect();
```

## API Reference

### ApiClient

#### Constructor

```typescript
new ApiClient(config?: ApiClientConfig)
```

**Config Options:**

- `baseURL` (string): API base URL (default: `http://localhost:8080`)
- `apiKey` (string): Optional API key for authentication
- `timeout` (number): Request timeout in milliseconds (default: `30000`)
- `retryConfig` (RetryConfig): Retry configuration

**RetryConfig:**

- `maxAttempts` (number): Maximum retry attempts (default: `3`)
- `initialDelayMs` (number): Initial delay between retries (default: `1000`)
- `maxDelayMs` (number): Maximum delay between retries (default: `10000`)
- `backoffMultiplier` (number): Backoff multiplier (default: `2`)

#### Health Endpoints

```typescript
// Check API health
await client.health();
// Returns: { status: string, timestamp: string }

// Check API readiness
await client.ready();
// Returns: { status: string, checks: Record<string, string> }
```

#### Session Management

```typescript
// Create session
await client.createSession({ session_id: "id", name: "name" });

// List sessions
await client.listSessions();

// Get session
await client.getSession("session-id");

// Delete session
await client.deleteSession("session-id");

// Reconnect session
await client.reconnectSession("session-id");

// Disconnect session
await client.disconnectSession("session-id");
```

#### Messages

```typescript
// Send message
await client.sendMessage({
  session_id: "my-session",
  to: "1234567890@s.whatsapp.net",
  type: "text",
  content: { text: "Hello!" },
});
```

#### Contacts

```typescript
// Check if phone is on WhatsApp
await client.checkPhone({
  session_id: "my-session",
  phone: "+1234567890",
});

// Get contact profile
await client.getContactProfile({
  session_id: "my-session",
  jid: "1234567890@s.whatsapp.net",
});

// Get all contacts
await client.getContacts("my-session");

// Get all chats
await client.getChats("my-session");
```

#### Groups

```typescript
// Sync groups
await client.syncGroups("my-session");
```

#### Presence

```typescript
// Send presence update
await client.sendPresence({
  session_id: "my-session",
  jid: "1234567890@s.whatsapp.net",
  state: "available",
});
```

#### Reactions

```typescript
// Send reaction
await client.sendReaction({
  session_id: "my-session",
  message_id: "msg-id",
  chat_jid: "1234567890@s.whatsapp.net",
  emoji: "👍",
});

// Remove reaction
await client.removeReaction({
  session_id: "my-session",
  message_id: "msg-id",
  chat_jid: "1234567890@s.whatsapp.net",
});
```

#### Receipts

```typescript
// Send read receipt
await client.sendReceipt({
  session_id: "my-session",
  message_ids: ["msg-id-1", "msg-id-2"],
  chat_jid: "1234567890@s.whatsapp.net",
});
```

#### Events

```typescript
// Query events
await client.queryEvents({
  session_id: "my-session",
  event_type: "message.received",
  limit: 50,
  offset: 0,
});
```

### WebSocketManager

#### Methods

```typescript
// Connect to WebSocket
ws.connect();

// Disconnect from WebSocket
ws.disconnect();

// Send message
ws.send({ type: "ping" });

// Subscribe to messages (returns unsubscribe function)
const unsubscribe = ws.subscribe((event) => {
  console.log("Event:", event);
});

// Subscribe to connection events
ws.onOpen(() => console.log("Connected"));
ws.onClose(() => console.log("Disconnected"));
ws.onError((error) => console.error("Error:", error));

// Check connection state
ws.isConnected(); // boolean
ws.getReadyState(); // WebSocket ready state
ws.getReconnectAttempts(); // number
```

#### Factory Functions

```typescript
// Create QR WebSocket
const qrWs = createQRWebSocket(sessionId, config);

// Create Event WebSocket
const eventWs = createEventWebSocket(config);
```

**Config Options:**

- `baseURL` (string): WebSocket base URL (default: `ws://localhost:8080`)
- `maxReconnectAttempts` (number): Maximum reconnection attempts (default: `5`)
- `reconnectDelay` (number): Initial reconnection delay in ms (default: `1000`)

## Error Handling

The API client throws `ApiClientError` for API errors:

```typescript
import { ApiClientError } from "@whatspire/api";

try {
  await client.createSession({ session_id: "test", name: "Test" });
} catch (error) {
  if (error instanceof ApiClientError) {
    console.error("API Error:", {
      message: error.message,
      code: error.code,
      status: error.status,
      details: error.details,
    });
  }
}
```

## Schema Validation

All requests and responses are validated using Zod schemas from `@whatspire/schema`. Invalid data will throw a `ZodError`:

```typescript
import { ZodError } from "zod";

try {
  await client.sendMessage({
    session_id: "test",
    to: "invalid-jid", // Will fail validation
    type: "text",
    content: { text: "Hello" },
  });
} catch (error) {
  if (error instanceof ZodError) {
    console.error("Validation error:", error.errors);
  }
}
```

## Type Exports

All types are re-exported from `@whatspire/schema` for convenience:

```typescript
import type {
  Session,
  CreateSessionRequest,
  SendMessageRequest,
  Contact,
  Group,
  ApiResponse,
  ApiError,
} from "@whatspire/api";
```

## Development

```bash
# Type check
bun run type-check

# Run tests
bun test
```

## License

Private package for Whatspire project.
