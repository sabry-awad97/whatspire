import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { useApiClient, useSession, useDeleteSession } from "@whatspire/hooks";
import { Loader2 } from "lucide-react";

import { SessionDetails } from "@/components/sessions/session-details";

export const Route = createFileRoute("/sessions/$sessionId")({
  component: SessionDetailsPage,
});

function SessionDetailsPage() {
  const { sessionId } = Route.useParams();
  const navigate = useNavigate();
  const client = useApiClient();

  // Use hooks package to fetch session
  const { data: session, isLoading, error } = useSession(client, sessionId);

  // Delete session mutation
  const deleteSession = useDeleteSession(client, {
    onSuccess: () => {
      navigate({ to: "/sessions" });
    },
  });

  if (isLoading) {
    return (
      <div className="min-h-screen network-bg flex items-center justify-center">
        <div className="flex flex-col items-center space-y-4">
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
          <p className="text-sm text-muted-foreground">Loading session...</p>
        </div>
      </div>
    );
  }

  if (error || !session) {
    return (
      <div className="min-h-screen network-bg flex items-center justify-center">
        <div className="glass-card-enhanced p-8 text-center">
          <h2 className="text-2xl font-bold mb-2">Session Not Found</h2>
          <p className="text-muted-foreground mb-4">
            {error?.message || "The session you're looking for doesn't exist."}
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
