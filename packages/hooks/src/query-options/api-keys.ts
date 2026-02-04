/**
 * API Key Query Options
 * Centralized query configurations for API key management operations
 */
import { queryOptions } from "@tanstack/react-query";
import { ApiClient } from "@whatspire/api";

// ============================================================================
// Query Keys Factory
// ============================================================================

export const apiKeyKeys = {
  all: ["apikeys"] as const,
  lists: () => [...apiKeyKeys.all, "list"] as const,
  list: (filters?: Record<string, unknown>) =>
    [...apiKeyKeys.lists(), filters] as const,
  details: () => [...apiKeyKeys.all, "detail"] as const,
  detail: (id: string) => [...apiKeyKeys.details(), id] as const,
} as const;

// ============================================================================
// Query Options Factories
// ============================================================================

/**
 * Query options for listing API keys
 * @param client - API client instance
 * @param params - Optional filters and pagination parameters
 * @returns Query options for useQuery
 *
 * @example
 * ```tsx
 * const { data } = useQuery(listAPIKeysOptions(apiClient, { page: 1, limit: 50 }));
 * ```
 */
export const listAPIKeysOptions = (
  client: ApiClient,
  params?: {
    page?: number;
    limit?: number;
    role?: "read" | "write" | "admin";
    status?: "active" | "revoked";
  },
) =>
  queryOptions({
    queryKey: apiKeyKeys.list(params),
    queryFn: () => client.listAPIKeys(params),
    staleTime: 30000, // 30 seconds
  });

/**
 * Query options for getting API key details
 * @param client - API client instance
 * @param id - API key ID
 * @returns Query options for useQuery
 *
 * @example
 * ```tsx
 * const { data } = useQuery(getAPIKeyDetailsOptions(apiClient, "key_123"));
 * ```
 */
export const getAPIKeyDetailsOptions = (client: ApiClient, id: string) =>
  queryOptions({
    queryKey: apiKeyKeys.detail(id),
    queryFn: () => client.getAPIKeyDetails(id),
    staleTime: 30000, // 30 seconds
    enabled: !!id, // Only fetch if ID is provided
  });
