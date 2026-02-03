import { useQuery } from "@tanstack/react-query";
import { AlertCircle, Loader2 } from "lucide-react";

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

  // Note: In a real implementation, you would fetch sessions from the API
  // For now, we're using the local store
  const { isLoading, error } = useQuery({
    queryKey: ["sessions"],
    queryFn: async () => {
      // Placeholder for API call
      // const response = await apiClient.getSessions();
      // return response.sessions;
      return sessions;
    },
    enabled: false, // Disable for now since we don't have a getSessions endpoint
  });

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center py-12 space-y-2">
        <AlertCircle className="h-8 w-8 text-destructive" />
        <p className="text-sm text-muted-foreground">Failed to load sessions</p>
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
