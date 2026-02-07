/**
 * WebSocket Hook
 * Reusable React hook for WebSocket connections with automatic lifecycle management
 */
import { useEffect, useRef, useCallback, useState } from "react";
import { WebSocketManager, type WebSocketConfig } from "@whatspire/api";

// ============================================================================
// Types
// ============================================================================

export interface UseWebSocketOptions<T> extends Omit<WebSocketConfig, "url"> {
  /**
   * Whether to automatically connect on mount
   * @default true
   */
  autoConnect?: boolean;

  /**
   * Callback when connection opens
   */
  onOpen?: () => void;

  /**
   * Callback when connection closes
   */
  onClose?: () => void;

  /**
   * Callback when an error occurs
   */
  onError?: (error: Event) => void;

  /**
   * Callback when a message is received
   */
  onMessage?: (event: T) => void;

  /**
   * Whether the hook is enabled
   * @default true
   */
  enabled?: boolean;
}

export interface UseWebSocketReturn<T> {
  /**
   * Send a message through the WebSocket
   */
  send: (data: unknown) => void;

  /**
   * Manually connect to the WebSocket
   */
  connect: () => void;

  /**
   * Manually disconnect from the WebSocket
   */
  disconnect: () => void;

  /**
   * Whether the WebSocket is currently connected
   */
  isConnected: boolean;

  /**
   * Current WebSocket ready state
   */
  readyState: number | undefined;

  /**
   * Current reconnect attempt count
   */
  reconnectAttempts: number;

  /**
   * Latest message received (if not using onMessage callback)
   */
  lastMessage: T | null;
}

// ============================================================================
// Hook
// ============================================================================

/**
 * Hook for managing WebSocket connections with automatic lifecycle management
 *
 * Features:
 * - Automatic connection/disconnection on mount/unmount
 * - Reconnection with exponential backoff
 * - Type-safe message handling
 * - React-friendly state management
 *
 * @param url - WebSocket URL to connect to
 * @param options - Configuration options
 * @returns WebSocket control interface
 *
 * @example
 * ```tsx
 * // Basic usage with callback
 * const { isConnected, send } = useWebSocket<QRWebSocketEvent>(
 *   `ws://localhost:8080/ws/qr/${sessionId}`,
 *   {
 *     onMessage: (event) => {
 *       if (event.type === "qr") {
 *         setQrCode(event.data);
 *       }
 *     },
 *     onOpen: () => console.log("Connected"),
 *     onClose: () => console.log("Disconnected"),
 *   }
 * );
 *
 * // Usage with state
 * const { lastMessage, isConnected } = useWebSocket<EventWebSocketEvent>(
 *   "ws://localhost:8080/ws/events"
 * );
 *
 * // Conditional connection
 * const { connect, disconnect } = useWebSocket(url, {
 *   autoConnect: false,
 *   enabled: isReady,
 * });
 * ```
 */
