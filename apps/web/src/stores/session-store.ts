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
// Store
// ============================================================================

export const useSessionStore = create<SessionState>()(
  devtools(
    persist(
      (set, get) => ({
        // Initial state
        sessions: [],
        activeSessions: new Set(),
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
