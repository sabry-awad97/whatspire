/**
 * Event Query Options
 * Centralized query configurations for event-related operations
 */
import { queryOptions } from "@tanstack/react-query";
import { ApiClient } from "@whatspire/api";
import type { QueryEventsRequest } from "@whatspire/schema";

// ============================================================================
// Query Keys Factory
// ============================================================================

export const eventKeys = {
  all: ["events"] as const,
  lists: () => [...eventKeys.all, "list"] as const,
  list: (filters: QueryEventsRequest) =>
    [...eventKeys.lists(), filters] as const,
} as const;

// ============================================================================
// Query Options Factories
// ============================================================================

/**
 * Query options for querying events
 * @param client - API client instance
 * @param filters - Event query filters
 * @returns Query options for useQuery
 */
export const queryEventsOptions = (
  client: ApiClient,
  filters: QueryEventsRequest,
) =>
  queryOptions({
    queryKey: eventKeys.list(filters),
    queryFn: () => client.queryEvents(filters),
    staleTime: 1000 * 30, // 30 seconds (events are time-sensitive)
    gcTime: 1000 * 60 * 5, // 5 minutes
    enabled: !!filters.session_id || !!filters.event_type,
  });
