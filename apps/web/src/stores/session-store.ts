import { create } from "zustand";
import { devtools, persist } from "zustand/middleware";

import type { Session } from "@/lib/api-client";

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

  // QR Codes
  qrCodes: Map<string, QRCodeData>;

  // WebSocket connection status
  wsConnected: boolean;

  // Actions
  addSession: (session: Session) => void;
  updateSession: (sessionId: string, updates: Partial<Session>) => void;
  removeSession: (sessionId: string) => void;
  setActiveSessions: (sessionIds: string[]) => void;

  // QR Code actions
  setQRCode: (sessionId: string, qrCode: string) => void;
  clearQRCode: (sessionId: string) => void;

  // WebSocket actions
  setWsConnected: (connected: boolean) => void;

  // Utility actions
  getSession: (sessionId: string) => Session | undefined;
  isSessionActive: (sessionId: string) => boolean;
  clearAll: () => void;
}

// ============================================================================
// Mock Data
// ============================================================================

const MOCK_SESSIONS: Session[] = [
  {
    id: "business-account",
    status: "connected",
    jid: "1234567890@s.whatsapp.net",
    created_at: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString(), // 7 days ago
    updated_at: new Date().toISOString(),
  },
  {
    id: "personal-account",
    status: "connected",
    jid: "9876543210@s.whatsapp.net",
    created_at: new Date(Date.now() - 3 * 24 * 60 * 60 * 1000).toISOString(), // 3 days ago
    updated_at: new Date().toISOString(),
  },
  {
    id: "support-team",
    status: "disconnected",
    created_at: new Date(Date.now() - 1 * 24 * 60 * 60 * 1000).toISOString(), // 1 day ago
    updated_at: new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString(), // 2 hours ago
  },
  {
    id: "test-session",
    status: "pending",
    created_at: new Date(Date.now() - 30 * 60 * 1000).toISOString(), // 30 minutes ago
    updated_at: new Date(Date.now() - 30 * 60 * 1000).toISOString(),
  },
];

// ============================================================================
// Store
// ============================================================================

export const useSessionStore = create<SessionState>()(
  devtools(
    persist(
      (set, get) => ({
        // Initial state with mock sessions
        sessions: MOCK_SESSIONS,
        activeSessions: new Set(["business-account", "personal-account"]),
        qrCodes: new Map(),
        wsConnected: false,

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
          }),
      }),
      {
        name: "session-storage",
        // Only persist sessions, not QR codes or WebSocket status
        partialize: (state) => ({
          sessions: state.sessions,
        }),
      },
    ),
    { name: "SessionStore" },
  ),
);
