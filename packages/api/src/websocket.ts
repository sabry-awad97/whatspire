/**
 * WebSocket Manager with type-safe event handling
 * Provides real-time communication for QR codes and events
 */

// ============================================================================
// Event Types
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
// WebSocket Manager Configuration
// ============================================================================

export interface WebSocketConfig {
  url: string;
  maxReconnectAttempts?: number;
  reconnectDelay?: number;
  reconnectBackoffMultiplier?: number;
  maxReconnectDelay?: number;
}

// ============================================================================
// WebSocket Manager Class
// ============================================================================

export class WebSocketManager<T = unknown> {
  private ws: WebSocket | null = null;
  private url: string;
  private reconnectAttempts = 0;
  private maxReconnectAttempts: number;
  private reconnectDelay: number;
  private reconnectBackoffMultiplier: number;
  private maxReconnectDelay: number;
  private shouldReconnect = true;
  private listeners: Set<(event: T) => void> = new Set();
  private onOpenCallbacks: Set<() => void> = new Set();
  private onCloseCallbacks: Set<() => void> = new Set();
  private onErrorCallbacks: Set<(error: Event) => void> = new Set();
  private reconnectTimeoutId: ReturnType<typeof setTimeout> | null = null;

  constructor(config: WebSocketConfig) {
    this.url = config.url;
    this.maxReconnectAttempts = config.maxReconnectAttempts ?? 5;
    this.reconnectDelay = config.reconnectDelay ?? 1000;
    this.reconnectBackoffMultiplier = config.reconnectBackoffMultiplier ?? 2;
    this.maxReconnectDelay = config.maxReconnectDelay ?? 30000;
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
        console.log(`[WebSocket] Connected: ${this.url}`);
        this.reconnectAttempts = 0;
        this.onOpenCallbacks.forEach((cb) => cb());
      };

      this.ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data) as T;
          this.listeners.forEach((listener) => listener(data));
        } catch (error) {
          console.error("[WebSocket] Failed to parse message:", error);
        }
      };

      this.ws.onerror = (error) => {
        console.error("[WebSocket] Error:", error);
        this.onErrorCallbacks.forEach((cb) => cb(error));
      };

      this.ws.onclose = (event) => {
        console.log(`[WebSocket] Closed: ${this.url} (code: ${event.code})`);
        this.onCloseCallbacks.forEach((cb) => cb());

        if (
          this.shouldReconnect &&
          this.reconnectAttempts < this.maxReconnectAttempts
        ) {
          this.scheduleReconnect();
        }
      };
    } catch (error) {
      console.error("[WebSocket] Failed to create connection:", error);
    }
  }

  /**
   * Schedule reconnection with exponential backoff
   */
  private scheduleReconnect(): void {
    this.reconnectAttempts++;
    const delay = Math.min(
      this.reconnectDelay *
        Math.pow(this.reconnectBackoffMultiplier, this.reconnectAttempts - 1),
      this.maxReconnectDelay,
    );

    console.log(
      `[WebSocket] Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})`,
    );

    this.reconnectTimeoutId = setTimeout(() => {
      this.connect();
    }, delay);
  }

  /**
   * Disconnect from WebSocket
   */
  disconnect(): void {
    this.shouldReconnect = false;

    if (this.reconnectTimeoutId) {
      clearTimeout(this.reconnectTimeoutId);
      this.reconnectTimeoutId = null;
    }

    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }

    console.log(`[WebSocket] Disconnected: ${this.url}`);
  }

  /**
   * Send message through WebSocket
   */
  send(data: unknown): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data));
    } else {
      console.warn("[WebSocket] Cannot send - not connected");
    }
  }

  /**
   * Subscribe to WebSocket messages
   * @returns Unsubscribe function
   */
  subscribe(listener: (event: T) => void): () => void {
    this.listeners.add(listener);
    return () => this.listeners.delete(listener);
  }

  /**
   * Subscribe to connection open event
   * @returns Unsubscribe function
   */
  onOpen(callback: () => void): () => void {
    this.onOpenCallbacks.add(callback);
    return () => this.onOpenCallbacks.delete(callback);
  }

  /**
   * Subscribe to connection close event
   * @returns Unsubscribe function
   */
  onClose(callback: () => void): () => void {
    this.onCloseCallbacks.add(callback);
    return () => this.onCloseCallbacks.delete(callback);
  }

  /**
   * Subscribe to error event
   * @returns Unsubscribe function
   */
  onError(callback: (error: Event) => void): () => void {
    this.onErrorCallbacks.add(callback);
    return () => this.onErrorCallbacks.delete(callback);
  }

  /**
   * Check if WebSocket is connected
   */
  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }

  /**
   * Get WebSocket ready state
   */
  getReadyState(): number | undefined {
    return this.ws?.readyState;
  }

  /**
   * Get current reconnect attempt count
   */
  getReconnectAttempts(): number {
    return this.reconnectAttempts;
  }

  /**
   * Reset reconnect attempts counter
   */
  resetReconnectAttempts(): void {
    this.reconnectAttempts = 0;
  }
}

// ============================================================================
// Factory Functions
// ============================================================================

export interface WebSocketFactoryConfig {
  baseURL?: string;
  maxReconnectAttempts?: number;
  reconnectDelay?: number;
}

/**
 * Create QR WebSocket for session authentication
 * @param sessionId - Session identifier
 * @param config - WebSocket configuration
 * @returns WebSocket manager for QR events
 */
export function createQRWebSocket(
  sessionId: string,
  config: WebSocketFactoryConfig = {},
): WebSocketManager<QRWebSocketEvent> {
  const baseURL = config.baseURL || "ws://localhost:8080";
  return new WebSocketManager<QRWebSocketEvent>({
    url: `${baseURL}/ws/qr/${sessionId}`,
    maxReconnectAttempts: config.maxReconnectAttempts,
    reconnectDelay: config.reconnectDelay,
  });
}

/**
 * Create Event WebSocket for real-time events
 * @param config - WebSocket configuration
 * @returns WebSocket manager for event stream
 */
export function createEventWebSocket(
  config: WebSocketFactoryConfig = {},
): WebSocketManager<EventWebSocketEvent> {
  const baseURL = config.baseURL || "ws://localhost:8080";
  return new WebSocketManager<EventWebSocketEvent>({
    url: `${baseURL}/ws/events`,
    maxReconnectAttempts: config.maxReconnectAttempts,
    reconnectDelay: config.reconnectDelay,
  });
}
