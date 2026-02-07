/**
 * Settings Utilities
 * Manages application settings persistence using localStorage
 */

// ============================================================================
// Types
// ============================================================================

export interface AppSettings {
  apiEndpoint: string;
  apiKey?: string;
  updatedAt: string;
}

// ============================================================================
// Constants
// ============================================================================

const SETTINGS_KEY = "whatspire_app_settings";
const DEFAULT_API_ENDPOINT = "http://localhost:8080";

// ============================================================================
// Functions
// ============================================================================

/**
 * Load settings from localStorage
 * Returns default values if no settings are saved
 */
export function loadSettings(): AppSettings {
  try {
    const stored = localStorage.getItem(SETTINGS_KEY);
    if (!stored) {
      return {
        apiEndpoint: DEFAULT_API_ENDPOINT,
        updatedAt: new Date().toISOString(),
      };
    }

    const parsed = JSON.parse(stored) as AppSettings;

    // Validate structure
    if (!parsed.apiEndpoint || !parsed.updatedAt) {
      console.warn("[Settings] Invalid settings structure, using defaults");
      return {
        apiEndpoint: DEFAULT_API_ENDPOINT,
        updatedAt: new Date().toISOString(),
      };
    }

    return parsed;
  } catch (error) {
    console.error("[Settings] Failed to load settings:", error);
    return {
      apiEndpoint: DEFAULT_API_ENDPOINT,
      updatedAt: new Date().toISOString(),
    };
  }
}

/**
 * Save settings to localStorage
 */
export function saveSettings(settings: Omit<AppSettings, "updatedAt">): void {
  try {
    const withTimestamp: AppSettings = {
      ...settings,
      updatedAt: new Date().toISOString(),
    };

    localStorage.setItem(SETTINGS_KEY, JSON.stringify(withTimestamp));
    console.log("[Settings] Settings saved successfully");
  } catch (error) {
    console.error("[Settings] Failed to save settings:", error);
    throw new Error("Failed to save settings");
  }
}

/**
 * Clear all settings from localStorage
 */
export function clearSettings(): void {
  try {
    localStorage.removeItem(SETTINGS_KEY);
    console.log("[Settings] Settings cleared");
  } catch (error) {
    console.error("[Settings] Failed to clear settings:", error);
    throw new Error("Failed to clear settings");
  }
}

/**
 * Test connection to API endpoint
 * Returns true if connection is successful, false otherwise
 */
export async function testConnection(
  apiEndpoint: string,
  apiKey?: string,
): Promise<{ success: boolean; message: string; latency?: number }> {
  const startTime = performance.now();

  try {
    // Validate URL format
    try {
      new URL(apiEndpoint);
    } catch {
      return {
        success: false,
        message: "Invalid API endpoint URL format",
      };
    }

    // Call health endpoint
    const healthUrl = `${apiEndpoint}/health`;
    const headers: HeadersInit = {
      "Content-Type": "application/json",
    };

    // Add API key if provided
    if (apiKey) {
      headers["X-API-Key"] = apiKey;
    }

    const response = await fetch(healthUrl, {
      method: "GET",
      headers,
      signal: AbortSignal.timeout(5000), // 5 second timeout
    });

    const latency = Math.round(performance.now() - startTime);

    if (!response.ok) {
      return {
        success: false,
        message: `Connection failed: ${response.status} ${response.statusText}`,
        latency,
      };
    }

    const data = await response.json();

    // Check if response has expected structure
    if (data.status === "healthy" || data.data?.status === "healthy") {
      return {
        success: true,
        message: `Connection successful! (${latency}ms)`,
        latency,
      };
    }

    return {
      success: false,
      message: "Unexpected response from health endpoint",
      latency,
    };
  } catch (error) {
    const latency = Math.round(performance.now() - startTime);

    if (error instanceof Error) {
      if (error.name === "AbortError" || error.name === "TimeoutError") {
        return {
          success: false,
          message: "Connection timeout (>5s)",
          latency,
        };
      }

      if (error.message.includes("Failed to fetch")) {
        return {
          success: false,
          message: "Cannot reach API endpoint. Check your network connection.",
          latency,
        };
      }

      return {
        success: false,
        message: `Connection error: ${error.message}`,
        latency,
      };
    }

    return {
      success: false,
      message: "Unknown connection error",
      latency,
    };
  }
}
