import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { Plus } from "lucide-react";

import { Button } from "@/components/ui/button";
import { SessionList } from "@/components/sessions/session-list";

export const Route = createFileRoute("/sessions/")({
  component: SessionsComponent,
});

function SessionsComponent() {
  const navigate = useNavigate();

  return (
    <div className="min-h-screen network-bg p-6">
      <div className="max-w-7xl mx-auto space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div className="space-y-2">
            <h1 className="text-3xl font-bold gradient-text">Sessions</h1>
            <p className="text-muted-foreground">
              Manage your WhatsApp sessions
            </p>
          </div>
          <Button
            onClick={() => navigate({ to: "/sessions/new" })}
            className="glass-card hover-glow-teal"
          >
            <Plus className="mr-2 h-4 w-4" />
            Add Session
          </Button>
        </div>

        {/* Content */}
        <SessionList />
      </div>
    </div>
  );
}
