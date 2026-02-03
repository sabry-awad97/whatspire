/**
 * Session Query Options
 * Centralized query configurations for session-related operations
 */
import { queryOptions } from "@tanstack/react-query";
import { ApiClient } from "@whatspire/api";

// ============================================================================
// Query Keys Factory
// ============================================================================

export const sessionKeys = {
  all: ["sessions"] as const,
  lists: () => [...sessionKeys.all, "list"] as const,
  list: (filters?: Record<string, unknown>) =>
    [...sessionKeys.lists(), filters] as const,
  details: () => [...sessionKeys.all, "detail"] as const,
  detail: (id: string) => [...sessionKeys.details(), id] as const,
} as const;

// ============================================================================
// Query Options Factories
// ============================================================================

/**
 * Query options for listing all sessions
 * @param client - API client instance
 * @returns Query options for useQuery
 */
export const listSessionsOptions = (client: ApiClient) =>
  queryOptions({
    queryKey: sessionKeys.lists(),
    queryFn: () => client.listSessions(),
    staleTime: 1000 * 60 * 2, // 2 minutes
    gcTime: 1000 * 60 * 10, // 10 minutes
  });

/**
 * Query options for getting a single session
 * @param client - API client instance
 * @param sessionId - Session ID to fetch
 * @returns Query options for useQuery
 */
export const sessionDetailOptions = (client: ApiClient, sessionId: string) =>
  queryOptions({
    queryKey: sessionKeys.detail(sessionId),
    queryFn: () => client.getSession(sessionId),
    staleTime: 1000 * 60 * 1, // 1 minute
    gcTime: 1000 * 60 * 5, // 5 minutes
    enabled: !!sessionId,
  });
