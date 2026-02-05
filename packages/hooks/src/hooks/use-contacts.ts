/**
 * Contact Hooks
 * Custom React hooks for contact-related operations
 */
import { useQuery } from "@tanstack/react-query";
import { ApiClient } from "@whatspire/api";
import {
  listContactsOptions,
  contactProfileOptions,
  listChatsOptions,
  checkPhoneOptions,
} from "../query-options/contacts";

// ============================================================================
// Query Hooks
// ============================================================================

/**
 * Hook to fetch contacts for a session
 * @param client - API client instance
 * @param sessionId - Session ID
 * @param options - Additional query options
 * @returns Query result with contacts list
 */
export function useContacts(
  client: ApiClient,
  sessionId: string,
  options?: {
    staleTime?: number;
    gcTime?: number;
    enabled?: boolean;
    refetchOnWindowFocus?: boolean;
  },
) {
  return useQuery({
    ...listContactsOptions(client, sessionId),
    ...options,
  });
}

/**
 * Hook to fetch contact profile
 * @param client - API client instance
 * @param sessionId - Session ID
 * @param jid - Contact JID
 * @param options - Additional query options
 * @returns Query result with contact profile
 */
export function useContactProfile(
  client: ApiClient,
  sessionId: string,
  jid: string,
  options?: {
    staleTime?: number;
    gcTime?: number;
    enabled?: boolean;
    refetchOnWindowFocus?: boolean;
  },
) {
  return useQuery({
    ...contactProfileOptions(client, sessionId, jid),
    ...options,
  });
}

/**
 * Hook to fetch chats for a session
 * @param client - API client instance
 * @param sessionId - Session ID
 * @param options - Additional query options
 * @returns Query result with chats list
 */
export function useChats(
  client: ApiClient,
  sessionId: string,
  options?: {
    staleTime?: number;
    gcTime?: number;
    enabled?: boolean;
    refetchOnWindowFocus?: boolean;
  },
) {
  return useQuery({
    ...listChatsOptions(client, sessionId),
    ...options,
  });
}

/**
 * Hook to check if phone is on WhatsApp
 * @param client - API client instance
 * @param sessionId - Session ID
 * @param phone - Phone number to check
 * @param options - Additional query options
 * @returns Query result with phone check status
 */
export function useCheckPhone(
  client: ApiClient,
  sessionId: string,
  phone: string,
  options?: {
    staleTime?: number;
    gcTime?: number;
    enabled?: boolean;
    refetchOnWindowFocus?: boolean;
  },
) {
  return useQuery({
    ...checkPhoneOptions(client, sessionId, phone),
    ...options,
  });
}
