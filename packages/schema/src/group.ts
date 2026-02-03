/**
 * Group-related schemas
 * Matches backend: dto/response.go (GroupResponse, ParticipantResponse)
 */
import { z } from "zod";
import {
  jidSchema,
  participantRoleSchema,
  timestampSchema,
  urlSchema,
} from "./common";

/**
 * Participant response schema
 * Matches: dto.ParticipantResponse
 */
export const participantSchema = z.object({
  jid: jidSchema,
  role: participantRoleSchema,
  display_name: z.string().optional().nullable(),
  avatar_url: urlSchema.optional().nullable(),
});

export type Participant = z.infer<typeof participantSchema>;

/**
 * Group response schema
 * Matches: dto.GroupResponse
 */
export const groupSchema = z.object({
  jid: jidSchema,
  name: z.string(),
  description: z.string().optional().nullable(),
  avatar_url: urlSchema.optional().nullable(),
  is_announce: z.boolean(),
  is_locked: z.boolean(),
  is_ephemeral: z.boolean(),
  ephemeral_time: z.number().int().optional().nullable(),
  owner_jid: jidSchema.optional().nullable(),
  member_count: z.number().int().min(0),
  group_created_at: timestampSchema.optional().nullable(),
  participants: z.array(participantSchema).optional(),
});

export type Group = z.infer<typeof groupSchema>;

/**
 * Group list response schema
 */
export const groupListSchema = z.array(groupSchema);
export type GroupList = z.infer<typeof groupListSchema>;
