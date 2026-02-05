/**
 * @whatspire/schema
 *
 * Centralized Zod schemas for the Whatspire WhatsApp API
 * All schemas match the backend Go DTOs to ensure type safety and consistency
 */

// Common schemas and utilities
export * from "./common";

// Session schemas
export * from "./session";

// Message schemas
export * from "./message";

// Contact schemas
export * from "./contact";

// Group schemas
export * from "./group";

// Presence schemas
export * from "./presence";

// Reaction schemas
export * from "./reaction";

// Receipt schemas
export * from "./receipt";

// Event schemas
export * from "./event";

// API Key schemas
export * from "./apikey";
