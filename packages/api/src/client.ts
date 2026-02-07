/**
 * API Client with type-safe schema validation
 * Uses @whatspire/schema for runtime validation and type inference
 */
import axios, { AxiosError, type AxiosInstance } from "axios";
import { z } from "zod";
import {
  apiResponseSchema,
  sessionSchema,
  sessionListSchema,
  createSessionRequestSchema,
  sendMessageRequestSchema,
  contactSchema,
  contactListSchema,
  chatListSchema,
  checkPhoneRequestSchema,
  groupListSchema,
  sendPresenceRequestSchema,
  presenceResponseSchema,
  sendReactionRequestSchema,
  removeReactionRequestSchema,
  reactionResponseSchema,
  sendReceiptRequestSchema,
  receiptResponseSchema,
  queryEventsResponseSchema,
  createAPIKeySchema,
  createAPIKeyResponseSchema,
  revokeAPIKeySchema,
  revokeAPIKeyResponseSchema,
  listAPIKeysRequestSchema,
  listAPIKeysResponseSchema,
  apiKeyDetailsResponseSchema,
  webhookConfigSchema,
  updateWebhookConfigRequestSchema,
  type ApiResponse,
  type Session,
  type CreateSessionRequest,
  type SendMessageRequest,
  type Contact,
  type ContactList,
  type Chat,
  type ChatList,
  type CheckPhoneRequest,
  type GetProfileRequest,
  type Group,
  type GroupList,
  type SendPresenceRequest,
  type PresenceResponse,
  type SendReactionRequest,
  type RemoveReactionRequest,
  type ReactionResponse,
  type SendReceiptRequest,
  type ReceiptResponse,
  type QueryEventsRequest,
  type QueryEventsResponse,
  type CreateAPIKeyRequest,
  type CreateAPIKeyResponse,
  type RevokeAPIKeyRequest,
  type RevokeAPIKeyResponse,
  type ListAPIKeysRequest,
  type ListAPIKeysResponse,
  type APIKeyDetailsResponse,
  type WebhookConfig,
  type UpdateWebhookConfigRequest,
} from "@whatspire/schema";

// ============================================================================
// Configuration Types
// ============================================================================

export interface ApiClientConfig {
  baseURL?: string;
  apiKey?: string;
  timeout?: number;
  retryConfig?: Partial<RetryConfig>;
}

export interface RetryConfig {
  maxAttempts: number;
  initialDelayMs: number;
  maxDelayMs: number;
  backoffMultiplier: number;
}

const DEFAULT_RETRY_CONFIG: RetryConfig = {
  maxAttempts: 3,
  initialDelayMs: 1000,
  maxDelayMs: 10000,
  backoffMultiplier: 2,
};

// ============================================================================
// Custom Error Class
// ============================================================================

export class ApiClientError extends Error {
  constructor(
    message: string,
    public code: string,
    public status?: number,
    public details?: Record<string, string>,
  ) {
    super(message);
    this.name = "ApiClientError";
    Object.setPrototypeOf(this, ApiClientError.prototype);
  }

  toJSON() {
    return {
      name: this.name,
      message: this.message,
      code: this.code,
      status: this.status,
      details: this.details,
    };
  }
}

// ============================================================================
// API Client Class
// ============================================================================

export class ApiClient {
  private client: AxiosInstance;
  private retryConfig: RetryConfig;

  constructor(config: ApiClientConfig = {}) {
    const {
      baseURL = "http://localhost:8080",
      apiKey,
      timeout = 30000,
      retryConfig = {},
    } = config;

    this.retryConfig = { ...DEFAULT_RETRY_CONFIG, ...retryConfig };

    this.client = axios.create({
      baseURL,
      timeout,
      headers: {
        "Content-Type": "application/json",
        ...(apiKey && { "X-API-Key": apiKey }),
      },
    });

    // Response interceptor for error handling
    this.client.interceptors.response.use(
      (response) => response,
      (error: AxiosError<ApiResponse<never>>) => {
        if (error.response?.data?.error) {
          const apiError = error.response.data.error;
          throw new ApiClientError(
            apiError.message,
            apiError.code,
            error.response.status,
            apiError.details,
          );
        }
        throw error;
      },
    );
  }

