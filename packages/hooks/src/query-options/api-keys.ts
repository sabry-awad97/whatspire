/**
 * API Key Query Options
 * Centralized query configurations for API key management operations
 */

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

// Note: Query options will be implemented in Phase 5 (US-003)
// when the listAPIKeys endpoint is added to the API client
