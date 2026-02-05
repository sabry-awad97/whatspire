/**
 * Message Hooks
 * Custom React hooks for message-related operations
 */
import { useMutation } from "@tanstack/react-query";
import { ApiClient } from "@whatspire/api";
import type {
  PresenceResponse,
  ReactionResponse,
  ReceiptResponse,
} from "@whatspire/schema";
import {
  sendMessageMutation,
  sendPresenceMutation,
  sendReactionMutation,
  removeReactionMutation,
  sendReceiptMutation,
} from "../mutation-options/messages";

// ============================================================================
// Mutation Hooks
// ============================================================================

/**
 * Hook to send a WhatsApp message
 * @param client - API client instance
 * @param options - Mutation callbacks
 * @returns Mutation result for sending message
 */
export function useSendMessage(
  client: ApiClient,
  options?: {
    onSuccess?: (data: { id: string; status: string }) => void;
    onError?: (error: Error) => void;
    onSettled?: () => void;
  },
) {
  return useMutation({
    ...sendMessageMutation(client),
    ...options,
  });
}

/**
 * Hook to send presence update
 * @param client - API client instance
 * @param options - Mutation callbacks
 * @returns Mutation result for sending presence
 */
export function useSendPresence(
  client: ApiClient,
  options?: {
    onSuccess?: (data: PresenceResponse) => void;
    onError?: (error: Error) => void;
    onSettled?: () => void;
  },
) {
  return useMutation({
    ...sendPresenceMutation(client),
    ...options,
  });
}

/**
 * Hook to send a reaction to a message
 * @param client - API client instance
 * @param options - Mutation callbacks
 * @returns Mutation result for sending reaction
 */
export function useSendReaction(
  client: ApiClient,
  options?: {
    onSuccess?: (data: ReactionResponse) => void;
    onError?: (error: Error) => void;
    onSettled?: () => void;
  },
) {
  return useMutation({
    ...sendReactionMutation(client),
    ...options,
  });
}

/**
 * Hook to remove a reaction from a message
 * @param client - API client instance
 * @param options - Mutation callbacks
 * @returns Mutation result for removing reaction
 */
export function useRemoveReaction(
  client: ApiClient,
  options?: {
    onSuccess?: (data: ReactionResponse) => void;
    onError?: (error: Error) => void;
    onSettled?: () => void;
  },
) {
  return useMutation({
    ...removeReactionMutation(client),
    ...options,
  });
}

/**
 * Hook to send read receipts
 * @param client - API client instance
 * @param options - Mutation callbacks
 * @returns Mutation result for sending receipts
 */
export function useSendReceipt(
  client: ApiClient,
  options?: {
    onSuccess?: (data: ReceiptResponse) => void;
    onError?: (error: Error) => void;
    onSettled?: () => void;
  },
) {
  return useMutation({
    ...sendReceiptMutation(client),
    ...options,
  });
}
