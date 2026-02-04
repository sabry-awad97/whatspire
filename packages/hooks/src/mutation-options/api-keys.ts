/**
 * API Key Mutation Options
 * Centralized mutation configurations for API key management operations
 */
import { type MutationOptions, type QueryClient } from "@tanstack/react-query";
import { ApiClient, ApiClientError } from "@whatspire/api";
import type {
  CreateAPIKeyRequest,
  CreateAPIKeyResponse,
} from "@whatspire/schema";

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
    // Note: We'll add the query key when we implement the list query
    queryClient.invalidateQueries({ queryKey: ["apikeys"] });
  },
  onError: (error) => {
    console.error("Failed to create API key:", error);
  },
});
