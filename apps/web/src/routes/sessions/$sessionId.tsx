import { createFileRoute, useNavigate } from "@tanstack/react-router";

import { SessionDetails } from "@/components/sessions/session-details";
import { useSessionStore } from "@/stores/session-store";

export const Route = createFileRoute("/sessions/$sessionId")({
  component: SessionDetailsPage,
});

function SessionDetailsPage() {
  const { sessionId } = Route.useParams();
  const navigate = useNavigate();
  const { getSession } = useSessionStore();

  const session = getSession(sessionId);

  if (!session) {
    return (
      <div className="min-h-screen network-bg flex items-center justify-center">
        <div className="glass-card-enhanced p-8 text-center">
          <h2 className="text-2xl font-bold mb-2">Session Not Found</h2>
          <p className="text-muted-foreground mb-4">
            The session you're looking for doesn't exist.
          </p>
          <button
            onClick={() => navigate({ to: "/sessions" })}
            className="text-primary hover:underline"
          >
            Back to Sessions
          </button>
        </div>
      </div>
    );
  }

  return (
    <SessionDetails
      session={session}
      onBack={() => navigate({ to: "/sessions" })}
    />
  );
}
