/**
 * Webhook Mutation Options
 * Centralized mutation configurations for webhook-related operations
 */
import { type MutationOptions, type QueryClient } from "@tanstack/react-query";
import { ApiClient, ApiClientError } from "@whatspire/api";
import type {
  WebhookConfig,
  UpdateWebhookConfigRequest,
} from "@whatspire/schema";
import { webhookKeys } from "../query-options/webhooks";

// ============================================================================
// Mutation Options Factories
// ============================================================================

/**
 * Mutation options for updating webhook configuration
 * @param client - API client instance
 * @param queryClient - Query client for cache updates
 * @returns Mutation options for useMutation
 */
export const updateWebhookConfigMutation = (
  client: ApiClient,
  queryClient: QueryClient,
): MutationOptions<
  WebhookConfig,
  ApiClientError,
  { sessionId: string; data: UpdateWebhookConfigRequest }
> => ({
  mutationFn: ({ sessionId, data }) =>
    client.updateWebhookConfig(sessionId, data),
  onSuccess: (webhookConfig, { sessionId }) => {
    // Update webhook config cache
    queryClient.setQueryData(webhookKeys.config(sessionId), webhookConfig);
  },
  onError: (error) => {
    console.error("Failed to update webhook config:", error);
  },
});

/**
 * Mutation options for rotating webhook secret
 * @param client - API client instance
 * @param queryClient - Query client for cache updates
 * @returns Mutation options for useMutation
 */
export const rotateWebhookSecretMutation = (
  client: ApiClient,
  queryClient: QueryClient,
): MutationOptions<WebhookConfig, ApiClientError, string> => ({
  mutationFn: (sessionId) => client.rotateWebhookSecret(sessionId),
  onSuccess: (webhookConfig, sessionId) => {
    // Update webhook config cache with new secret
    queryClient.setQueryData(webhookKeys.config(sessionId), webhookConfig);
  },
  onError: (error) => {
    console.error("Failed to rotate webhook secret:", error);
  },
});

/**
 * Mutation options for deleting webhook configuration
 * @param client - API client instance
 * @param queryClient - Query client for cache updates
 * @returns Mutation options for useMutation
 */
export const deleteWebhookConfigMutation = (
  client: ApiClient,
  queryClient: QueryClient,
): MutationOptions<void, ApiClientError, string> => ({
  mutationFn: (sessionId) => client.deleteWebhookConfig(sessionId),
  onSuccess: (_, sessionId) => {
    // Remove webhook config cache
    queryClient.removeQueries({ queryKey: webhookKeys.config(sessionId) });
  },
  onError: (error) => {
    console.error("Failed to delete webhook config:", error);
  },
});
