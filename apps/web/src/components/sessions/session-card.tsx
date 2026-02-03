import {
  Circle,
  MoreVertical,
  Power,
  Trash2,
  Wifi,
  WifiOff,
} from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

import type { Session } from "@/lib/api-client";
import { apiClient } from "@/lib/api-client";
import { cn } from "@/lib/utils";
import { useSessionStore } from "@/stores/session-store";

import { Button } from "../ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "../ui/dropdown-menu";

// ============================================================================
// Types
// ============================================================================

interface SessionCardProps {
  session: Session;
  onSelect?: (session: Session) => void;
}

// ============================================================================
// Component
// ============================================================================

export function SessionCard({ session, onSelect }: SessionCardProps) {
  const [isLoading, setIsLoading] = useState(false);
  const { updateSession, removeSession } = useSessionStore();

  const statusConfig = {
    connected: {
      label: "Connected",
      color: "text-emerald",
      glowClass: "glow-emerald",
      icon: Wifi,
    },
    pending: {
      label: "Pending",
      color: "text-amber",
      glowClass: "glow-amber",
      icon: Circle,
    },
    disconnected: {
      label: "Disconnected",
      color: "text-muted-foreground",
      glowClass: "",
      icon: WifiOff,
    },
    error: {
      label: "Error",
      color: "text-destructive",
      glowClass: "",
      icon: WifiOff,
    },
  };

  const config = statusConfig[session.status];
  const StatusIcon = config.icon;

  const handleReconnect = async () => {
    setIsLoading(true);
    try {
      const updated = await apiClient.reconnectSession(session.id);
      updateSession(session.id, updated);
      toast.success("Session reconnected successfully");
    } catch (error) {
      toast.error(
        error instanceof Error ? error.message : "Failed to reconnect session",
      );
    } finally {
      setIsLoading(false);
    }
  };

  const handleDisconnect = async () => {
    setIsLoading(true);
    try {
      const updated = await apiClient.disconnectSession(session.id);
      updateSession(session.id, updated);
      toast.success("Session disconnected");
    } catch (error) {
      toast.error(
        error instanceof Error ? error.message : "Failed to disconnect session",
      );
    } finally {
      setIsLoading(false);
    }
  };

  const handleDelete = async () => {
    if (!confirm("Are you sure you want to delete this session?")) {
      return;
    }

    setIsLoading(true);
    try {
      await apiClient.unregisterSession(session.id);
      removeSession(session.id);
      toast.success("Session deleted");
    } catch (error) {
      toast.error(
        error instanceof Error ? error.message : "Failed to delete session",
      );
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div
      className={cn(
        "glass-card-enhanced p-4 hover-lift ripple transition-all duration-300 cursor-pointer",
        config.glowClass && `hover:${config.glowClass}`,
        isLoading && "opacity-50 pointer-events-none",
      )}
      onClick={() => onSelect?.(session)}
    >
      <div className="flex items-start justify-between">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 mb-2">
            <StatusIcon className={cn("h-4 w-4", config.color)} />
            <span className={cn("text-sm font-medium", config.color)}>
              {config.label}
            </span>
          </div>

          <h3 className="font-semibold text-lg truncate mb-1">{session.id}</h3>

          {session.jid && (
            <p className="text-sm text-muted-foreground truncate">
              {session.jid}
            </p>
          )}

          <p className="text-xs text-muted-foreground mt-2">
            Created: {new Date(session.created_at).toLocaleString()}
          </p>
        </div>

        <DropdownMenu>
          <DropdownMenuTrigger onClick={(e) => e.stopPropagation()}>
            <Button
              variant="ghost"
              size="icon"
              disabled={isLoading}
              className="glass-card hover-lift"
            >
              <MoreVertical className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="glass-card-enhanced">
            {session.status === "disconnected" && (
              <DropdownMenuItem
                onClick={handleReconnect}
                className="hover-glow-teal"
              >
                <Power className="mr-2 h-4 w-4" />
                Reconnect
              </DropdownMenuItem>
            )}
            {session.status === "connected" && (
              <DropdownMenuItem
                onClick={handleDisconnect}
                className="hover-glow-amber"
              >
                <WifiOff className="mr-2 h-4 w-4" />
                Disconnect
              </DropdownMenuItem>
            )}
            <DropdownMenuItem
              onClick={handleDelete}
              className="text-destructive hover:text-destructive"
            >
              <Trash2 className="mr-2 h-4 w-4" />
              Delete
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </div>
  );
}
