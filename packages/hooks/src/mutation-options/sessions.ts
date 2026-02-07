/**
 * Session Mutation Options
 * Centralized mutation configurations for session-related operations
 */
import { type MutationOptions, type QueryClient } from "@tanstack/react-query";
import { ApiClient, ApiClientError } from "@whatspire/api";
import type { Session, CreateSessionRequest } from "@whatspire/schema";
import { sessionKeys } from "../query-options/sessions";

// ============================================================================
// Mutation Options Factories
// ============================================================================

/**
 * Mutation options for creating a session
 * @param client - API client instance
 * @param queryClient - Query client for cache updates
 * @returns Mutation options for useMutation
 */
export const createSessionMutation = (
  client: ApiClient,
  queryClient: QueryClient,
): MutationOptions<Session, ApiClientError, CreateSessionRequest> => ({
  mutationFn: (data) => client.createSession(data),
  onSuccess: () => {
    // Invalidate sessions list to trigger refetch
    queryClient.invalidateQueries({ queryKey: sessionKeys.lists() });
  },
  onError: (error) => {
    console.error("Failed to create session:", error);
  },
});

/**
 * Mutation options for deleting a session
 * @param client - API client instance
 * @param queryClient - Query client for cache updates
 * @returns Mutation options for useMutation
 */
export const deleteSessionMutation = (
  client: ApiClient,
  queryClient: QueryClient,
): MutationOptions<void, ApiClientError, string> => ({
  mutationFn: (sessionId) => client.deleteSession(sessionId),
  onSuccess: (_, sessionId) => {
    // Invalidate sessions list to trigger refetch
    queryClient.invalidateQueries({ queryKey: sessionKeys.lists() });

    // Remove session detail cache
    queryClient.removeQueries({ queryKey: sessionKeys.detail(sessionId) });
  },
  onError: (error) => {
    console.error("Failed to delete session:", error);
  },
});

/**
 * Mutation options for updating a session
 * @param client - API client instance
 * @param queryClient - Query client for cache updates
 * @returns Mutation options for useMutation
 */
export const updateSessionMutation = (
  client: ApiClient,
  queryClient: QueryClient,
): MutationOptions<
  Session,
  ApiClientError,
  {
    sessionId: string;
    data: {
      name?: string;
      config?: {
        account_protection?: boolean;
        message_logging?: boolean;
        read_messages?: boolean;
        auto_reject_calls?: boolean;
        always_online?: boolean;
        ignore_groups?: boolean;
        ignore_broadcasts?: boolean;
        ignore_channels?: boolean;
      };
    };
  }
> => ({
  mutationFn: ({ sessionId, data }) => client.updateSession(sessionId, data),
  onSuccess: (session) => {
    // Update session detail cache
    queryClient.setQueryData(sessionKeys.detail(session.id), session);

    // Invalidate sessions list to trigger refetch
    queryClient.invalidateQueries({ queryKey: sessionKeys.lists() });
  },
  onError: (error) => {
    console.error("Failed to update session:", error);
  },
});

/**
 * Mutation options for reconnecting a session
 * @param client - API client instance
 * @param queryClient - Query client for cache updates
 * @returns Mutation options for useMutation
 */
export const reconnectSessionMutation = (
  client: ApiClient,
  queryClient: QueryClient,
): MutationOptions<
  void,
  ApiClientError,
  string,
  { previousSessions?: Session[] }
> => ({
  mutationFn: (sessionId) => client.reconnectSession(sessionId),
  onMutate: async (sessionId) => {
    // Cancel outgoing refetches
    await queryClient.cancelQueries({ queryKey: sessionKeys.lists() });

    // Snapshot previous value
    const previousSessions = queryClient.getQueryData<Session[]>(
      sessionKeys.lists(),
    );

    // Optimistically update to "connecting" status
    if (previousSessions) {
      queryClient.setQueryData<Session[]>(sessionKeys.lists(), (old) =>
        old?.map((session) =>
          session.id === sessionId
            ? { ...session, status: "connecting" as const }
            : session,
        ),
      );
    }

    return { previousSessions };
  },
  onSuccess: async () => {
    // Real-time WebSocket will update the status automatically
    // No need to invalidate or wait
  },
  onError: (error, _sessionId, context) => {
    // Rollback on error
    if (context?.previousSessions) {
      queryClient.setQueryData(sessionKeys.lists(), context.previousSessions);
    }
    console.error("Failed to reconnect session:", error);
  },
});

/**
 * Mutation options for disconnecting a session
 * @param client - API client instance
 * @param queryClient - Query client for cache updates
 * @returns Mutation options for useMutation
 */
export const disconnectSessionMutation = (
  client: ApiClient,
  queryClient: QueryClient,
): MutationOptions<
  void,
  ApiClientError,
  string,
  { previousSessions?: Session[] }
> => ({
  mutationFn: (sessionId) => client.disconnectSession(sessionId),
  onMutate: async (sessionId) => {
    // Cancel outgoing refetches
    await queryClient.cancelQueries({ queryKey: sessionKeys.lists() });

    // Snapshot previous value
    const previousSessions = queryClient.getQueryData<Session[]>(
      sessionKeys.lists(),
    );

    // Optimistically update to "disconnected" status
    if (previousSessions) {
      queryClient.setQueryData<Session[]>(sessionKeys.lists(), (old) =>
        old?.map((session) =>
          session.id === sessionId
            ? { ...session, status: "disconnected" as const }
            : session,
        ),
      );
    }

    return { previousSessions };
  },
  onSuccess: async () => {
    // Real-time WebSocket will update the status automatically
    // No need to invalidate or wait
  },
  onError: (error, _sessionId, context) => {
    // Rollback on error
    if (context?.previousSessions) {
      queryClient.setQueryData(sessionKeys.lists(), context.previousSessions);
    }
    console.error("Failed to disconnect session:", error);
  },
});
