// ============================================================================
// WebSocket Event Types
// ============================================================================

export interface QREvent {
  type: "qr";
  data: string; // base64 encoded QR image
}

export interface AuthenticatedEvent {
  type: "authenticated";
  data: string; // JID
}

export interface ErrorEvent {
  type: "error";
  message: string;
}

export interface TimeoutEvent {
  type: "timeout";
}

export type QRWebSocketEvent =
  | QREvent
  | AuthenticatedEvent
  | ErrorEvent
  | TimeoutEvent;

export interface MessageReceivedEvent {
  type: "message.received";
  payload: unknown;
}

export interface MessageSentEvent {
  type: "message.sent";
  payload: unknown;
}

export interface MessageDeliveredEvent {
  type: "message.delivered";
  payload: unknown;
}

export interface MessageReadEvent {
  type: "message.read";
  payload: unknown;
}

export interface MessageReactionEvent {
  type: "message.reaction";
  payload: unknown;
}

export interface PresenceUpdateEvent {
  type: "presence.update";
  payload: unknown;
}

export interface SessionConnectedEvent {
  type: "session.connected";
  payload: unknown;
}

export interface SessionDisconnectedEvent {
  type: "session.disconnected";
  payload: unknown;
}

export type EventWebSocketEvent =
  | MessageReceivedEvent
  | MessageSentEvent
  | MessageDeliveredEvent
  | MessageReadEvent
  | MessageReactionEvent
  | PresenceUpdateEvent
  | SessionConnectedEvent
  | SessionDisconnectedEvent;

// ============================================================================
// WebSocket Manager
// ============================================================================

export class WebSocketManager<T = unknown> {
  private ws: WebSocket | null = null;
  private url: string;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  private listeners: Set<(event: T) => void> = new Set();
  private onOpenCallbacks: Set<() => void> = new Set();
  private onCloseCallbacks: Set<() => void> = new Set();
  private onErrorCallbacks: Set<(error: Event) => void> = new Set();
  private shouldReconnect = true;

  constructor(url: string) {
    this.url = url;
  }

  /**
   * Connect to WebSocket
   */
  connect(): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      return;
    }

    try {
      this.ws = new WebSocket(this.url);

      this.ws.onopen = () => {
        console.log(`WebSocket connected: ${this.url}`);
        this.reconnectAttempts = 0;
        this.onOpenCallbacks.forEach((cb) => cb());
      };

      this.ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data) as T;
          this.listeners.forEach((listener) => listener(data));
        } catch (error) {
          console.error("Failed to parse WebSocket message:", error);
        }
      };

      this.ws.onerror = (error) => {
        console.error("WebSocket error:", error);
        this.onErrorCallbacks.forEach((cb) => cb(error));
      };

      this.ws.onclose = () => {
        console.log(`WebSocket closed: ${this.url}`);
        this.onCloseCallbacks.forEach((cb) => cb());

        if (
          this.shouldReconnect &&
          this.reconnectAttempts < this.maxReconnectAttempts
        ) {
          this.reconnectAttempts++;
          const delay =
            this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);
          console.log(
            `Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts})`,
          );
          setTimeout(() => this.connect(), delay);
        }
      };
    } catch (error) {
      console.error("Failed to create WebSocket:", error);
    }
  }

  /**
   * Disconnect from WebSocket
   */
  disconnect(): void {
    this.shouldReconnect = false;
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  /**
   * Send message through WebSocket
   */
  send(data: unknown): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data));
    } else {
      console.warn("WebSocket is not connected");
    }
  }

  /**
   * Subscribe to WebSocket messages
   */
  subscribe(listener: (event: T) => void): () => void {
    this.listeners.add(listener);
    return () => this.listeners.delete(listener);
  }

  /**
   * Subscribe to connection open event
   */
  onOpen(callback: () => void): () => void {
    this.onOpenCallbacks.add(callback);
    return () => this.onOpenCallbacks.delete(callback);
  }

  /**
   * Subscribe to connection close event
   */
  onClose(callback: () => void): () => void {
    this.onCloseCallbacks.add(callback);
    return () => this.onCloseCallbacks.delete(callback);
  }

  /**
   * Subscribe to error event
   */
  onError(callback: (error: Event) => void): () => void {
    this.onErrorCallbacks.add(callback);
    return () => this.onErrorCallbacks.delete(callback);
  }

  /**
   * Get connection state
   */
  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }

  /**
   * Get connection state
   */
  getReadyState(): number | undefined {
    return this.ws?.readyState;
  }
}

// ============================================================================
// Factory Functions
// ============================================================================

const WS_BASE_URL =
  import.meta.env.VITE_SERVER_URL?.replace("http", "ws") ||
  "ws://localhost:8080";

export function createQRWebSocket(
  sessionId: string,
): WebSocketManager<QRWebSocketEvent> {
  return new WebSocketManager<QRWebSocketEvent>(
    `${WS_BASE_URL}/ws/qr/${sessionId}`,
  );
}

export function createEventWebSocket(): WebSocketManager<EventWebSocketEvent> {
  return new WebSocketManager<EventWebSocketEvent>(`${WS_BASE_URL}/ws/events`);
}
