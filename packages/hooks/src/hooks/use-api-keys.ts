/**
 * API Key Hooks
 * Custom React hooks for API key management operations
 */
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { ApiClient } from "@whatspire/api";
import type { CreateAPIKeyResponse } from "@whatspire/schema";
import { createAPIKeyMutation } from "../mutation-options/api-keys";

// ============================================================================
// Mutation Hooks
// ============================================================================

/**
 * Hook to create a new API key
 * @param client - API client instance
 * @param options - Mutation callbacks (onSuccess, onError, etc.)
 * @returns Mutation result for creating API key
 *
 * @example
 * ```tsx
 * const createAPIKey = useCreateAPIKey(client, {
 *   onSuccess: (response) => {
 *     console.log("Created:", response.api_key);
 *     console.log("Plain key (shown once):", response.plain_key);
 *   },
 * });
 *
 * createAPIKey.mutate({ role: "read", description: "Read-only key" });
 * ```
 */
export function useCreateAPIKey(
  client: ApiClient,
  options?: {
    onSuccess?: (data: CreateAPIKeyResponse) => void;
    onError?: (error: Error) => void;
    onSettled?: () => void;
  },
) {
  const queryClient = useQueryClient();

  return useMutation({
    ...createAPIKeyMutation(client, queryClient),
    ...options,
  });
}
