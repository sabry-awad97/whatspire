/**
 * Reaction-related schemas
 * Matches backend: dto/reaction.go
 */
import { z } from "zod";
import { jidSchema, sessionIdSchema, timestampSchema } from "./common";

/**
 * Send reaction request schema
 * Matches: dto.SendReactionRequest
 */
export const sendReactionRequestSchema = z.object({
  session_id: sessionIdSchema,
  chat_jid: jidSchema,
  message_id: z.string().min(1),
  emoji: z.string().min(1),
});

export type SendReactionRequest = z.infer<typeof sendReactionRequestSchema>;

/**
 * Remove reaction request schema
 * Matches: dto.RemoveReactionRequest
 */
export const removeReactionRequestSchema = z.object({
  session_id: sessionIdSchema,
  chat_jid: jidSchema,
  message_id: z.string().min(1),
});

export type RemoveReactionRequest = z.infer<typeof removeReactionRequestSchema>;

/**
 * Reaction response schema
 * Matches: dto.ReactionResponse
 */
export const reactionResponseSchema = z.object({
  message_id: z.string(),
  emoji: z.string(),
  timestamp: timestampSchema,
});

export type ReactionResponse = z.infer<typeof reactionResponseSchema>;
