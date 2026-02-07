/**
 * Webhook Hooks
 * Custom React hooks for webhook configuration operations
 */
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { ApiClient } from "@whatspire/api";
import type { WebhookConfig } from "@whatspire/schema";
import { webhookConfigOptions } from "../query-options/webhooks";
import {
  updateWebhookConfigMutation,
  rotateWebhookSecretMutation,
  deleteWebhookConfigMutation,
} from "../mutation-options/webhooks";

// ============================================================================
// Query Hooks
// ============================================================================

/**
 * Hook to fetch webhook configuration for a session
 * @param client - API client instance
 * @param sessionId - Session ID to fetch webhook config for
 * @param options - Additional query options (staleTime, enabled, etc.)
 * @returns Query result with webhook configuration
 *
 * @example
 * ```tsx
 * const { data: webhookConfig, isLoading } = useWebhookConfig(client, sessionId, {
 *   enabled: !!sessionId,
 * });
 * ```
 */
export function useWebhookConfig(
  client: ApiClient,
  sessionId: string,
  options?: {
    staleTime?: number;
    gcTime?: number;
    enabled?: boolean;
    refetchOnWindowFocus?: boolean;
    refetchOnMount?: boolean;
    refetchOnReconnect?: boolean;
  },
) {
  return useQuery({
    ...webhookConfigOptions(client, sessionId),
    ...options,
  });
}

// ============================================================================
// Mutation Hooks
// ============================================================================

/**
 * Hook to update webhook configuration
 * @param client - API client instance
 * @param options - Mutation callbacks (onSuccess, onError, etc.)
 * @returns Mutation result for updating webhook config
 *
 * @example
 * ```tsx
 * const updateWebhook = useUpdateWebhookConfig(client, {
 *   onSuccess: (config) => {
 *     console.log("Updated:", config);
 *   },
 * });
 *
 * updateWebhook.mutate({
 *   sessionId: "session-id",
 *   data: {
 *     enabled: true,
 *     url: "https://api.example.com/webhook",
 *     events: ["messages.received"],
 *     ignore_groups: false,
 *     ignore_broadcasts: false,
 *     ignore_channels: false,
 *   },
 * });
 * ```
 */
export function useUpdateWebhookConfig(
  client: ApiClient,
  options?: {
    onSuccess?: (data: WebhookConfig) => void;
    onError?: (error: Error) => void;
    onSettled?: () => void;
  },
) {
  const queryClient = useQueryClient();

  return useMutation({
    ...updateWebhookConfigMutation(client, queryClient),
    ...options,
  });
}

/**
 * Hook to rotate webhook secret
 * @param client - API client instance
 * @param options - Mutation callbacks (onSuccess, onError, etc.)
 * @returns Mutation result for rotating webhook secret
 *
 * @example
 * ```tsx
 * const rotateSecret = useRotateWebhookSecret(client, {
 *   onSuccess: (config) => {
 *     console.log("New secret:", config.secret);
 *   },
 * });
 *
 * rotateSecret.mutate("session-id");
 * ```
 */
export function useRotateWebhookSecret(
  client: ApiClient,
  options?: {
    onSuccess?: (data: WebhookConfig) => void;
    onError?: (error: Error) => void;
    onSettled?: () => void;
  },
) {
  const queryClient = useQueryClient();

  return useMutation({
    ...rotateWebhookSecretMutation(client, queryClient),
    ...options,
  });
}

/**
 * Hook to delete webhook configuration
 * @param client - API client instance
 * @param options - Mutation callbacks (onSuccess, onError, etc.)
 * @returns Mutation result for deleting webhook config
 *
 * @example
 * ```tsx
 * const deleteWebhook = useDeleteWebhookConfig(client, {
 *   onSuccess: () => {
 *     console.log("Deleted successfully");
 *   },
 * });
 *
 * deleteWebhook.mutate("session-id");
 * ```
 */
export function useDeleteWebhookConfig(
  client: ApiClient,
  options?: {
    onSuccess?: () => void;
    onError?: (error: Error) => void;
    onSettled?: () => void;
  },
) {
  const queryClient = useQueryClient();

  return useMutation({
    ...deleteWebhookConfigMutation(client, queryClient),
    ...options,
  });
}
