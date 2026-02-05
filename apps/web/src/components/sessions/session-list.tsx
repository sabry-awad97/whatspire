import { useEffect } from "react";
import { Loader2 } from "lucide-react";

import type { Session } from "@/lib/api-client";
import { useSessionStore } from "@/stores/session-store";

import { SessionCard } from "./session-card";

// ============================================================================
// Types
// ============================================================================

interface SessionListProps {
  onSelectSession?: (session: Session) => void;
}

// ============================================================================
// Component
// ============================================================================

export function SessionList({ onSelectSession }: SessionListProps) {
  const { sessions, isLoading, error, fetchSessions } = useSessionStore();

  useEffect(() => {
    fetchSessions();
  }, [fetchSessions]);

  if (isLoading) {
    return (
      <div className="flex flex-col items-center justify-center py-12 space-y-4">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
        <p className="text-sm text-muted-foreground">Loading sessions...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center py-12 space-y-4">
        <div className="glass-card p-6 text-center max-w-md">
          <p className="text-destructive font-medium mb-2">
            Failed to load sessions
          </p>
          <p className="text-sm text-muted-foreground mb-4">{error}</p>
          <button
            onClick={() => fetchSessions()}
            className="text-sm text-primary hover:underline"
          >
            Try again
          </button>
        </div>
      </div>
    );
  }

  if (sessions.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 space-y-2">
        <p className="text-sm text-muted-foreground">No sessions yet</p>
        <p className="text-xs text-muted-foreground">
          Click "Add Session" to create your first session
        </p>
      </div>
    );
  }

  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
      {sessions.map((session) => (
        <SessionCard
          key={session.id}
          session={session}
          onSelect={onSelectSession}
        />
      ))}
    </div>
  );
}
