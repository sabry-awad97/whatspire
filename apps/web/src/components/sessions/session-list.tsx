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
  const { sessions } = useSessionStore();

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
