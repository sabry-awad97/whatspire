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
 * Mutation options for reconnecting a session
 * @param client - API client instance
 * @param queryClient - Query client for cache updates
 * @returns Mutation options for useMutation
 */
export const reconnectSessionMutation = (
  client: ApiClient,
  queryClient: QueryClient,
): MutationOptions<Session, ApiClientError, string> => ({
  mutationFn: (sessionId) => client.reconnectSession(sessionId),
  onSuccess: (_, sessionId) => {
    // Invalidate sessions list and detail to trigger refetch
    queryClient.invalidateQueries({ queryKey: sessionKeys.lists() });
    queryClient.invalidateQueries({ queryKey: sessionKeys.detail(sessionId) });
  },
  onError: (error) => {
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
): MutationOptions<Session, ApiClientError, string> => ({
  mutationFn: (sessionId) => client.disconnectSession(sessionId),
  onSuccess: (_, sessionId) => {
    // Invalidate sessions list and detail to trigger refetch
    queryClient.invalidateQueries({ queryKey: sessionKeys.lists() });
    queryClient.invalidateQueries({ queryKey: sessionKeys.detail(sessionId) });
  },
  onError: (error) => {
    console.error("Failed to disconnect session:", error);
  },
});