export function useWebSocket<T = unknown>(
  url: string,
  options: UseWebSocketOptions<T> = {},
): UseWebSocketReturn<T> {
  const {
    autoConnect = true,
    enabled = true,
    onOpen,
    onClose,
    onError,
    onMessage,
    maxReconnectAttempts,
    reconnectDelay,
    reconnectBackoffMultiplier,
    maxReconnectDelay,
  } = options;

  // WebSocket manager instance (persisted across renders)
  const wsRef = useRef<WebSocketManager<T> | null>(null);

  // State for tracking connection status
  const [isConnected, setIsConnected] = useState(false);
  const [readyState, setReadyState] = useState<number | undefined>(undefined);
  const [reconnectAttempts, setReconnectAttempts] = useState(0);
  const [lastMessage, setLastMessage] = useState<T | null>(null);

  // Stable callback refs to avoid recreating WebSocket on callback changes
  const onOpenRef = useRef(onOpen);
  const onCloseRef = useRef(onClose);
  const onErrorRef = useRef(onError);
  const onMessageRef = useRef(onMessage);

  // Update callback refs when they change
  useEffect(() => {
    onOpenRef.current = onOpen;
  }, [onOpen]);

  useEffect(() => {
    onCloseRef.current = onClose;
  }, [onClose]);

  useEffect(() => {
    onErrorRef.current = onError;
  }, [onError]);

  useEffect(() => {
    onMessageRef.current = onMessage;
  }, [onMessage]);

  // Initialize WebSocket manager
  useEffect(() => {
    if (!enabled) return;

    // Create WebSocket manager
    const ws = new WebSocketManager<T>({
      url,
      maxReconnectAttempts,
      reconnectDelay,
      reconnectBackoffMultiplier,
      maxReconnectDelay,
    });

    wsRef.current = ws;

    // Subscribe to connection events
    const unsubscribeOpen = ws.onOpen(() => {
      setIsConnected(true);
      setReadyState(ws.getReadyState());
      setReconnectAttempts(ws.getReconnectAttempts());
      onOpenRef.current?.();
    });

    const unsubscribeClose = ws.onClose(() => {
      setIsConnected(false);
      setReadyState(ws.getReadyState());
      setReconnectAttempts(ws.getReconnectAttempts());
      onCloseRef.current?.();
    });

    const unsubscribeError = ws.onError((error) => {
      onErrorRef.current?.(error);
    });

    const unsubscribeMessage = ws.subscribe((event) => {
      setLastMessage(event);
      onMessageRef.current?.(event);
    });

    // Auto-connect if enabled
    if (autoConnect) {
      ws.connect();
    }

    // Cleanup on unmount
    return () => {
      unsubscribeOpen();
      unsubscribeClose();
      unsubscribeError();
      unsubscribeMessage();
      ws.disconnect();
      wsRef.current = null;
    };
  }, [
    url,
    enabled,
    autoConnect,
    maxReconnectAttempts,
    reconnectDelay,
    reconnectBackoffMultiplier,
    maxReconnectDelay,
  ]);

  // Stable send function
  const send = useCallback((data: unknown) => {
    wsRef.current?.send(data);
  }, []);

  // Stable connect function
  const connect = useCallback(() => {
    wsRef.current?.connect();
  }, []);

  // Stable disconnect function
  const disconnect = useCallback(() => {
    wsRef.current?.disconnect();
  }, []);

  return {
    send,
    connect,
    disconnect,
    isConnected,
    readyState,
    reconnectAttempts,
    lastMessage,
  };
}

// ============================================================================
// Specialized Hooks
// ============================================================================

import type { QRWebSocketEvent, EventWebSocketEvent } from "@whatspire/api";

/**
 * Hook for QR code WebSocket connection
 * Specialized hook for session authentication via QR code
 *
 * @param sessionId - Session ID to authenticate
 * @param options - Configuration options
 * @returns WebSocket control interface with QR event types
 *
 * @example
 * ```tsx
 * const { lastMessage, isConnected } = useQRWebSocket(sessionId, {
 *   onMessage: (event) => {
 *     switch (event.type) {
 *       case "qr":
 *         setQrCode(event.data);
 *         break;
 *       case "authenticated":
 *         console.log("Authenticated:", event.data);
 *         break;
 *       case "error":
 *         console.error("Error:", event.message);
 *         break;
 *       case "timeout":
 *         console.log("QR code expired");
 *         break;
 *     }
 *   },
 * });
 * ```
 */
export function useQRWebSocket(
  sessionId: string,
  options: Omit<UseWebSocketOptions<QRWebSocketEvent>, "url"> & {
    baseURL?: string;
  } = {},
): UseWebSocketReturn<QRWebSocketEvent> {
  const { baseURL = "ws://localhost:8080", ...wsOptions } = options;
  const url = `${baseURL}/ws/qr/${sessionId}`;

  return useWebSocket<QRWebSocketEvent>(url, wsOptions);
}

/**
 * Hook for Events WebSocket connection
 * Specialized hook for real-time event streaming
 *
 * @param options - Configuration options
 * @returns WebSocket control interface with event types
 *
 * @example
 * ```tsx
 * const { lastMessage, isConnected } = useEventWebSocket({
 *   onMessage: (event) => {
 *     switch (event.type) {
 *       case "message.received":
 *         console.log("New message:", event.payload);
 *         break;
 *       case "presence.update":
 *         console.log("Presence update:", event.payload);
 *         break;
 *       case "session.connected":
 *         console.log("Session connected:", event.payload);
 *         break;
 *     }
 *   },
 * });
 * ```
 */
export function useEventWebSocket(
  options: Omit<UseWebSocketOptions<EventWebSocketEvent>, "url"> & {
    baseURL?: string;
  } = {},
): UseWebSocketReturn<EventWebSocketEvent> {
  const { baseURL = "ws://localhost:8080", ...wsOptions } = options;
  const url = `${baseURL}/ws/events`;

  return useWebSocket<EventWebSocketEvent>(url, wsOptions);
}