  // ==========================================================================
  // Private Helper Methods
  // ==========================================================================

  /**
   * Execute request with exponential backoff retry logic
   */
  private async executeWithRetry<T>(
    requestFn: () => Promise<T>,
    attempt: number = 1,
  ): Promise<T> {
    try {
      return await requestFn();
    } catch (error) {
      const shouldRetry = this.shouldRetry(error, attempt);

      if (!shouldRetry) {
        throw error;
      }

      const delay = this.calculateDelay(attempt);
      await this.sleep(delay);

      return this.executeWithRetry(requestFn, attempt + 1);
    }
  }

  /**
   * Determine if request should be retried
   */
  private shouldRetry(error: unknown, attempt: number): boolean {
    if (attempt >= this.retryConfig.maxAttempts) {
      return false;
    }

    if (axios.isAxiosError(error)) {
      // Don't retry on 4xx errors (except 429 rate limit)
      if (error.response?.status) {
        const status = error.response.status;
        if (status >= 400 && status < 500 && status !== 429) {
          return false;
        }
      }

      // Retry on network errors, timeouts, and 5xx errors
      return (
        !error.response ||
        error.code === "ECONNABORTED" ||
        error.code === "ETIMEDOUT" ||
        (error.response.status >= 500 && error.response.status < 600)
      );
    }

    return false;
  }

  /**
   * Calculate exponential backoff delay
   */
  private calculateDelay(attempt: number): number {
    const delay =
      this.retryConfig.initialDelayMs *
      Math.pow(this.retryConfig.backoffMultiplier, attempt - 1);
    return Math.min(delay, this.retryConfig.maxDelayMs);
  }

