/**
 * @whatspire/api
 *
 * Type-safe API client with schema validation
 * Built on @whatspire/schema for runtime validation and type inference
 */

// API Client
export { ApiClient, ApiClientError } from "./client";
export type { ApiClientConfig, RetryConfig } from "./client";

// WebSocket
export {
  WebSocketManager,
  createQRWebSocket,
  createEventWebSocket,
} from "./websocket";
export type {
  WebSocketConfig,
  WebSocketFactoryConfig,
  QRWebSocketEvent,
  QREvent,
  AuthenticatedEvent,
  ErrorEvent,
  TimeoutEvent,
  EventWebSocketEvent,
  MessageReceivedEvent,
  MessageSentEvent,
  MessageDeliveredEvent,
  MessageReadEvent,
  MessageReactionEvent,
  PresenceUpdateEvent,
  SessionConnectedEvent,
  SessionDisconnectedEvent,
} from "./websocket";
