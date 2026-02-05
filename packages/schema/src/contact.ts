/**
 * Contact-related schemas
 * Matches backend: dto/contact.go
 */
import { z } from "zod";
import { jidSchema, sessionIdSchema, urlSchema } from "./common";

/**
 * Contact response schema
 * Matches: dto.ContactResponse
 */
export const contactSchema = z.object({
  jid: jidSchema,
  name: z.string(),
  avatar_url: urlSchema.optional().nullable(),
  status: z.string().optional().nullable(),
  is_on_whatsapp: z.boolean(),
});

export type Contact = z.infer<typeof contactSchema>;

/**
 * Contact list response schema
 * Matches: dto.ContactListResponse
 */
export const contactListSchema = z.object({
  contacts: z.array(contactSchema),
});

export type ContactList = z.infer<typeof contactListSchema>;

/**
 * Check phone request schema
 * Matches: dto.CheckPhoneRequest
 */
export const checkPhoneRequestSchema = z.object({
  session_id: sessionIdSchema,
  phone: z.string().min(1),
});

export type CheckPhoneRequest = z.infer<typeof checkPhoneRequestSchema>;

/**
 * Get profile request schema
 * Matches: dto.GetProfileRequest
 */
export const getProfileRequestSchema = z.object({
  session_id: sessionIdSchema,
  jid: jidSchema,
});

export type GetProfileRequest = z.infer<typeof getProfileRequestSchema>;

/**
 * Chat response schema
 * Matches: dto.ChatResponse
 */
export const chatSchema = z.object({
  jid: jidSchema,
  name: z.string(),
  last_message_time: z.string().datetime().optional().nullable(),
  unread_count: z.number().int().min(0),
  is_group: z.boolean(),
  avatar_url: urlSchema.optional().nullable(),
  archived: z.boolean(),
  pinned: z.boolean(),
});

export type Chat = z.infer<typeof chatSchema>;

/**
 * Chat list response schema
 * Matches: dto.ChatListResponse
 */
export const chatListSchema = z.object({
  chats: z.array(chatSchema),
});

export type ChatList = z.infer<typeof chatListSchema>;
