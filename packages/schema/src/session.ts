/**
 * Session-related schemas
 * Matches backend: dto/response.go (SessionResponse) and dto/request.go (CreateSessionRequest)
 */
import { z } from "zod";
import {
  sessionIdSchema,
  sessionStatusSchema,
  timestampSchema,
} from "./common";

/**
 * Session response schema
 * Matches: dto.SessionResponse and entity.Session
 */
export const sessionSchema = z.object({
  id: sessionIdSchema,
  jid: z.string().optional(),
  name: z.string().min(1).max(100),
  status: sessionStatusSchema,
  created_at: timestampSchema,
  updated_at: timestampSchema,
  // History sync configuration (optional, from entity.Session)
  history_sync_enabled: z.boolean().optional(),
  full_sync: z.boolean().optional(),
  sync_since: timestampSchema.optional(),
});

export type Session = z.infer<typeof sessionSchema>;

/**
 * Create session request schema
 * Matches: dto.CreateSessionRequest
 */
export const createSessionRequestSchema = z.object({
  name: z.string().min(1).max(100),
  config: z
    .object({
      account_protection: z.boolean().optional(),
      message_logging: z.boolean().optional(),
      read_messages: z.boolean().optional(),
      auto_reject_calls: z.boolean().optional(),
      always_online: z.boolean().optional(),
      ignore_groups: z.boolean().optional(),
      ignore_broadcasts: z.boolean().optional(),
      ignore_channels: z.boolean().optional(),
    })
    .optional(),
});

export type CreateSessionRequest = z.infer<typeof createSessionRequestSchema>;

/**
 * Get session request schema
 * Matches: dto.GetSessionRequest
 */
export const getSessionRequestSchema = z.object({
  id: sessionIdSchema,
});

export type GetSessionRequest = z.infer<typeof getSessionRequestSchema>;

/**
 * Delete session request schema
 * Matches: dto.DeleteSessionRequest
 */
export const deleteSessionRequestSchema = z.object({
  id: sessionIdSchema,
});

export type DeleteSessionRequest = z.infer<typeof deleteSessionRequestSchema>;

/**
 * Start QR auth request schema
 * Matches: dto.StartQRAuthRequest
 */
export const startQRAuthRequestSchema = z.object({
  session_id: sessionIdSchema,
});

export type StartQRAuthRequest = z.infer<typeof startQRAuthRequestSchema>;

/**
 * Session list response schema
 */
export const sessionListSchema = z.array(sessionSchema);
export type SessionList = z.infer<typeof sessionListSchema>;
