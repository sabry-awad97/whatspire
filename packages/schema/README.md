# @whatspire/schema

Centralized Zod schemas for the Whatspire WhatsApp API. All schemas are designed to match the backend Go DTOs exactly, ensuring type safety and consistency across the frontend and backend.

## Features

- ✅ **Type-safe**: Full TypeScript support with Zod inference
- ✅ **Backend-aligned**: Schemas match Go DTOs exactly
- ✅ **Validation**: Built-in runtime validation with Zod
- ✅ **Reusable**: Shared across all frontend applications
- ✅ **Well-documented**: Each schema includes backend reference

## Installation

```bash
# From the workspace root
bun install
```

## Usage

```typescript
import {
  sessionSchema,
  createSessionRequestSchema,
  apiResponseSchema,
} from "@whatspire/schema";

// Validate session data
const session = sessionSchema.parse(data);

// Validate API request
const request = createSessionRequestSchema.parse({
  name: "My Session",
});

// Validate API response
const response = apiResponseSchema(sessionSchema).parse(apiData);
```

## Schema Organization

### Common (`common.ts`)

- `apiResponseSchema` - Standard API response wrapper
- `apiErrorSchema` - Error response structure
- Common field schemas (JID, session ID, phone number, etc.)
- Enums (message types, session status, presence state, etc.)

### Session (`session.ts`)

- `sessionSchema` - Session response
- `createSessionRequestSchema` - Create session request
- `getSessionRequestSchema` - Get session request
- `deleteSessionRequestSchema` - Delete session request
- `startQRAuthRequestSchema` - Start QR auth request

### Message (`message.ts`)

- `sendMessageRequestSchema` - Send message request
- `sendMessageContentInputSchema` - Message content

### Contact (`contact.ts`)

- `contactSchema` - Contact response
- `contactListSchema` - Contact list response
- `chatSchema` - Chat response
- `chatListSchema` - Chat list response
- `checkPhoneRequestSchema` - Check phone request
- `getProfileRequestSchema` - Get profile request

### Group (`group.ts`)

- `groupSchema` - Group response
- `participantSchema` - Participant response
- `groupListSchema` - Group list response

### Presence (`presence.ts`)

- `sendPresenceRequestSchema` - Send presence request
- `presenceResponseSchema` - Presence response

### Reaction (`reaction.ts`)

- `sendReactionRequestSchema` - Send reaction request
- `removeReactionRequestSchema` - Remove reaction request
- `reactionResponseSchema` - Reaction response

### Receipt (`receipt.ts`)

- `sendReceiptRequestSchema` - Send receipt request
- `receiptResponseSchema` - Receipt response

### Event (`event.ts`)

- `eventSchema` - Event DTO
- `queryEventsRequestSchema` - Query events request
- `queryEventsResponseSchema` - Query events response
- `replayEventsRequestSchema` - Replay events request
- `replayEventsResponseSchema` - Replay events response

## Backend Mapping

Each schema includes a comment indicating which backend DTO it matches:

```typescript
/**
 * Session response schema
 * Matches: dto.SessionResponse
 */
export const sessionSchema = z.object({
  // ...
});
```

This ensures that any changes to backend DTOs can be easily tracked and updated in the frontend schemas.

## Type Inference

All schemas export TypeScript types using Zod's inference:

```typescript
export type Session = z.infer<typeof sessionSchema>;
```

Use these types throughout your application for consistent typing.

## Validation

Schemas include runtime validation that matches backend validation rules:

```typescript
// Phone number validation (E.164 format)
export const phoneNumberSchema = z.string().regex(/^\+?[1-9]\d{1,14}$/);

// Message type validation with content requirements
export const sendMessageRequestSchema = z
  .object({
    // ...
  })
  .refine((data) => {
    // Validate content based on message type
    switch (data.type) {
      case "text":
        return data.content.text != null && data.content.text.length > 0;
      // ...
    }
  });
```

## Development

```bash
# Type check
bun run type-check

# Build (if needed)
bun run build
```

## Contributing

When adding new schemas:

1. Match the backend DTO structure exactly
2. Include a comment referencing the backend DTO
3. Export both the schema and the inferred type
4. Add validation rules that match backend validation
5. Update this README with the new schema
6. Update `src/index.ts` to export the new schema

## Backend References

All schemas are based on the following backend files:

- `apps/server/internal/application/dto/response.go`
- `apps/server/internal/application/dto/request.go`
- `apps/server/internal/application/dto/contact.go`
- `apps/server/internal/application/dto/event.go`
- `apps/server/internal/application/dto/presence.go`
- `apps/server/internal/application/dto/reaction.go`
- `apps/server/internal/application/dto/receipt.go`
