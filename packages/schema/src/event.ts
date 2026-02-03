/**
 * Event-related schemas
 * Matches backend: dto/event.go
 */
import { z } from "zod";
import { sessionIdSchema, timestampSchema } from "./common";

/**
 * Event DTO schema
 * Matches: dto.EventDTO
 */
export const eventSchema = z.object({
  id: z.string(),
  type: z.string(),
  session_id: sessionIdSchema,
  data: z.instanceof(Uint8Array).optional(),
  timestamp: timestampSchema,
});

export type Event = z.infer<typeof eventSchema>;

/**
 * Query events request schema
 * Matches: dto.QueryEventsRequest
 */
export const queryEventsRequestSchema = z
  .object({
    session_id: sessionIdSchema.optional(),
    event_type: z.string().optional(),
    since: timestampSchema.optional(),
    until: timestampSchema.optional(),
    limit: z.number().int().min(1).max(1000).default(100),
    offset: z.number().int().min(0).default(0),
  })
  .refine(
    (data) => {
      // Validate limit
      if (data.limit && (data.limit < 1 || data.limit > 1000)) {
        return false;
      }
      // Validate offset
      if (data.offset && data.offset < 0) {
        return false;
      }
      return true;
    },
    {
      message: "Invalid limit or offset values",
    },
  );

export type QueryEventsRequest = z.infer<typeof queryEventsRequestSchema>;

/**
 * Query events response schema
 * Matches: dto.QueryEventsResponse
 */
export const queryEventsResponseSchema = z.object({
  events: z.array(eventSchema),
  total: z.number().int().min(0),
  limit: z.number().int(),
  offset: z.number().int(),
});

export type QueryEventsResponse = z.infer<typeof queryEventsResponseSchema>;

/**
 * Replay events request schema
 * Matches: dto.ReplayEventsRequest
 */
export const replayEventsRequestSchema = z
  .object({
    session_id: sessionIdSchema.optional(),
    event_type: z.string().optional(),
    since: timestampSchema.optional(),
    until: timestampSchema.optional(),
    dry_run: z.boolean().optional(),
  })
  .refine(
    (data) => {
      // At least one filter must be specified
      return !!(data.session_id || data.event_type || data.since || data.until);
    },
    {
      message:
        "At least one filter (session_id, event_type, since, or until) must be specified",
    },
  );

export type ReplayEventsRequest = z.infer<typeof replayEventsRequestSchema>;

/**
 * Replay events response schema
 * Matches: dto.ReplayEventsResponse
 */
export const replayEventsResponseSchema = z.object({
  success: z.boolean(),
  events_found: z.number().int().min(0),
  events_replayed: z.number().int().min(0),
  events_failed: z.number().int().min(0).optional(),
  dry_run: z.boolean(),
  message: z.string(),
  last_error: z.any().optional(),
});

export type ReplayEventsResponse = z.infer<typeof replayEventsResponseSchema>;
