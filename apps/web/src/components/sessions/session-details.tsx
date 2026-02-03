import { Calendar, Circle, Hash, User, Wifi, WifiOff } from "lucide-react";

import type { Session } from "@/lib/api-client";
import { cn } from "@/lib/utils";

// ============================================================================
// Types
// ============================================================================

interface SessionDetailsProps {
  session: Session;
}

// ============================================================================
// Component
// ============================================================================

export function SessionDetails({ session }: SessionDetailsProps) {
  const statusConfig = {
    connected: {
      label: "Connected",
      color: "text-emerald",
      bgColor: "bg-emerald/10",
      glowClass: "glow-emerald",
      icon: Wifi,
    },
    pending: {
      label: "Pending Authentication",
      color: "text-amber",
      bgColor: "bg-amber/10",
      glowClass: "glow-amber",
      icon: Circle,
    },
    disconnected: {
      label: "Disconnected",
      color: "text-muted-foreground",
      bgColor: "bg-muted/10",
      glowClass: "",
      icon: WifiOff,
    },
    error: {
      label: "Error",
      color: "text-destructive",
      bgColor: "bg-destructive/10",
      glowClass: "",
      icon: WifiOff,
    },
  };

  const config = statusConfig[session.status];
  const StatusIcon = config.icon;

  return (
    <div
      className={cn(
        "glass-card-enhanced p-6 animate-scale-in",
        config.glowClass,
      )}
    >
      <h2 className="text-2xl font-bold mb-6 gradient-text">Session Details</h2>

      <div className="space-y-6">
        {/* Status */}
        <div className="flex items-start gap-3">
          <div className={cn("p-2 rounded-lg glass-card", config.bgColor)}>
            <StatusIcon className={cn("h-5 w-5", config.color)} />
          </div>
          <div className="flex-1">
            <p className="text-sm text-muted-foreground">Status</p>
            <p className={cn("font-medium", config.color)}>{config.label}</p>
          </div>
        </div>

        {/* Session ID */}
        <div className="flex items-start gap-3">
          <div className="p-2 rounded-lg glass-card bg-teal/10 glow-teal-sm">
            <Hash className="h-5 w-5 text-teal" />
          </div>
          <div className="flex-1">
            <p className="text-sm text-muted-foreground">Session ID</p>
            <p className="font-medium break-all">{session.id}</p>
          </div>
        </div>

        {/* JID */}
        {session.jid && (
          <div className="flex items-start gap-3">
            <div className="p-2 rounded-lg glass-card bg-primary/10">
              <User className="h-5 w-5 text-primary" />
            </div>
            <div className="flex-1">
              <p className="text-sm text-muted-foreground">WhatsApp JID</p>
              <p className="font-medium break-all">{session.jid}</p>
            </div>
          </div>
        )}

        {/* Created At */}
        <div className="flex items-start gap-3">
          <div className="p-2 rounded-lg glass-card bg-primary/10">
            <Calendar className="h-5 w-5 text-primary" />
          </div>
          <div className="flex-1">
            <p className="text-sm text-muted-foreground">Created</p>
            <p className="font-medium">
              {new Date(session.created_at).toLocaleString()}
            </p>
          </div>
        </div>

        {/* Updated At */}
        {session.updated_at && (
          <div className="flex items-start gap-3">
            <div className="p-2 rounded-lg glass-card bg-primary/10">
              <Calendar className="h-5 w-5 text-primary" />
            </div>
            <div className="flex-1">
              <p className="text-sm text-muted-foreground">Last Updated</p>
              <p className="font-medium">
                {new Date(session.updated_at).toLocaleString()}
              </p>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
