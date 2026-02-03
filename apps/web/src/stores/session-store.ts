import { create } from "zustand";
import { devtools, persist } from "zustand/middleware";

import { apiClient, type Session } from "@/lib/api-client";

// ============================================================================
// Types
// ============================================================================

export interface QRCodeData {
  sessionId: string;
  qrCode: string; // base64 encoded image
  timestamp: number;
}

export interface SessionState {
  // Sessions
  sessions: Session[];
  activeSessions: Set<string>;

  // Loading and error state
  isLoading: boolean;
  error: string | null;

  // QR Codes
  qrCodes: Map<string, QRCodeData>;

  // WebSocket connection status
  wsConnected: boolean;

  // Actions
  fetchSessions: () => Promise<void>;
  addSession: (session: Session) => void;
  updateSession: (sessionId: string, updates: Partial<Session>) => void;
  removeSession: (sessionId: string) => void;
  setActiveSessions: (sessionIds: string[]) => void;
  setSessions: (sessions: Session[]) => void;

  // QR Code actions
  setQRCode: (sessionId: string, qrCode: string) => void;
  clearQRCode: (sessionId: string) => void;

  // WebSocket actions
  setWsConnected: (connected: boolean) => void;

  // Utility actions
  getSession: (sessionId: string) => Session | undefined;
  isSessionActive: (sessionId: string) => boolean;
  clearAll: () => void;
  setError: (error: string | null) => void;
}

// ============================================================================
// Store
// ============================================================================

export const useSessionStore = create<SessionState>()(
  devtools(
    persist(
      (set, get) => ({
        // Initial state - empty, will fetch from API
        sessions: [],
        activeSessions: new Set(),
        qrCodes: new Map(),
        wsConnected: false,
        isLoading: false,
        error: null,

        // Fetch sessions from API
        fetchSessions: async () => {
          set({ isLoading: true, error: null });
          try {
            const response = await apiClient.listSessions();
            const activeSessions = new Set(
              response.sessions
                .filter((s) => s.status === "connected")
                .map((s) => s.id),
            );
            set({
              sessions: response.sessions,
              activeSessions,
              isLoading: false,
            });
          } catch (error) {
            const message =
              error instanceof Error
                ? error.message
                : "Failed to fetch sessions";
            set({ error: message, isLoading: false });
          }
        },

        // Set sessions directly
        setSessions: (sessions) => {
          const activeSessions = new Set(
            sessions.filter((s) => s.status === "connected").map((s) => s.id),
          );
          set({ sessions, activeSessions });
        },

        // Session actions
        addSession: (session) =>
          set((state) => {
            const exists = state.sessions.some((s) => s.id === session.id);
            if (exists) {
              return {
                sessions: state.sessions.map((s) =>
                  s.id === session.id ? session : s,
                ),
              };
            }
            return { sessions: [...state.sessions, session] };
          }),

        updateSession: (sessionId, updates) =>
          set((state) => ({
            sessions: state.sessions.map((s) =>
              s.id === sessionId ? { ...s, ...updates } : s,
            ),
          })),

        removeSession: (sessionId) =>
          set((state) => {
            const newActiveSessions = new Set(state.activeSessions);
            newActiveSessions.delete(sessionId);

            const newQRCodes = new Map(state.qrCodes);
            newQRCodes.delete(sessionId);

            return {
              sessions: state.sessions.filter((s) => s.id !== sessionId),
              activeSessions: newActiveSessions,
              qrCodes: newQRCodes,
            };
          }),

        setActiveSessions: (sessionIds) =>
          set({ activeSessions: new Set(sessionIds) }),

        // QR Code actions
        setQRCode: (sessionId, qrCode) =>
          set((state) => {
            const newQRCodes = new Map(state.qrCodes);
            newQRCodes.set(sessionId, {
              sessionId,
              qrCode,
              timestamp: Date.now(),
            });
            return { qrCodes: newQRCodes };
          }),

        clearQRCode: (sessionId) =>
          set((state) => {
            const newQRCodes = new Map(state.qrCodes);
            newQRCodes.delete(sessionId);
            return { qrCodes: newQRCodes };
          }),

        // WebSocket actions
        setWsConnected: (connected) => set({ wsConnected: connected }),

        // Utility actions
        getSession: (sessionId) => {
          return get().sessions.find((s) => s.id === sessionId);
        },

        isSessionActive: (sessionId) => {
          return get().activeSessions.has(sessionId);
        },

        clearAll: () =>
          set({
            sessions: [],
            activeSessions: new Set(),
            qrCodes: new Map(),
            wsConnected: false,
            error: null,
          }),

        setError: (error) => set({ error }),
      }),
      {
        name: "session-storage",
        // Only persist sessions, not QR codes, loading, or WebSocket status
        partialize: (state) => ({
          sessions: state.sessions,
        }),
      },
    ),
    { name: "SessionStore" },
  ),
);
