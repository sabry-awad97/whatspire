/**
 * Contact Query Options
 * Centralized query configurations for contact-related operations
 */
import { queryOptions } from "@tanstack/react-query";
import { ApiClient } from "@whatspire/api";

// ============================================================================
// Query Keys Factory
// ============================================================================

export const contactKeys = {
  all: ["contacts"] as const,
  lists: () => [...contactKeys.all, "list"] as const,
  list: (sessionId: string) => [...contactKeys.lists(), sessionId] as const,
  details: () => [...contactKeys.all, "detail"] as const,
  detail: (sessionId: string, jid: string) =>
    [...contactKeys.details(), sessionId, jid] as const,
  chats: () => [...contactKeys.all, "chats"] as const,
  chatList: (sessionId: string) => [...contactKeys.chats(), sessionId] as const,
  phone: () => [...contactKeys.all, "phone"] as const,
  phoneCheck: (sessionId: string, phone: string) =>
    [...contactKeys.phone(), sessionId, phone] as const,
} as const;

// ============================================================================
// Query Options Factories
// ============================================================================

/**
 * Query options for listing contacts
 * @param client - API client instance
 * @param sessionId - Session ID
 * @returns Query options for useQuery
 */
export const listContactsOptions = (client: ApiClient, sessionId: string) =>
  queryOptions({
    queryKey: contactKeys.list(sessionId),
    queryFn: () => client.getContacts(sessionId),
    staleTime: 1000 * 60 * 5, // 5 minutes
    gcTime: 1000 * 60 * 15, // 15 minutes
    enabled: !!sessionId,
  });

/**
 * Query options for getting contact profile
 * @param client - API client instance
 * @param sessionId - Session ID
 * @param jid - Contact JID
 * @returns Query options for useQuery
 */
export const contactProfileOptions = (
  client: ApiClient,
  sessionId: string,
  jid: string,
) =>
  queryOptions({
    queryKey: contactKeys.detail(sessionId, jid),
    queryFn: () => client.getContactProfile({ session_id: sessionId, jid }),
    staleTime: 1000 * 60 * 10, // 10 minutes
    gcTime: 1000 * 60 * 30, // 30 minutes
    enabled: !!sessionId && !!jid,
  });

/**
 * Query options for listing chats
 * @param client - API client instance
 * @param sessionId - Session ID
 * @returns Query options for useQuery
 */
export const listChatsOptions = (client: ApiClient, sessionId: string) =>
  queryOptions({
    queryKey: contactKeys.chatList(sessionId),
    queryFn: () => client.getChats(sessionId),
    staleTime: 1000 * 60 * 2, // 2 minutes
    gcTime: 1000 * 60 * 10, // 10 minutes
    enabled: !!sessionId,
  });

/**
 * Query options for checking if phone is on WhatsApp
 * @param client - API client instance
 * @param sessionId - Session ID
 * @param phone - Phone number to check
 * @returns Query options for useQuery
 */
export const checkPhoneOptions = (
  client: ApiClient,
  sessionId: string,
  phone: string,
) =>
  queryOptions({
    queryKey: contactKeys.phoneCheck(sessionId, phone),
    queryFn: () => client.checkPhone({ session_id: sessionId, phone }),
    staleTime: 1000 * 60 * 60, // 1 hour (phone status doesn't change often)
    gcTime: 1000 * 60 * 120, // 2 hours
    enabled: !!sessionId && !!phone,
  });
