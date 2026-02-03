import axios, { AxiosError, type AxiosInstance } from "axios";

// ============================================================================
// Types
// ============================================================================

export interface ApiError {
  code: string;
  message: string;
  details?: Record<string, unknown>;
}

export interface ApiErrorResponse {
  error: ApiError;
}

export interface Session {
  id: string;
  status: "pending" | "connected" | "disconnected" | "error";
  jid?: string;
  created_at: string;
  updated_at?: string;
}

export interface RegisterSessionRequest {
  session_id: string;
  name: string;
}

export interface SendMessageRequest {
  session_id: string;
  to: string;
  type: "text" | "image" | "document" | "video" | "audio";
  content: {
    text?: string;
    url?: string;
    caption?: string;
    filename?: string;
  };
}

export interface MessageResponse {
  id: string;
  session_id: string;
  to: string;
  status: "queued" | "sent" | "delivered" | "read" | "failed";
  timestamp: string;
}

export interface Contact {
  jid: string;
  name: string;
  push_name?: string;
  status?: string;
  picture_url?: string;
}

export interface Chat {
  jid: string;
  name: string;
  last_message_at?: string;
  unread_count: number;
}

export interface Group {
  jid: string;
  name: string;
  owner_jid: string;
  participant_count: number;
}

export interface CheckPhoneResponse {
  phone: string;
  on_whatsapp: boolean;
  jid?: string;
}

export interface HealthResponse {
  status: "healthy" | "unhealthy";
  timestamp: string;
}

export interface ReadyResponse {
  status: "ready" | "not_ready";
  checks: Record<string, string>;
}

// ============================================================================
// Retry Configuration
// ============================================================================

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
// API Client Class
// ============================================================================

export class ApiClient {
  private client: AxiosInstance;
  private retryConfig: RetryConfig;

  constructor(
    baseURL: string = import.meta.env.VITE_SERVER_URL ||
      "http://localhost:8080",
    apiKey?: string,
    retryConfig: Partial<RetryConfig> = {},
  ) {
    this.retryConfig = { ...DEFAULT_RETRY_CONFIG, ...retryConfig };

    this.client = axios.create({
      baseURL,
      headers: {
        "Content-Type": "application/json",
        ...(apiKey && { "X-API-Key": apiKey }),
      },
      timeout: 30000,
    });

    // Response interceptor for error handling
    this.client.interceptors.response.use(
      (response) => response,
      (error: AxiosError<ApiErrorResponse>) => {
        if (error.response?.data?.error) {
          throw new ApiClientError(
            error.response.data.error.message,
            error.response.data.error.code,
            error.response.status,
            error.response.data.error.details,
          );
        }
        throw error;
      },
    );
  }

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

    // Retry on network errors
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

  async health(): Promise<HealthResponse> {
    return this.executeWithRetry(async () => {
      const response = await this.client.get<HealthResponse>("/health");
      return response.data;
    });
  }

  async ready(): Promise<ReadyResponse> {
    return this.executeWithRetry(async () => {
      const response = await this.client.get<ReadyResponse>("/ready");
      return response.data;
    });
  }

  // ==========================================================================
  // Session Management
  // ==========================================================================

  async registerSession(data: RegisterSessionRequest): Promise<Session> {
    return this.executeWithRetry(async () => {
      const response = await this.client.post<Session>(
        "/api/internal/sessions/register",
        data,
      );
      return response.data;
    });
  }

  async reconnectSession(sessionId: string): Promise<Session> {
    return this.executeWithRetry(async () => {
      const response = await this.client.post<Session>(
        `/api/internal/sessions/${sessionId}/reconnect`,
      );
      return response.data;
    });
  }

  async disconnectSession(sessionId: string): Promise<Session> {
    return this.executeWithRetry(async () => {
      const response = await this.client.post<Session>(
        `/api/internal/sessions/${sessionId}/disconnect`,
      );
      return response.data;
    });
  }

  async unregisterSession(sessionId: string): Promise<void> {
    return this.executeWithRetry(async () => {
      await this.client.post(`/api/internal/sessions/${sessionId}/unregister`);
    });
  }

  // ==========================================================================
  // Messages
  // ==========================================================================

  async sendMessage(data: SendMessageRequest): Promise<MessageResponse> {
    return this.executeWithRetry(async () => {
      const response = await this.client.post<MessageResponse>(
        "/api/messages",
        data,
      );
      return response.data;
    });
  }

  // ==========================================================================
  // Contacts
  // ==========================================================================

  async checkPhone(
    sessionId: string,
    phone: string,
  ): Promise<CheckPhoneResponse> {
    return this.executeWithRetry(async () => {
      const response = await this.client.get<CheckPhoneResponse>(
        "/api/contacts/check",
        {
          params: { session_id: sessionId, phone },
        },
      );
      return response.data;
    });
  }

  async getContactProfile(sessionId: string, jid: string): Promise<Contact> {
    return this.executeWithRetry(async () => {
      const response = await this.client.get<Contact>(
        `/api/contacts/${jid}/profile`,
        {
          params: { session_id: sessionId },
        },
      );
      return response.data;
    });
  }

  async getContacts(sessionId: string): Promise<{ contacts: Contact[] }> {
    return this.executeWithRetry(async () => {
      const response = await this.client.get<{ contacts: Contact[] }>(
        `/api/sessions/${sessionId}/contacts`,
      );
      return response.data;
    });
  }

  async getChats(sessionId: string): Promise<{ chats: Chat[] }> {
    return this.executeWithRetry(async () => {
      const response = await this.client.get<{ chats: Chat[] }>(
        `/api/sessions/${sessionId}/chats`,
      );
      return response.data;
    });
  }

  // ==========================================================================
  // Groups
  // ==========================================================================

  async syncGroups(sessionId: string): Promise<{ groups: Group[] }> {
    return this.executeWithRetry(async () => {
      const response = await this.client.post<{ groups: Group[] }>(
        `/api/sessions/${sessionId}/groups/sync`,
      );
      return response.data;
    });
  }
}

// ============================================================================
// Custom Error Class
// ============================================================================

export class ApiClientError extends Error {
  constructor(
    message: string,
    public code: string,
    public status?: number,
    public details?: Record<string, unknown>,
  ) {
    super(message);
    this.name = "ApiClientError";
  }
}

// ============================================================================
// Default Export
// ============================================================================

export const apiClient = new ApiClient();
