# useWebSocket Hook

Reusable React hook for managing WebSocket connections with automatic lifecycle management, reconnection logic, and type-safe event handling.

## Features

- ✅ **Automatic Lifecycle Management** - Connects on mount, disconnects on unmount
- ✅ **Reconnection with Exponential Backoff** - Automatically reconnects with configurable retry logic
- ✅ **Type-Safe Event Handling** - Full TypeScript support for message types
- ✅ **React-Friendly State** - Built-in state management for connection status
- ✅ **Flexible API** - Use callbacks or state-based approach
- ✅ **Specialized Hooks** - Pre-configured hooks for QR and Events WebSockets

## Installation

The hook is part of the `@whatspire/hooks` package:

```tsx
import {
  useWebSocket,
  useQRWebSocket,
  useEventWebSocket,
} from "@whatspire/hooks";
```

## Basic Usage

### Generic WebSocket Hook

```tsx
import { useWebSocket } from "@whatspire/hooks";

function MyComponent() {
  const { isConnected, lastMessage, send } = useWebSocket<MyEventType>(
    "ws://localhost:8080/ws/my-endpoint",
    {
      onMessage: (event) => {
        console.log("Received:", event);
      },
      onOpen: () => console.log("Connected"),
      onClose: () => console.log("Disconnected"),
      onError: (error) => console.error("Error:", error),
    },
  );

  return (
    <div>
      <p>Status: {isConnected ? "Connected" : "Disconnected"}</p>
      <button onClick={() => send({ type: "ping" })}>Send Ping</button>
    </div>
  );
}
```

### QR WebSocket Hook

Specialized hook for session authentication via QR code:

```tsx
import { useQRWebSocket } from "@whatspire/hooks";

function QRCodeDisplay({ sessionId }: { sessionId: string }) {
  const [qrCode, setQrCode] = useState("");
  const [status, setStatus] = useState<"loading" | "ready" | "authenticated">(
    "loading",
  );

  const { isConnected, connect } = useQRWebSocket(sessionId, {
    onMessage: (event) => {
      switch (event.type) {
        case "qr":
          setQrCode(event.data);
          setStatus("ready");
          break;
        case "authenticated":
          setStatus("authenticated");
          console.log("Authenticated with JID:", event.data);
          break;
        case "error":
          console.error("Auth error:", event.message);
          break;
        case "timeout":
          console.log("QR code expired");
          break;
      }
    },
  });

  return (
    <div>
      {status === "loading" && <p>Loading QR code...</p>}
      {status === "ready" && <img src={qrCode} alt="QR Code" />}
      {status === "authenticated" && <p>✓ Authenticated!</p>}
    </div>
  );
}
```

### Event WebSocket Hook

Specialized hook for real-time event streaming:

```tsx
import { useEventWebSocket } from "@whatspire/hooks";

function EventStream() {
  const [events, setEvents] = useState<any[]>([]);

  const { isConnected, lastMessage } = useEventWebSocket({
    onMessage: (event) => {
      switch (event.type) {
        case "message.received":
          console.log("New message:", event.payload);
          setEvents((prev) => [...prev, event]);
          break;
        case "presence.update":
          console.log("Presence update:", event.payload);
          break;
        case "session.connected":
          console.log("Session connected:", event.payload);
          break;
        case "session.disconnected":
          console.log("Session disconnected:", event.payload);
          break;
      }
    },
  });

  return (
    <div>
      <p>Status: {isConnected ? "🟢 Live" : "🔴 Offline"}</p>
      <ul>
        {events.map((event, i) => (
          <li key={i}>{event.type}</li>
        ))}
      </ul>
    </div>
  );
}
```

## API Reference

### `useWebSocket<T>(url, options)`

Main WebSocket hook with full control over connection lifecycle.

#### Parameters

- **`url`** (string, required) - WebSocket URL to connect to
- **`options`** (object, optional) - Configuration options

#### Options

| Option                       | Type                   | Default | Description                         |
| ---------------------------- | ---------------------- | ------- | ----------------------------------- |
| `autoConnect`                | boolean                | `true`  | Automatically connect on mount      |
| `enabled`                    | boolean                | `true`  | Whether the hook is enabled         |
| `onOpen`                     | () => void             | -       | Callback when connection opens      |
| `onClose`                    | () => void             | -       | Callback when connection closes     |
| `onError`                    | (error: Event) => void | -       | Callback when an error occurs       |
| `onMessage`                  | (event: T) => void     | -       | Callback when a message is received |
| `maxReconnectAttempts`       | number                 | `5`     | Maximum reconnection attempts       |
| `reconnectDelay`             | number                 | `1000`  | Initial reconnection delay (ms)     |
| `reconnectBackoffMultiplier` | number                 | `2`     | Backoff multiplier for reconnection |
| `maxReconnectDelay`          | number                 | `30000` | Maximum reconnection delay (ms)     |

#### Return Value

