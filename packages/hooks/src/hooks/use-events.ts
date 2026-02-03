/**
 * Event Hooks
 * Custom React hooks for event-related operations
 */
import { useQuery } from "@tanstack/react-query";
import { ApiClient } from "@whatspire/api";
import type { QueryEventsRequest } from "@whatspire/schema";
import { queryEventsOptions } from "../query-options/events";

// ============================================================================
// Query Hooks
// ============================================================================

/**
 * Hook to query events with filters
 * @param client - API client instance
 * @param filters - Event query filters
 * @param options - Additional query options
 * @returns Query result with events
 */
export function useEvents(
  client: ApiClient,
  filters: QueryEventsRequest,
  options?: {
    staleTime?: number;
    gcTime?: number;
    enabled?: boolean;
    refetchOnWindowFocus?: boolean;
  },
) {
  return useQuery({
    ...queryEventsOptions(client, filters),
    ...options,
  });
}
