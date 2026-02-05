/**
 * Whatspire API Provider
 * React context provider for API client instance
 */
import React, { createContext, useContext, useMemo } from "react";
import { ApiClient, type ApiClientConfig } from "@whatspire/api";

// ============================================================================
// Context
// ============================================================================

const ApiClientContext = createContext<ApiClient | null>(null);

// ============================================================================
// Provider Component
// ============================================================================

export interface WhatspireProviderProps {
  children: React.ReactNode;
  config?: ApiClientConfig;
  client?: ApiClient;
}

/**
 * Provider component for Whatspire API client
 * Provides API client instance to all child components
 *
 * @example
 * ```tsx
 * <WhatspireProvider config={{ baseURL: "http://localhost:8080" }}>
 *   <App />
 * </WhatspireProvider>
 * ```
 */
export function WhatspireProvider({
  children,
  config,
  client: providedClient,
}: WhatspireProviderProps) {
  const client = useMemo(() => {
    if (providedClient) return providedClient;
    return new ApiClient(config);
  }, [providedClient, config]);

  return (
    <ApiClientContext.Provider value={client}>
      {children}
    </ApiClientContext.Provider>
  );
}

// ============================================================================
// Hook
// ============================================================================

/**
 * Hook to access the API client instance
 * Must be used within a WhatspireProvider
 *
 * @throws Error if used outside of WhatspireProvider
 * @returns API client instance
 *
 * @example
 * ```tsx
 * function MyComponent() {
 *   const client = useApiClient();
 *   const { data: sessions } = useSessions(client);
 *   return <div>{sessions?.length} sessions</div>;
 * }
 * ```
 */
export function useApiClient(): ApiClient {
  const client = useContext(ApiClientContext);

  if (!client) {
    throw new Error(
      "useApiClient must be used within a WhatspireProvider. " +
        "Wrap your component tree with <WhatspireProvider>.",
    );
  }

  return client;
}
