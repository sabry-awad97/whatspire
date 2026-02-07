import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { Plus } from "lucide-react";
import { useSessions } from "@/hooks";
import type { Session } from "@whatspire/schema";

import { Button } from "@/components/ui/button";
import { SessionCard } from "@/components/sessions/session-card";
import { Loader2 } from "lucide-react";

export const Route = createFileRoute("/sessions/")({
  component: SessionsComponent,
});

function SessionsComponent() {
  const navigate = useNavigate();

  // Use the reusable sessions hook
  const { data: sessions, isLoading, error, refetch } = useSessions();

  const sessionCount = sessions?.length || 0;

  return (
    <div className="min-h-screen network-bg">
      {/* Header Section */}
      <div className="glass-card border-b border-border/50 px-6 py-6">
        <div className="mx-auto">
          <div className="flex items-center justify-between mb-4">
            <div className="space-y-1">
              <h1 className="text-3xl font-bold gradient-text">
                WhatsApp Sessions
              </h1>
              <p className="text-sm text-muted-foreground">
                Manage your WhatsApp sessions and connections
              </p>
            </div>
            <div className="flex items-center gap-3">
              <Button
                variant="outline"
                size="icon"
                onClick={() => refetch()}
                className="glass-card hover-glow-teal"
                disabled={isLoading}
              >
                <svg
                  className={`h-4 w-4 ${isLoading ? "animate-spin" : ""}`}
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
                  />
                </svg>
              </Button>
              <Button
                onClick={() => navigate({ to: "/sessions/new" })}
                className="glass-card hover-glow-teal"
              >
                <Plus className="mr-2 h-4 w-4" />
                New Session
              </Button>
            </div>
          </div>

          {/* Stats Bar */}
          {!isLoading && sessionCount > 0 && (
            <div className="flex items-center gap-2 text-sm">
              <div className="glass-card px-3 py-1.5 rounded-lg">
                <span className="text-muted-foreground">Current Plan: </span>
                <span className="font-medium text-foreground">Basic</span>
              </div>
              <div className="glass-card px-3 py-1.5 rounded-lg">
                <span className="text-muted-foreground">
                  {sessionCount} of 1 WhatsApp session
                  {sessionCount !== 1 ? "s" : ""} used
                </span>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Content Section */}
      <div className="max-w-7xl mx-auto p-6">
        {/* Content */}
        {isLoading ? (
          <div className="flex flex-col items-center justify-center py-12 space-y-4">
            <Loader2 className="h-8 w-8 animate-spin text-primary" />
            <p className="text-sm text-muted-foreground">Loading sessions...</p>
          </div>
        ) : error ? (
          <div className="flex flex-col items-center justify-center py-12 space-y-4">
            <div className="glass-card p-6 text-center max-w-md">
              <p className="text-destructive font-medium mb-2">
                Failed to load sessions
              </p>
              <p className="text-sm text-muted-foreground mb-4">
                {error.message}
              </p>
            </div>
          </div>
        ) : !sessions || sessions.length === 0 ? (
          <div className="glass-card-enhanced p-12 text-center">
            <div className="max-w-md mx-auto space-y-6">
              {/* Icon */}
              <div className="relative inline-flex">
                <div className="absolute inset-0 blur-2xl opacity-20">
                  <div className="w-20 h-20 rounded-full bg-teal" />
                </div>
                <div className="relative p-5 rounded-2xl glass-card glow-teal-sm bg-teal/10">
                  <svg
                    className="w-10 h-10 text-teal"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={1.5}
                      d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"
                    />
                  </svg>
                </div>
              </div>

              {/* Text */}
              <div className="space-y-2">
                <h3 className="text-xl font-semibold text-foreground">
                  No WhatsApp Sessions
                </h3>
                <p className="text-sm text-muted-foreground">
                  You haven't created any WhatsApp sessions yet. Create your
                  first session to get started.
                </p>
              </div>

              {/* Action Button */}
              <Button
                onClick={() => navigate({ to: "/sessions/new" })}
                className="glass-card hover-glow-teal"
                size="lg"
              >
                <Plus className="mr-2 h-5 w-5" />
                Create Session
              </Button>

              {/* Additional Info */}
              <div className="pt-4 border-t border-border/50">
                <p className="text-xs text-muted-foreground">
                  Need help?{" "}
                  <button className="text-teal hover:text-teal/80 underline underline-offset-2">
                    View documentation
                  </button>
                </p>
              </div>
            </div>
          </div>
        ) : (
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {sessions.map((session: Session) => (
              <SessionCard key={session.id} session={session} />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
