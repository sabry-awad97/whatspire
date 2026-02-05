/**
 * API Key Mutation Options
 * Centralized mutation configurations for API key management operations
 */
import { type MutationOptions, type QueryClient } from "@tanstack/react-query";
import { ApiClient, ApiClientError } from "@whatspire/api";
import type {
  CreateAPIKeyRequest,
  CreateAPIKeyResponse,
  RevokeAPIKeyResponse,
} from "@whatspire/schema";
import { apiKeyKeys } from "../query-options/api-keys";

// ============================================================================
// Mutation Options Factories
// ============================================================================

/**
 * Mutation options for creating an API key
 * @param client - API client instance
 * @param queryClient - Query client for cache updates
 * @returns Mutation options for useMutation
 *
 * @example
 * ```tsx
 * const mutation = useMutation(createAPIKeyMutation(apiClient, queryClient));
 * mutation.mutate({ role: "read", description: "Read-only key" });
 * ```
 */
export const createAPIKeyMutation = (
  client: ApiClient,
  queryClient: QueryClient,
): MutationOptions<
  CreateAPIKeyResponse,
  ApiClientError,
  CreateAPIKeyRequest
> => ({
  mutationFn: (data) => client.createAPIKey(data),
  onSuccess: () => {
    // Invalidate API keys list to trigger refetch
    queryClient.invalidateQueries({ queryKey: apiKeyKeys.lists() });
  },
  onError: (error) => {
    console.error("Failed to create API key:", error);
  },
});

/**
 * Mutation options for revoking an API key
 * @param client - API client instance
 * @param queryClient - Query client for cache updates
 * @returns Mutation options for useMutation
 *
 * @example
 * ```tsx
 * const mutation = useMutation(revokeAPIKeyMutation(apiClient, queryClient));
 * mutation.mutate({ id: "key-123", reason: "Compromised" });
 * ```
 */
export const revokeAPIKeyMutation = (
  client: ApiClient,
  queryClient: QueryClient,
): MutationOptions<
  RevokeAPIKeyResponse,
  ApiClientError,
  { id: string; reason?: string }
> => ({
  mutationFn: ({ id, reason }) => client.revokeAPIKey(id, reason),
  onSuccess: (_, { id }) => {
    // Invalidate API keys list to trigger refetch
    queryClient.invalidateQueries({ queryKey: apiKeyKeys.lists() });

    // Remove API key detail cache
    queryClient.removeQueries({ queryKey: apiKeyKeys.detail(id) });
  },
  onError: (error) => {
    console.error("Failed to revoke API key:", error);
  },
});
