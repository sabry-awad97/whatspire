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
  onSuccess: (newSession) => {
    // Update the sessions list cache
    queryClient.setQueryData<Session[]>(sessionKeys.lists(), (old) => {
      if (!old) return [newSession];
      return [...old, newSession];
    });

    // Set the new session detail cache
    queryClient.setQueryData(sessionKeys.detail(newSession.id), newSession);
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
    // Remove from sessions list cache
    queryClient.setQueryData<Session[]>(sessionKeys.lists(), (old) => {
      if (!old) return [];
      return old.filter((session) => session.id !== sessionId);
    });

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
  onSuccess: (updatedSession) => {
    // Update sessions list cache
    queryClient.setQueryData<Session[]>(sessionKeys.lists(), (old) => {
      if (!old) return [updatedSession];
      return old.map((session) =>
        session.id === updatedSession.id ? updatedSession : session,
      );
    });

    // Update session detail cache
    queryClient.setQueryData(
      sessionKeys.detail(updatedSession.id),
      updatedSession,
    );
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
  onSuccess: (updatedSession) => {
    // Update sessions list cache
    queryClient.setQueryData<Session[]>(sessionKeys.lists(), (old) => {
      if (!old) return [updatedSession];
      return old.map((session) =>
        session.id === updatedSession.id ? updatedSession : session,
      );
    });

    // Update session detail cache
    queryClient.setQueryData(
      sessionKeys.detail(updatedSession.id),
      updatedSession,
    );
  },
  onError: (error) => {
    console.error("Failed to disconnect session:", error);
  },
});
