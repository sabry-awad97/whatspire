/**
 * Group Query Options
 * Centralized query configurations for group-related operations
 */
import { queryOptions } from "@tanstack/react-query";
import { ApiClient } from "@whatspire/api";

// ============================================================================
// Query Keys Factory
// ============================================================================

export const groupKeys = {
  all: ["groups"] as const,
  lists: () => [...groupKeys.all, "list"] as const,
  list: (sessionId: string) => [...groupKeys.lists(), sessionId] as const,
} as const;

// ============================================================================
// Query Options Factories
// ============================================================================

/**
 * Query options for syncing groups
 * Note: This is a POST operation but we treat it as a query since it fetches data
 * @param client - API client instance
 * @param sessionId - Session ID
 * @returns Query options for useQuery
 */
export const syncGroupsOptions = (client: ApiClient, sessionId: string) =>
  queryOptions({
    queryKey: groupKeys.list(sessionId),
    queryFn: () => client.syncGroups(sessionId),
    staleTime: 1000 * 60 * 10, // 10 minutes
    gcTime: 1000 * 60 * 30, // 30 minutes
    enabled: !!sessionId,
  });
