/**
 * Common schemas and utilities used across all schemas
 */
import { z } from "zod";

/**
 * Standard API Response wrapper
 * Matches: dto.APIResponse[T]
 */
export const apiErrorSchema = z.object({
  code: z.string(),
  message: z.string(),
  details: z.record(z.string(), z.string()).optional(),
});

export type ApiError = z.infer<typeof apiErrorSchema>;

export const apiResponseSchema = <T extends z.ZodTypeAny>(dataSchema: T) =>
  z.object({
    success: z.boolean(),
    data: dataSchema.optional(),
    error: apiErrorSchema.optional(),
  });

export type ApiResponse<T> = {
  success: boolean;
  data?: T;
  error?: ApiError;
};

/**
 * Common field schemas
 */
export const jidSchema = z.string().min(1).describe("WhatsApp JID");
export const sessionIdSchema = z.string().min(1).describe("Session ID");
export const phoneNumberSchema = z
  .string()
  .regex(/^\+?[1-9]\d{1,14}$/)
  .describe("E.164 phone number");
export const timestampSchema = z
  .string()
  .datetime()
  .describe("ISO 8601 timestamp");
export const urlSchema = z.string().url().describe("Valid URL");

/**
 * Message type enum
 * Matches: entity.MessageType
 */
export const messageTypeSchema = z.enum([
  "text",
  "image",
  "document",
  "audio",
  "video",
  "sticker",
]);
export type MessageType = z.infer<typeof messageTypeSchema>;

/**
 * Message status enum
 * Matches: entity.MessageStatus
 */
export const messageStatusSchema = z.enum([
  "pending",
  "sent",
  "delivered",
  "read",
  "failed",
]);
export type MessageStatus = z.infer<typeof messageStatusSchema>;

/**
 * Session status enum
 * Matches: entity.Status
 */
export const sessionStatusSchema = z.enum([
  "connected",
  "disconnected",
  "connecting",
  "logged_out",
  "pending",
]);
export type SessionStatus = z.infer<typeof sessionStatusSchema>;

/**
 * Presence state enum
 */
export const presenceStateSchema = z.enum([
  "typing",
  "paused",
  "online",
  "offline",
]);
export type PresenceState = z.infer<typeof presenceStateSchema>;

/**
 * Participant role enum
 */
export const participantRoleSchema = z.enum(["admin", "superadmin", "member"]);
export type ParticipantRole = z.infer<typeof participantRoleSchema>;
