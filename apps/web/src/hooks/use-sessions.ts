/**
 * Reusable Sessions Hook
 * Provides consistent session query configuration across the application
 */
import { useApiClient, useSessions as useSessionsBase } from "@whatspire/hooks";

/**
 * Hook to fetch all sessions with optimized refetch settings
 *
 * This hook provides a consistent configuration for session queries across the app:
 * - Immediate refetch on window focus
 * - Immediate refetch on mount
 * - No stale time (always fresh data after mutations)
 *
 * @returns Query result with sessions list
 *
 * @example
 * ```tsx
 * function MyComponent() {
 *   const { data: sessions, isLoading, error, refetch } = useSessions();
 *
 *   if (isLoading) return <div>Loading...</div>;
 *   if (error) return <div>Error: {error.message}</div>;
 *
 *   return (
 *     <div>
 *       {sessions.map(session => (
 *         <div key={session.id}>{session.name}</div>
 *       ))}
 *     </div>
 *   );
 * }
 * ```
 */
export function useSessions() {
  const client = useApiClient();

  return useSessionsBase(client, {
    // Refetch when window regains focus
    refetchOnWindowFocus: true,

    // Refetch when component mounts
    refetchOnMount: true,

    // Always consider data stale for immediate updates after mutations
    // This ensures UI updates immediately after reconnect/disconnect/delete
    staleTime: 0,

    // Keep data in cache for 5 minutes after component unmounts
    gcTime: 1000 * 60 * 5,
  });
}
