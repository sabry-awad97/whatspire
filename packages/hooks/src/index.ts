/**
 * @whatspire/hooks
 *
 * Professional React hooks package with TanStack Query integration
 * Provides type-safe, optimized hooks for the Whatspire WhatsApp API
 */

// Provider
export { WhatspireProvider, useApiClient } from "./provider";
export type { WhatspireProviderProps } from "./provider";

// Query Options (for advanced usage)
export * from "./query-options/sessions";
export * from "./query-options/contacts";
export * from "./query-options/groups";
export * from "./query-options/events";
export * from "./query-options/api-keys";

// Mutation Options (for advanced usage)
export * from "./mutation-options/sessions";
export * from "./mutation-options/messages";
export * from "./mutation-options/api-keys";

// Hooks
export * from "./hooks/use-sessions";
export * from "./hooks/use-contacts";
export * from "./hooks/use-messages";
export * from "./hooks/use-groups";
export * from "./hooks/use-events";
export * from "./hooks/use-api-keys";

// Re-export commonly used types from dependencies
export type {
  UseQueryOptions,
  UseMutationOptions,
  QueryClient,
} from "@tanstack/react-query";
export type { ApiClient, ApiClientConfig } from "@whatspire/api";
