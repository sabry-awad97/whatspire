import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";

import { AddSessionDialog } from "@/components/sessions/add-session-dialog";
import { SessionDetails } from "@/components/sessions/session-details";
import { SessionList } from "@/components/sessions/session-list";
import type { Session } from "@/lib/api-client";

export const Route = createFileRoute("/sessions/")({
  component: SessionsComponent,
});

function SessionsComponent() {
  const [selectedSession, setSelectedSession] = useState<Session | null>(null);

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
          <AddSessionDialog />
        </div>

        {/* Content */}
        <div className="grid gap-6 lg:grid-cols-3">
          <div className="lg:col-span-2">
            <SessionList onSelectSession={setSelectedSession} />
          </div>

          {selectedSession && (
            <div className="lg:col-span-1 animate-scale-in">
              <SessionDetails session={selectedSession} />
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
