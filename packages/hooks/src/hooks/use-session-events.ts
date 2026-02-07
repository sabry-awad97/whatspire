/**
 * Session Events Hook
 * Listens to real-time session events via WebSocket and updates the query cache
 */
import { useEffect } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useEventWebSocket } from "./use-websocket";
import { sessionKeys } from "../query-options/sessions";
import type { EventWebSocketEvent } from "@whatspire/api";

/**
 * Type guard to check if event is a connection event with session_id
 */
function isConnectionEvent(
  event: EventWebSocketEvent,
): event is Extract<
  EventWebSocketEvent,
  { type: "connection.connected" | "connection.disconnected" }
> {
  return (
    event.type === "connection.connected" ||
    event.type === "connection.disconnected"
  );
}

/**
 * Type guard to check if event is a session event with session_id
 */
function isSessionEvent(
  event: EventWebSocketEvent,
): event is Extract<
  EventWebSocketEvent,
  { type: "session.connected" | "session.disconnected" }
> {
  return (
    event.type === "session.connected" || event.type === "session.disconnected"
  );
}

/**
 * Hook to listen for real-time session status updates
 *
 * Automatically updates the session list when:
 * - A session connects
 * - A session disconnects
 *
 * This provides instant UI updates without polling or manual refetch.
 *
 * @example
 * ```tsx
 * import { useSessionEvents, useSessions } from "@whatspire/hooks";
 *
 * function SessionList() {
 *   const { data: sessions } = useSessions(client);
 *
 *   // Enable real-time updates
 *   useSessionEvents();
 *
 *   return <div>{sessions.map(s => <SessionCard session={s} />)}</div>;
 * }
 * ```
 */
export function useSessionEvents() {
  const queryClient = useQueryClient();

  const { lastMessage, isConnected, send } = useEventWebSocket({
    onMessage: (event) => {
      // Handle authentication response
      if (event.type === "auth_response") {
        if (event.success) {
          console.log("[Real-time] Authentication successful");
        } else {
          console.error(
            "[Real-time] Authentication failed:",
            event.message || "Unknown error",
          );
        }
        return;
      }

      // Handle connection events (backend sends "connection.connected" / "connection.disconnected")
      if (isConnectionEvent(event) || isSessionEvent(event)) {
        const sessionId = event.session_id;

        if (sessionId) {
          // Invalidate queries to trigger refetch
          queryClient.invalidateQueries({
            queryKey: sessionKeys.lists(),
            refetchType: "active",
          });

          const action =
            event.type === "connection.connected" ||
            event.type === "session.connected"
              ? "connected"
              : "disconnected";

          console.log(
            `[Real-time] Session ${sessionId} ${action} - refetching`,
          );
        }
      }
    },
    onOpen: () => {
      console.log("[Real-time] Connected to event stream");
      // Send authentication message immediately after connection
      send({ type: "auth", api_key: "" });
      console.log("[Real-time] Sent authentication message");
    },
    onClose: () => {
      console.log("[Real-time] Disconnected from event stream");
    },
    onError: (error) => {
      console.error("[Real-time] WebSocket error:", error);
    },
  });

  // Log connection status changes
  useEffect(() => {
    if (isConnected) {
      console.log("[Real-time] Event stream active");
    } else {
      console.log("[Real-time] Event stream inactive");
    }
  }, [isConnected]);

  return {
    isConnected,
    lastMessage,
  };
}
