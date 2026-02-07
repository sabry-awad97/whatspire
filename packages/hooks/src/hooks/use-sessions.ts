/**
 * Session Hooks
 * Custom React hooks for session-related operations
 */
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { ApiClient } from "@whatspire/api";
import type { Session } from "@whatspire/schema";
import {
  listSessionsOptions,
  sessionDetailOptions,
} from "../query-options/sessions";
import {
  createSessionMutation,
  deleteSessionMutation,
  reconnectSessionMutation,
  disconnectSessionMutation,
} from "../mutation-options/sessions";

// ============================================================================
// Query Hooks
// ============================================================================

/**
 * Hook to fetch all sessions
 * @param client - API client instance
 * @param options - Additional query options (staleTime, enabled, etc.)
 * @returns Query result with sessions list
 *
 * @example
 * ```tsx
 * const { data: sessions, isLoading } = useSessions(client, {
 *   staleTime: 1000 * 60 * 5,
 *   enabled: true,
 * });
 * ```
 */
export function useSessions(
  client: ApiClient,
  options?: {
    staleTime?: number;
    gcTime?: number;
    enabled?: boolean;
    refetchOnWindowFocus?: boolean;
    refetchOnMount?: boolean;
    refetchOnReconnect?: boolean;
  },
) {
  return useQuery({
    ...listSessionsOptions(client),
    ...options,
  });
}

/**
 * Hook to fetch a single session
 * @param client - API client instance
 * @param sessionId - Session ID to fetch
 * @param options - Additional query options (staleTime, enabled, etc.)
 * @returns Query result with session details
 *
 * @example
 * ```tsx
 * const { data: session } = useSession(client, "session-id", {
 *   enabled: !!sessionId,
 * });
 * ```
 */
export function useSession(
  client: ApiClient,
  sessionId: string,
  options?: {
    staleTime?: number;
    gcTime?: number;
    enabled?: boolean;
    refetchOnWindowFocus?: boolean;
    refetchOnMount?: boolean;
    refetchOnReconnect?: boolean;
  },
) {
  return useQuery({
    ...sessionDetailOptions(client, sessionId),
    ...options,
  });
}

// ============================================================================
// Mutation Hooks
// ============================================================================

/**
 * Hook to create a new session
 * @param client - API client instance
 * @param options - Mutation callbacks (onSuccess, onError, etc.)
 * @returns Mutation result for creating session
 *
 * @example
 * ```tsx
 * const createSession = useCreateSession(client, {
 *   onSuccess: (session) => {
 *     console.log("Created:", session);
 *   },
 * });
 *
 * createSession.mutate({ name: "My Session" });
 * ```
 */
export function useCreateSession(
  client: ApiClient,
  options?: {
    onSuccess?: (data: Session) => void;
    onError?: (error: Error) => void;
    onSettled?: () => void;
  },
) {
  const queryClient = useQueryClient();

  return useMutation({
    ...createSessionMutation(client, queryClient),
    ...options,
  });
}

/**
 * Hook to delete a session
 * @param client - API client instance
 * @param options - Mutation callbacks (onSuccess, onError, etc.)
 * @returns Mutation result for deleting session
 *
 * @example
 * ```tsx
 * const deleteSession = useDeleteSession(client, {
 *   onSuccess: () => {
 *     console.log("Deleted successfully");
 *   },
 * });
 *
 * deleteSession.mutate("session-id");
 * ```
 */
export function useDeleteSession(
  client: ApiClient,
  options?: {
    onSuccess?: () => void;
    onError?: (error: Error) => void;
    onSettled?: () => void;
  },
) {
  const queryClient = useQueryClient();

  return useMutation({
    ...deleteSessionMutation(client, queryClient),
    ...options,
  });
}

/**
 * Hook to reconnect a session
 * @param client - API client instance
 * @param options - Mutation callbacks (onSuccess, onError, etc.)
 * @returns Mutation result for reconnecting session
 *
 * @example
 * ```tsx
 * const reconnect = useReconnectSession(client, {
 *   onSuccess: () => {
 *     console.log("Reconnected successfully");
 *   },
 * });
 *
 * reconnect.mutate("session-id");
 * ```
 */
export function useReconnectSession(
  client: ApiClient,
  options?: {
    onSuccess?: () => void;
    onError?: (error: Error) => void;
    onSettled?: () => void;
  },
) {
  const queryClient = useQueryClient();

  return useMutation({
    ...reconnectSessionMutation(client, queryClient),
    ...options,
  });
}

/**
 * Hook to disconnect a session
 * @param client - API client instance
 * @param options - Mutation callbacks (onSuccess, onError, etc.)
 * @returns Mutation result for disconnecting session
 *
 * @example
 * ```tsx
 * const disconnect = useDisconnectSession(client, {
 *   onSuccess: () => {
 *     console.log("Disconnected successfully");
 *   },
 * });
 *
 * disconnect.mutate("session-id");
 * ```
 */
export function useDisconnectSession(
  client: ApiClient,
  options?: {
    onSuccess?: () => void;
    onError?: (error: Error) => void;
    onSettled?: () => void;
  },
) {
  const queryClient = useQueryClient();

  return useMutation({
    ...disconnectSessionMutation(client, queryClient),
    ...options,
  });
}
