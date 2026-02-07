/**
 * Webhook Query Options
 * Centralized query configurations for webhook-related operations
 */
import { queryOptions } from "@tanstack/react-query";
import { ApiClient } from "@whatspire/api";

// ============================================================================
// Query Keys Factory
// ============================================================================

export const webhookKeys = {
  all: ["webhooks"] as const,
  configs: () => [...webhookKeys.all, "config"] as const,
  config: (sessionId: string) => [...webhookKeys.configs(), sessionId] as const,
} as const;

// ============================================================================
// Query Options Factories
// ============================================================================

/**
 * Query options for getting webhook configuration for a session
 * @param client - API client instance
 * @param sessionId - Session ID to fetch webhook config for
 * @returns Query options for useQuery
 */
export const webhookConfigOptions = (client: ApiClient, sessionId: string) =>
  queryOptions({
    queryKey: webhookKeys.config(sessionId),
    queryFn: () => client.getWebhookConfig(sessionId),
    staleTime: 1000 * 60 * 5, // 5 minutes
    gcTime: 1000 * 60 * 10, // 10 minutes
    enabled: !!sessionId,
  });
