/**
 * Message Mutation Options
 * Centralized mutation configurations for message-related operations
 */
import { type MutationOptions } from "@tanstack/react-query";
import { ApiClient, ApiClientError } from "@whatspire/api";
import type {
  SendMessageRequest,
  SendPresenceRequest,
  PresenceResponse,
  SendReactionRequest,
  RemoveReactionRequest,
  ReactionResponse,
  SendReceiptRequest,
  ReceiptResponse,
} from "@whatspire/schema";

// ============================================================================
// Mutation Options Factories
// ============================================================================

/**
 * Mutation options for sending a message
 * @param client - API client instance
 * @returns Mutation options for useMutation
 */
export const sendMessageMutation = (
  client: ApiClient,
): MutationOptions<
  { id: string; status: string },
  ApiClientError,
  SendMessageRequest
> => ({
  mutationFn: (data) => client.sendMessage(data),
  onError: (error) => {
    console.error("Failed to send message:", error);
  },
});

/**
 * Mutation options for sending presence
 * @param client - API client instance
 * @returns Mutation options for useMutation
 */
export const sendPresenceMutation = (
  client: ApiClient,
): MutationOptions<PresenceResponse, ApiClientError, SendPresenceRequest> => ({
  mutationFn: (data) => client.sendPresence(data),
  onError: (error) => {
    console.error("Failed to send presence:", error);
  },
});

/**
 * Mutation options for sending a reaction
 * @param client - API client instance
 * @returns Mutation options for useMutation
 */
export const sendReactionMutation = (
  client: ApiClient,
): MutationOptions<ReactionResponse, ApiClientError, SendReactionRequest> => ({
  mutationFn: (data) => client.sendReaction(data),
  onError: (error) => {
    console.error("Failed to send reaction:", error);
  },
});

/**
 * Mutation options for removing a reaction
 * @param client - API client instance
 * @returns Mutation options for useMutation
 */
export const removeReactionMutation = (
  client: ApiClient,
): MutationOptions<
  ReactionResponse,
  ApiClientError,
  RemoveReactionRequest
> => ({
  mutationFn: (data) => client.removeReaction(data),
  onError: (error) => {
    console.error("Failed to remove reaction:", error);
  },
});

/**
 * Mutation options for sending read receipts
 * @param client - API client instance
 * @returns Mutation options for useMutation
 */
export const sendReceiptMutation = (
  client: ApiClient,
): MutationOptions<ReceiptResponse, ApiClientError, SendReceiptRequest> => ({
  mutationFn: (data) => client.sendReceipt(data),
  onError: (error) => {
    console.error("Failed to send receipt:", error);
  },
});
