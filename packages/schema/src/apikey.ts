import { z } from "zod";

/**
 * API Key role enum
 */
export const apiKeyRoleSchema = z.enum(["read", "write", "admin"]);
export type APIKeyRole = z.infer<typeof apiKeyRoleSchema>;

/**
 * API Key status enum
 */
export const apiKeyStatusSchema = z.enum(["active", "revoked"]);
export type APIKeyStatus = z.infer<typeof apiKeyStatusSchema>;

/**
 * Schema for creating a new API key
 */
export const createAPIKeySchema = z.object({
  role: apiKeyRoleSchema,
  description: z.string().max(500).optional(),
});
export type CreateAPIKeyRequest = z.infer<typeof createAPIKeySchema>;

/**
 * Schema for revoking an API key
 */
export const revokeAPIKeySchema = z.object({
  reason: z.string().max(500).optional(),
});
export type RevokeAPIKeyRequest = z.infer<typeof revokeAPIKeySchema>;

/**
 * Schema for API key response
 */
export const apiKeySchema = z.object({
  id: z.string().uuid(),
  masked_key: z.string(),
  role: apiKeyRoleSchema,
  description: z.string().optional().nullable(),
  created_at: z.string().datetime(),
  last_used_at: z.string().datetime().optional().nullable(),
  is_active: z.boolean(),
  revoked_at: z.string().datetime().optional().nullable(),
  revoked_by: z.string().optional().nullable(),
  revocation_reason: z.string().optional().nullable(),
});
export type APIKey = z.infer<typeof apiKeySchema>;

/**
 * Schema for create API key response
 */
export const createAPIKeyResponseSchema = z.object({
  api_key: apiKeySchema,
  plain_key: z.string(), // Only returned once during creation
});
export type CreateAPIKeyResponse = z.infer<typeof createAPIKeyResponseSchema>;

/**
 * Schema for list API keys request
 */
export const listAPIKeysRequestSchema = z.object({
  page: z.number().int().min(1).optional(),
  limit: z.number().int().min(1).max(100).optional(),
  role: apiKeyRoleSchema.optional(),
  status: apiKeyStatusSchema.optional(),
});
export type ListAPIKeysRequest = z.infer<typeof listAPIKeysRequestSchema>;

/**
 * Schema for pagination info
 */
export const paginationInfoSchema = z.object({
  page: z.number().int(),
  limit: z.number().int(),
  total: z.number().int(),
  total_pages: z.number().int(),
});
export type PaginationInfo = z.infer<typeof paginationInfoSchema>;

/**
 * Schema for list API keys response
 */
export const listAPIKeysResponseSchema = z.object({
  api_keys: z.array(apiKeySchema),
  pagination: paginationInfoSchema,
});
export type ListAPIKeysResponse = z.infer<typeof listAPIKeysResponseSchema>;

/**
 * Schema for revoke API key response
 */
export const revokeAPIKeyResponseSchema = z.object({
  id: z.string().uuid(),
  revoked_at: z.string().datetime(),
  revoked_by: z.string(),
});
export type RevokeAPIKeyResponse = z.infer<typeof revokeAPIKeyResponseSchema>;

/**
 * Schema for usage statistics
 */
export const usageStatsSchema = z.object({
  total_requests: z.number().int(),
  last_7_days: z.number().int(),
});
export type UsageStats = z.infer<typeof usageStatsSchema>;

/**
 * Schema for API key details response
 */
export const apiKeyDetailsResponseSchema = z.object({
  api_key: apiKeySchema,
  usage_stats: usageStatsSchema,
});
export type APIKeyDetailsResponse = z.infer<typeof apiKeyDetailsResponseSchema>;
