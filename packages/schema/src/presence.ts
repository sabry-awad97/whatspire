/**
 * Presence-related schemas
 * Matches backend: dto/presence.go
 */
import { z } from "zod";
import {
  jidSchema,
  sessionIdSchema,
  presenceStateSchema,
  timestampSchema,
} from "./common";

/**
 * Send presence request schema
 * Matches: dto.SendPresenceRequest
 */
export const sendPresenceRequestSchema = z.object({
  session_id: sessionIdSchema,
  chat_jid: jidSchema,
  state: presenceStateSchema,
});

export type SendPresenceRequest = z.infer<typeof sendPresenceRequestSchema>;

/**
 * Presence response schema
 * Matches: dto.PresenceResponse
 */
export const presenceResponseSchema = z.object({
  chat_jid: jidSchema,
  state: presenceStateSchema,
  timestamp: timestampSchema,
});

export type PresenceResponse = z.infer<typeof presenceResponseSchema>;
