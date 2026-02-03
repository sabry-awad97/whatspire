/**
 * Receipt-related schemas
 * Matches backend: dto/receipt.go
 */
import { z } from "zod";
import { jidSchema, sessionIdSchema, timestampSchema } from "./common";

/**
 * Send receipt request schema
 * Matches: dto.SendReceiptRequest
 */
export const sendReceiptRequestSchema = z.object({
  session_id: sessionIdSchema,
  chat_jid: jidSchema,
  message_ids: z.array(z.string().min(1)).min(1),
});

export type SendReceiptRequest = z.infer<typeof sendReceiptRequestSchema>;

/**
 * Receipt response schema
 * Matches: dto.ReceiptResponse
 */
export const receiptResponseSchema = z.object({
  processed_count: z.number().int().min(0),
  timestamp: timestampSchema,
});

export type ReceiptResponse = z.infer<typeof receiptResponseSchema>;
