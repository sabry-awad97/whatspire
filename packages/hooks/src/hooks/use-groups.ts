/**
 * Group Hooks
 * Custom React hooks for group-related operations
 */
import { useQuery } from "@tanstack/react-query";
import { ApiClient } from "@whatspire/api";
import { syncGroupsOptions } from "../query-options/groups";

// ============================================================================
// Query Hooks
// ============================================================================

/**
 * Hook to sync and fetch groups for a session
 * Note: This triggers a sync operation on the backend
 * @param client - API client instance
 * @param sessionId - Session ID
 * @param options - Additional query options
 * @returns Query result with groups list
 */
export function useGroups(
  client: ApiClient,
  sessionId: string,
  options?: {
    staleTime?: number;
    gcTime?: number;
    enabled?: boolean;
    refetchOnWindowFocus?: boolean;
  },
) {
  return useQuery({
    ...syncGroupsOptions(client, sessionId),
    ...options,
  });
}
