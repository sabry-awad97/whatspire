/**
 * Message-related schemas
 * Matches backend: dto/request.go (SendMessageRequest, SendMessageContentInput)
 */
import { z } from "zod";
import {
  sessionIdSchema,
  messageTypeSchema,
  phoneNumberSchema,
  urlSchema,
} from "./common";

/**
 * Message content input schema
 * Matches: dto.SendMessageContentInput
 */
export const sendMessageContentInputSchema = z.object({
  text: z.string().max(4096).optional(),
  image_url: urlSchema.optional(),
  doc_url: urlSchema.optional(),
  audio_url: urlSchema.optional(),
  video_url: urlSchema.optional(),
  caption: z.string().max(1024).optional(),
  filename: z.string().max(255).optional(),
});

export type SendMessageContentInput = z.infer<
  typeof sendMessageContentInputSchema
>;

/**
 * Send message request schema with validation
 * Matches: dto.SendMessageRequest
 */
export const sendMessageRequestSchema = z
  .object({
    session_id: sessionIdSchema,
    to: phoneNumberSchema,
    type: messageTypeSchema,
    content: sendMessageContentInputSchema,
  })
  .refine(
    (data) => {
      // Validate content based on message type
      switch (data.type) {
        case "text":
          return data.content.text != null && data.content.text.length > 0;
        case "image":
          return (
            data.content.image_url != null && data.content.image_url.length > 0
          );
        case "document":
          return (
            data.content.doc_url != null && data.content.doc_url.length > 0
          );
        case "audio":
          return (
            data.content.audio_url != null && data.content.audio_url.length > 0
          );
        case "video":
          return (
            data.content.video_url != null && data.content.video_url.length > 0
          );
        default:
          return false;
      }
    },
    {
      message: "Content must match the message type",
    },
  );

export type SendMessageRequest = z.infer<typeof sendMessageRequestSchema>;