  /**
   * Sleep utility
   */
  private sleep(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms));
  }

  // ==========================================================================
  // Health Endpoints
  // ==========================================================================

  /**
   * Check API health status
   */
  async health(): Promise<{ status: string; timestamp: string }> {
    return this.executeWithRetry(async () => {
      const response = await this.client.get<{
        status: string;
        timestamp: string;
      }>("/health");
      return response.data;
    });
  }

  /**
   * Check API readiness
   */
  async ready(): Promise<{ status: string; checks: Record<string, string> }> {
    return this.executeWithRetry(async () => {
      const response = await this.client.get<{
        status: string;
        checks: Record<string, string>;
      }>("/ready");
      return response.data;
    });
  }

  // ==========================================================================
  // Session Management
  // ==========================================================================

  /**
   * Create a new WhatsApp session
   */
  async createSession(data: CreateSessionRequest): Promise<Session> {
    return this.executeWithRetry(async () => {
      // Validate request
      const validatedRequest = createSessionRequestSchema.parse(data);

      // Make request
      const response = await this.client.post<ApiResponse<Session>>(
        "/api/sessions",
        validatedRequest,
      );

      // Validate response
      const validatedResponse = apiResponseSchema(sessionSchema).parse(
        response.data,
      );

      if (!validatedResponse.success || !validatedResponse.data) {
        throw new ApiClientError(
          validatedResponse.error?.message || "Failed to create session",
          validatedResponse.error?.code || "CREATE_SESSION_FAILED",
          response.status,
          validatedResponse.error?.details,
        );
      }

      return validatedResponse.data;
    });
  }

  /**
   * List all sessions
   */
  async listSessions(): Promise<Session[]> {
    return this.executeWithRetry(async () => {
      const response =
        await this.client.get<ApiResponse<{ sessions: Session[] }>>(
          "/api/sessions",
        );

      const validatedResponse = apiResponseSchema(
        z.object({ sessions: sessionListSchema }),
      ).parse(response.data);

      if (!validatedResponse.success || !validatedResponse.data) {
        throw new ApiClientError(
          validatedResponse.error?.message || "Failed to list sessions",
          validatedResponse.error?.code || "LIST_SESSIONS_FAILED",
          response.status,
          validatedResponse.error?.details,
        );
      }

      return validatedResponse.data.sessions;
    });
  }

  /**
   * Get session by ID
   */
  async getSession(sessionId: string): Promise<Session> {
    return this.executeWithRetry(async () => {
      const response = await this.client.get<ApiResponse<Session>>(
        `/api/sessions/${sessionId}`,
      );

      const validatedResponse = apiResponseSchema(sessionSchema).parse(
        response.data,
      );

      if (!validatedResponse.success || !validatedResponse.data) {
        throw new ApiClientError(
          validatedResponse.error?.message || "Failed to get session",
          validatedResponse.error?.code || "GET_SESSION_FAILED",
          response.status,
          validatedResponse.error?.details,
        );
      }

      return validatedResponse.data;
    });
  }

  /**
   * Delete session
   */
  async deleteSession(sessionId: string): Promise<void> {
    await this.executeWithRetry(async () => {
      await this.client.delete(`/api/sessions/${sessionId}`);
    });
  }

  /**
   * Update session settings
   */
  async updateSession(
    sessionId: string,
    data: {
      name?: string;
      config?: {
        account_protection?: boolean;
        message_logging?: boolean;
        read_messages?: boolean;
        auto_reject_calls?: boolean;
        always_online?: boolean;
        ignore_groups?: boolean;
        ignore_broadcasts?: boolean;
        ignore_channels?: boolean;
      };
    },
  ): Promise<Session> {
    return this.executeWithRetry(async () => {
      const response = await this.client.patch<ApiResponse<Session>>(
        `/api/sessions/${sessionId}`,
        data,
      );

      const validatedResponse = apiResponseSchema(sessionSchema).parse(
        response.data,
      );

      if (!validatedResponse.success || !validatedResponse.data) {
        throw new ApiClientError(
          validatedResponse.error?.message || "Failed to update session",
          validatedResponse.error?.code || "UPDATE_SESSION_FAILED",
          response.status,
          validatedResponse.error?.details,
        );
      }

      return validatedResponse.data;
    });
  }

  /**
   * Reconnect session
   */
  async reconnectSession(sessionId: string): Promise<void> {
    await this.executeWithRetry(async () => {
      const response = await this.client.post<
        ApiResponse<{ success: boolean; message: string }>
      >(`/api/internal/sessions/${sessionId}/reconnect`);

      if (!response.data.success) {
        throw new ApiClientError(
          response.data.error?.message || "Failed to reconnect session",
          response.data.error?.code || "RECONNECT_SESSION_FAILED",
          response.status,
          response.data.error?.details,
        );
      }
    });
  }

  /**
   * Disconnect session
   */
  async disconnectSession(sessionId: string): Promise<void> {
    await this.executeWithRetry(async () => {
      const response = await this.client.post<
        ApiResponse<{ success: boolean; message: string }>
      >(`/api/internal/sessions/${sessionId}/disconnect`);

      if (!response.data.success) {
        throw new ApiClientError(
          response.data.error?.message || "Failed to disconnect session",
          response.data.error?.code || "DISCONNECT_SESSION_FAILED",
          response.status,
          response.data.error?.details,
        );
      }
    });
  }

  // ==========================================================================
  // Webhook Configuration
  // ==========================================================================

  /**
   * Get webhook configuration for a session
   */
  async getWebhookConfig(sessionId: string): Promise<WebhookConfig> {
    return this.executeWithRetry(async () => {
      const response = await this.client.get<ApiResponse<WebhookConfig>>(
        `/api/sessions/${sessionId}/webhook`,
      );

      const validatedResponse = apiResponseSchema(webhookConfigSchema).parse(
        response.data,
      );

      if (!validatedResponse.success || !validatedResponse.data) {
        throw new ApiClientError(
          validatedResponse.error?.message || "Failed to get webhook config",
          validatedResponse.error?.code || "GET_WEBHOOK_CONFIG_FAILED",
          response.status,
          validatedResponse.error?.details,
        );
      }

      return validatedResponse.data;
    });
  }

  /**
   * Update webhook configuration for a session
   */
  async updateWebhookConfig(
    sessionId: string,
    data: UpdateWebhookConfigRequest,
  ): Promise<WebhookConfig> {
    return this.executeWithRetry(async () => {
      // Validate request
      const validatedRequest = updateWebhookConfigRequestSchema.parse(data);

      const response = await this.client.put<ApiResponse<WebhookConfig>>(
        `/api/sessions/${sessionId}/webhook`,
        validatedRequest,
      );

      const validatedResponse = apiResponseSchema(webhookConfigSchema).parse(
        response.data,
      );

      if (!validatedResponse.success || !validatedResponse.data) {
        throw new ApiClientError(
          validatedResponse.error?.message || "Failed to update webhook config",
          validatedResponse.error?.code || "UPDATE_WEBHOOK_CONFIG_FAILED",
          response.status,
          validatedResponse.error?.details,
        );
      }

      return validatedResponse.data;
    });
  }

  /**
   * Rotate webhook secret for a session
   */
  async rotateWebhookSecret(sessionId: string): Promise<WebhookConfig> {
    return this.executeWithRetry(async () => {
      const response = await this.client.post<ApiResponse<WebhookConfig>>(
        `/api/sessions/${sessionId}/webhook/rotate-secret`,
      );

      const validatedResponse = apiResponseSchema(webhookConfigSchema).parse(
        response.data,
      );

      if (!validatedResponse.success || !validatedResponse.data) {
        throw new ApiClientError(
          validatedResponse.error?.message || "Failed to rotate webhook secret",
          validatedResponse.error?.code || "ROTATE_WEBHOOK_SECRET_FAILED",
          response.status,
          validatedResponse.error?.details,
        );
      }

      return validatedResponse.data;
    });
  }

  /**
   * Delete webhook configuration for a session
   */
  async deleteWebhookConfig(sessionId: string): Promise<void> {
    await this.executeWithRetry(async () => {
      await this.client.delete(`/api/sessions/${sessionId}/webhook`);
    });
  }

  // ==========================================================================
  // Messages
  // ==========================================================================

  /**
   * Send a WhatsApp message
   */
  async sendMessage(
    data: SendMessageRequest,
  ): Promise<{ id: string; status: string }> {
    return this.executeWithRetry(async () => {
      // Validate request
      const validatedRequest = sendMessageRequestSchema.parse(data);

      const response = await this.client.post<
        ApiResponse<{ id: string; status: string }>
      >("/api/messages", validatedRequest);

      if (!response.data.success || !response.data.data) {
        throw new ApiClientError(
          response.data.error?.message || "Failed to send message",
          response.data.error?.code || "SEND_MESSAGE_FAILED",
          response.status,
          response.data.error?.details,
        );
      }

      return response.data.data;
    });
  }

  // ==========================================================================
  // Contacts
  // ==========================================================================

  /**
   * Check if phone number is on WhatsApp
   */
  async checkPhone(
    data: CheckPhoneRequest,
  ): Promise<{ phone: string; on_whatsapp: boolean; jid?: string }> {
    return this.executeWithRetry(async () => {
      checkPhoneRequestSchema.parse(data);

      const response = await this.client.get<
        ApiResponse<{ phone: string; on_whatsapp: boolean; jid?: string }>
      >("/api/contacts/check", { params: data });

      if (!response.data.success || !response.data.data) {
        throw new ApiClientError(
          response.data.error?.message || "Failed to check phone",
          response.data.error?.code || "CHECK_PHONE_FAILED",
          response.status,
          response.data.error?.details,
        );
      }

      return response.data.data;
    });
  }

  /**
   * Get contact profile
   */
  async getContactProfile(data: GetProfileRequest): Promise<Contact> {
    return this.executeWithRetry(async () => {
      const response = await this.client.get<ApiResponse<Contact>>(
        `/api/contacts/${data.jid}/profile`,
        { params: { session_id: data.session_id } },
      );

      const validatedResponse = apiResponseSchema(contactSchema).parse(
        response.data,
      );

      if (!validatedResponse.success || !validatedResponse.data) {
        throw new ApiClientError(
          validatedResponse.error?.message || "Failed to get contact profile",
          validatedResponse.error?.code || "GET_PROFILE_FAILED",
          response.status,
          validatedResponse.error?.details,
        );
      }

      return validatedResponse.data;
    });
  }

  /**
   * Get all contacts for a session
   */
  async getContacts(sessionId: string): Promise<Contact[]> {
    return this.executeWithRetry(async () => {
      const response = await this.client.get<ApiResponse<ContactList>>(
        `/api/sessions/${sessionId}/contacts`,
      );

      const validatedResponse = apiResponseSchema(contactListSchema).parse(
        response.data,
      );

      if (!validatedResponse.success || !validatedResponse.data) {
        throw new ApiClientError(
          validatedResponse.error?.message || "Failed to get contacts",
          validatedResponse.error?.code || "GET_CONTACTS_FAILED",
          response.status,
          validatedResponse.error?.details,
        );
      }

      return validatedResponse.data.contacts;
    });
  }

  /**
   * Get all chats for a session
   */
  async getChats(sessionId: string): Promise<Chat[]> {
    return this.executeWithRetry(async () => {
      const response = await this.client.get<ApiResponse<ChatList>>(
        `/api/sessions/${sessionId}/chats`,
      );

      const validatedResponse = apiResponseSchema(chatListSchema).parse(
        response.data,
      );

      if (!validatedResponse.success || !validatedResponse.data) {
        throw new ApiClientError(
          validatedResponse.error?.message || "Failed to get chats",
          validatedResponse.error?.code || "GET_CHATS_FAILED",
          response.status,
          validatedResponse.error?.details,
        );
      }

      return validatedResponse.data.chats;
    });
  }

  // ==========================================================================
  // Groups
  // ==========================================================================

  /**
   * Sync groups for a session
   */
  async syncGroups(sessionId: string): Promise<Group[]> {
    return this.executeWithRetry(async () => {
      const response = await this.client.post<ApiResponse<GroupList>>(
        `/api/sessions/${sessionId}/groups/sync`,
      );

      const validatedResponse = apiResponseSchema(groupListSchema).parse(
        response.data,
      );

      if (!validatedResponse.success || !validatedResponse.data) {
        throw new ApiClientError(
          validatedResponse.error?.message || "Failed to sync groups",
          validatedResponse.error?.code || "SYNC_GROUPS_FAILED",
          response.status,
          validatedResponse.error?.details,
        );
      }

      return validatedResponse.data;
    });
  }

  // ==========================================================================
  // Presence
  // ==========================================================================

  /**
   * Send presence update
   */
  async sendPresence(data: SendPresenceRequest): Promise<PresenceResponse> {
    return this.executeWithRetry(async () => {
      const validatedRequest = sendPresenceRequestSchema.parse(data);

      const response = await this.client.post<ApiResponse<PresenceResponse>>(
        "/api/presence",
        validatedRequest,
      );

      const validatedResponse = apiResponseSchema(presenceResponseSchema).parse(
        response.data,
      );

      if (!validatedResponse.success || !validatedResponse.data) {
        throw new ApiClientError(
          validatedResponse.error?.message || "Failed to send presence",
          validatedResponse.error?.code || "SEND_PRESENCE_FAILED",
          response.status,
          validatedResponse.error?.details,
        );
      }

      return validatedResponse.data;
    });
  }

  // ==========================================================================
  // Reactions
  // ==========================================================================

  /**
   * Send reaction to a message
   */
  async sendReaction(data: SendReactionRequest): Promise<ReactionResponse> {
    return this.executeWithRetry(async () => {
      const validatedRequest = sendReactionRequestSchema.parse(data);

      const response = await this.client.post<ApiResponse<ReactionResponse>>(
        "/api/reactions",
        validatedRequest,
      );

      const validatedResponse = apiResponseSchema(reactionResponseSchema).parse(
        response.data,
      );

      if (!validatedResponse.success || !validatedResponse.data) {
        throw new ApiClientError(
          validatedResponse.error?.message || "Failed to send reaction",
          validatedResponse.error?.code || "SEND_REACTION_FAILED",
          response.status,
          validatedResponse.error?.details,
        );
      }

      return validatedResponse.data;
    });
  }

  /**
   * Remove reaction from a message
   */
  async removeReaction(data: RemoveReactionRequest): Promise<ReactionResponse> {
    return this.executeWithRetry(async () => {
      const validatedRequest = removeReactionRequestSchema.parse(data);

      const response = await this.client.post<ApiResponse<ReactionResponse>>(
        "/api/reactions/remove",
        validatedRequest,
      );

      const validatedResponse = apiResponseSchema(reactionResponseSchema).parse(
        response.data,
      );

      if (!validatedResponse.success || !validatedResponse.data) {
        throw new ApiClientError(
          validatedResponse.error?.message || "Failed to remove reaction",
          validatedResponse.error?.code || "REMOVE_REACTION_FAILED",
          response.status,
          validatedResponse.error?.details,
        );
      }

      return validatedResponse.data;
    });
  }

  // ==========================================================================
  // Receipts
  // ==========================================================================

  /**
   * Send read receipts for messages
   */
  async sendReceipt(data: SendReceiptRequest): Promise<ReceiptResponse> {
    return this.executeWithRetry(async () => {
      const validatedRequest = sendReceiptRequestSchema.parse(data);

      const response = await this.client.post<ApiResponse<ReceiptResponse>>(
        "/api/receipts",
        validatedRequest,
      );

      const validatedResponse = apiResponseSchema(receiptResponseSchema).parse(
        response.data,
      );

      if (!validatedResponse.success || !validatedResponse.data) {
        throw new ApiClientError(
          validatedResponse.error?.message || "Failed to send receipt",
          validatedResponse.error?.code || "SEND_RECEIPT_FAILED",
          response.status,
          validatedResponse.error?.details,
        );
      }

      return validatedResponse.data;
    });
  }

  // ==========================================================================
  // Events
  // ==========================================================================

  /**
   * Query events with filters
   */
  async queryEvents(data: QueryEventsRequest): Promise<QueryEventsResponse> {
    return this.executeWithRetry(async () => {
      const response = await this.client.get<ApiResponse<QueryEventsResponse>>(
        "/api/events",
        { params: data },
      );

      const validatedResponse = apiResponseSchema(
        queryEventsResponseSchema,
      ).parse(response.data);

      if (!validatedResponse.success || !validatedResponse.data) {
        throw new ApiClientError(
          validatedResponse.error?.message || "Failed to query events",
          validatedResponse.error?.code || "QUERY_EVENTS_FAILED",
          response.status,
          validatedResponse.error?.details,
        );
      }

      return validatedResponse.data;
    });
  }

  // ==========================================================================
  // API Keys
  // ==========================================================================

  /**
   * Create a new API key
   * @param data - API key creation request with role and optional description
   * @returns Created API key with plain key (only shown once)
   */
  async createAPIKey(data: CreateAPIKeyRequest): Promise<CreateAPIKeyResponse> {
    return this.executeWithRetry(async () => {
      // Validate request
      const validatedRequest = createAPIKeySchema.parse(data);

      // Make request
      const response = await this.client.post<
        ApiResponse<CreateAPIKeyResponse>
      >("/api/apikeys", validatedRequest);

      // Validate response
      const validatedResponse = apiResponseSchema(
        createAPIKeyResponseSchema,
      ).parse(response.data);

      if (!validatedResponse.success || !validatedResponse.data) {
        throw new ApiClientError(
          validatedResponse.error?.message || "Failed to create API key",
          validatedResponse.error?.code || "CREATE_API_KEY_FAILED",
          response.status,
          validatedResponse.error?.details,
        );
      }

      return validatedResponse.data;
    });
  }

  /**
   * Revoke an existing API key
   * @param id - API key ID to revoke
   * @param reason - Optional reason for revocation (for audit trail)
   * @returns Revocation details including timestamp and revoker
   * @throws ApiClientError if key not found or already revoked
   */
  async revokeAPIKey(
    id: string,
    reason?: string,
  ): Promise<RevokeAPIKeyResponse> {
    return this.executeWithRetry(async () => {
      // Prepare request body
      const requestBody: RevokeAPIKeyRequest = reason ? { reason } : {};

      // Validate request if reason is provided
      if (reason) {
        revokeAPIKeySchema.parse(requestBody);
      }

      // Make request
      const response = await this.client.delete<
        ApiResponse<RevokeAPIKeyResponse>
      >(`/api/apikeys/${id}`, { data: requestBody });

      // Validate response
      const validatedResponse = apiResponseSchema(
        revokeAPIKeyResponseSchema,
      ).parse(response.data);

      if (!validatedResponse.success || !validatedResponse.data) {
        throw new ApiClientError(
          validatedResponse.error?.message || "Failed to revoke API key",
          validatedResponse.error?.code || "REVOKE_API_KEY_FAILED",
          response.status,
          validatedResponse.error?.details,
        );
      }

      return validatedResponse.data;
    });
  }

  /**
   * List API keys with optional filtering and pagination
   * @param params - Optional filters and pagination parameters
   * @returns Paginated list of API keys with metadata
   * @throws ApiClientError if request fails
   */
  async listAPIKeys(params?: ListAPIKeysRequest): Promise<ListAPIKeysResponse> {
    return this.executeWithRetry(async () => {
      // Validate request parameters if provided
      if (params) {
        listAPIKeysRequestSchema.parse(params);
      }

      // Make request
      const response = await this.client.get<ApiResponse<ListAPIKeysResponse>>(
        "/api/apikeys",
        { params },
      );

      // Validate response
      const validatedResponse = apiResponseSchema(
        listAPIKeysResponseSchema,
      ).parse(response.data);

      if (!validatedResponse.success || !validatedResponse.data) {
        throw new ApiClientError(
          validatedResponse.error?.message || "Failed to list API keys",
          validatedResponse.error?.code || "LIST_API_KEYS_FAILED",
          response.status,
          validatedResponse.error?.details,
        );
      }

      return validatedResponse.data;
    });
  }

  /**
   * Get detailed information about a specific API key
   * @param id - API key ID to retrieve details for
   * @returns API key details including metadata and usage statistics
   * @throws ApiClientError if key not found or request fails
   */
  async getAPIKeyDetails(id: string): Promise<APIKeyDetailsResponse> {
    return this.executeWithRetry(async () => {
      // Make request
      const response = await this.client.get<
        ApiResponse<APIKeyDetailsResponse>
      >(`/api/apikeys/${id}`);

      // Validate response
      const validatedResponse = apiResponseSchema(
        apiKeyDetailsResponseSchema,
      ).parse(response.data);

      if (!validatedResponse.success || !validatedResponse.data) {
        throw new ApiClientError(
          validatedResponse.error?.message || "Failed to get API key details",
          validatedResponse.error?.code || "GET_API_KEY_DETAILS_FAILED",
          response.status,
          validatedResponse.error?.details,
        );
      }

      return validatedResponse.data;
    });
  }
}