| Property            | Type                    | Description                              |
| ------------------- | ----------------------- | ---------------------------------------- |
| `send`              | (data: unknown) => void | Send a message through WebSocket         |
| `connect`           | () => void              | Manually connect to WebSocket            |
| `disconnect`        | () => void              | Manually disconnect from WebSocket       |
| `isConnected`       | boolean                 | Whether WebSocket is currently connected |
| `readyState`        | number \| undefined     | Current WebSocket ready state            |
| `reconnectAttempts` | number                  | Current reconnect attempt count          |
| `lastMessage`       | T \| null               | Latest message received                  |

### `useQRWebSocket(sessionId, options)`

Specialized hook for QR code authentication.

#### Parameters

- **`sessionId`** (string, required) - Session ID to authenticate
- **`options`** (object, optional) - Same as `useWebSocket` options plus:
  - `baseURL` (string, default: `"ws://localhost:8080"`) - Base WebSocket URL

#### Event Types

```typescript
type QRWebSocketEvent =
  | { type: "qr"; data: string } // Base64 QR code image
  | { type: "authenticated"; data: string } // JID after successful auth
  | { type: "error"; message: string } // Error message
  | { type: "timeout" }; // QR code expired
```

### `useEventWebSocket(options)`

Specialized hook for real-time event streaming.

#### Parameters

- **`options`** (object, optional) - Same as `useWebSocket` options plus:
  - `baseURL` (string, default: `"ws://localhost:8080"`) - Base WebSocket URL

#### Event Types

```typescript
type EventWebSocketEvent =
  | { type: "message.received"; payload: unknown }
  | { type: "message.sent"; payload: unknown }
  | { type: "message.delivered"; payload: unknown }
  | { type: "message.read"; payload: unknown }
  | { type: "message.reaction"; payload: unknown }
  | { type: "presence.update"; payload: unknown }
  | { type: "session.connected"; payload: unknown }
  | { type: "session.disconnected"; payload: unknown };
```

## Advanced Usage

### Conditional Connection

```tsx
const { connect, disconnect, isConnected } = useWebSocket(url, {
  autoConnect: false,
  enabled: isReady,
});

// Manually control connection
useEffect(() => {
  if (shouldConnect) {
    connect();
  } else {
    disconnect();
  }
}, [shouldConnect]);
```

### State-Based Approach

```tsx
const { lastMessage, isConnected } = useWebSocket<MyEvent>(url);

// Use lastMessage in effects
useEffect(() => {
  if (lastMessage?.type === "update") {
    handleUpdate(lastMessage.data);
  }
}, [lastMessage]);
```

### Custom Reconnection Logic

```tsx
const { isConnected, reconnectAttempts } = useWebSocket(url, {
  maxReconnectAttempts: 10,
  reconnectDelay: 2000,
  reconnectBackoffMultiplier: 1.5,
  maxReconnectDelay: 60000,
});

// Show reconnection status
if (!isConnected && reconnectAttempts > 0) {
  return <p>Reconnecting... (attempt {reconnectAttempts})</p>;
}
```

### Sending Messages

```tsx
const { send, isConnected } = useWebSocket(url);

const handleSendMessage = () => {
  if (isConnected) {
    send({
      type: "chat.message",
      content: "Hello!",
      timestamp: Date.now(),
    });
  }
};
```

## Best Practices

1. **Use Specialized Hooks** - Prefer `useQRWebSocket` and `useEventWebSocket` for their specific use cases
2. **Handle All Event Types** - Always handle all possible event types in your `onMessage` callback
3. **Cleanup** - The hook automatically cleans up on unmount, no manual cleanup needed
4. **Error Handling** - Always provide an `onError` callback for production apps
5. **Type Safety** - Define your event types for full TypeScript support
6. **Conditional Enabling** - Use the `enabled` option to conditionally enable/disable the hook

## Migration from Direct WebSocketManager

### Before (using WebSocketManager directly)

```tsx
useEffect(() => {
  const ws = createQRWebSocket(sessionId);
  ws.connect();

  const unsubscribe = ws.subscribe((event) => {
    handleEvent(event);
  });

  return () => {
    unsubscribe();
    ws.disconnect();
  };
}, [sessionId]);
```

### After (using useQRWebSocket hook)

```tsx
useQRWebSocket(sessionId, {
  onMessage: (event) => {
    handleEvent(event);
  },
});
```

## Troubleshooting

### WebSocket not connecting

- Check that the `url` is correct and the server is running
- Verify the `enabled` option is `true`
- Check browser console for WebSocket errors

### Messages not received

- Ensure `onMessage` callback is provided or check `lastMessage` state
- Verify the WebSocket is connected (`isConnected === true`)
- Check that message format matches your type definition

### Reconnection not working

- Check `maxReconnectAttempts` is not set too low
- Verify server is accepting reconnections
- Monitor `reconnectAttempts` to see if retries are happening

## Examples

See the following files for complete examples:

- `apps/web/src/components/sessions/qr-code-display.tsx` - QR code authentication
- `apps/web/src/routes/index.tsx` - Dashboard with event streaming

## Related

- [WebSocketManager](../../api/src/websocket.ts) - Underlying WebSocket manager class
- [API Client](../../api/README.md) - REST API client
- [React Query Hooks](./README.md) - Other hooks in this package
