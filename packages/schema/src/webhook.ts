/**
 * Webhook configuration schemas
 * Matches backend: dto/response.go (WebhookConfigResponse) and dto/request.go (UpdateWebhookConfigRequest)
 */
import { z } from "zod";
import { sessionIdSchema, timestampSchema } from "./common";

/**
 * Webhook configuration response schema
 * Matches: dto.WebhookConfigResponse and entity.WebhookConfig
 */
export const webhookConfigSchema = z.object({
  id: z.string().uuid(),
  session_id: sessionIdSchema,
  enabled: z.boolean(),
  url: z.string().url().or(z.literal("")),
  secret: z.string(),
  events: z.array(z.string()),
  ignore_groups: z.boolean(),
  ignore_broadcasts: z.boolean(),
  ignore_channels: z.boolean(),
  created_at: timestampSchema,
  updated_at: timestampSchema,
});

export type WebhookConfig = z.infer<typeof webhookConfigSchema>;

/**
 * Update webhook configuration request schema
 * Matches: dto.UpdateWebhookConfigRequest
 */
export const updateWebhookConfigRequestSchema = z.object({
  enabled: z.boolean(),
  url: z.string().url().or(z.literal("")),
  events: z.array(z.string()),
  ignore_groups: z.boolean(),
  ignore_broadcasts: z.boolean(),
  ignore_channels: z.boolean(),
});

export type UpdateWebhookConfigRequest = z.infer<
  typeof updateWebhookConfigRequestSchema
>;
