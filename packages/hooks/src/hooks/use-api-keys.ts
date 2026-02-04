/**
 * API Key Hooks
 * Custom React hooks for API key management operations
 */
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { ApiClient } from "@whatspire/api";
import type { CreateAPIKeyResponse } from "@whatspire/schema";
import { createAPIKeyMutation } from "../mutation-options/api-keys";
import { listAPIKeysOptions } from "../query-options/api-keys";

// ============================================================================
// Query Hooks
// ============================================================================

/**
 * Hook to list API keys with optional filtering and pagination
 * @param client - API client instance
 * @param params - Optional filters and pagination parameters
 * @returns Query result for API keys list
 *
 * @example
 * ```tsx
 * const { data, isLoading } = useAPIKeys(client, {
 *   page: 1,
 *   limit: 50,
 *   role: "admin",
 *   status: "active",
 * });
 *
 * if (data) {
 *   console.log("API Keys:", data.api_keys);
 *   console.log("Pagination:", data.pagination);
 * }
 * ```
 */
export function useAPIKeys(
  client: ApiClient,
  params?: {
    page?: number;
    limit?: number;
    role?: "read" | "write" | "admin";
    status?: "active" | "revoked";
  },
) {
  return useQuery(listAPIKeysOptions(client, params));
}

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
