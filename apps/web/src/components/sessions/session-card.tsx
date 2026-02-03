import {
  Circle,
  Edit,
  MoreVertical,
  Power,
  Settings,
  Trash2,
  Wifi,
  WifiOff,
} from "lucide-react";
import { useState } from "react";
import { useNavigate } from "@tanstack/react-router";
import { toast } from "sonner";

import type { Session } from "@/lib/api-client";
import { cn } from "@/lib/utils";
import { useSessionStore } from "@/stores/session-store";

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "../ui/alert-dialog";
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
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const navigate = useNavigate();
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

  const handleReconnect = async (e: React.MouseEvent) => {
    e.stopPropagation();
    setIsLoading(true);
    try {
      // Mock reconnection - just update status locally
      await new Promise((resolve) => setTimeout(resolve, 1000));
      updateSession(session.id, {
        status: "connected",
        updated_at: new Date().toISOString(),
      });
      toast.success("Session reconnected successfully");
    } catch (error) {
      toast.error("Failed to reconnect session");
    } finally {
      setIsLoading(false);
    }
  };

  const handleDisconnect = async (e: React.MouseEvent) => {
    e.stopPropagation();
    setIsLoading(true);
    try {
      // Mock disconnection - just update status locally
      await new Promise((resolve) => setTimeout(resolve, 1000));
      updateSession(session.id, {
        status: "disconnected",
        updated_at: new Date().toISOString(),
      });
      toast.success("Session disconnected");
    } catch (error) {
      toast.error("Failed to disconnect session");
    } finally {
      setIsLoading(false);
    }
  };

  const handleDelete = async (e: React.MouseEvent) => {
    e.stopPropagation();
    setShowDeleteDialog(true);
  };

  const confirmDelete = async () => {
    setIsLoading(true);
    try {
      // Mock deletion - just remove from store
      await new Promise((resolve) => setTimeout(resolve, 500));
      removeSession(session.id);
      toast.success("Session deleted");
    } catch (error) {
      toast.error("Failed to delete session");
    } finally {
      setIsLoading(false);
      setShowDeleteDialog(false);
    }
  };

  const handleManage = (e: React.MouseEvent) => {
    e.stopPropagation();
    navigate({
      to: "/sessions/$sessionId",
      params: { sessionId: session.id },
    });
  };

  const handleEdit = (e: React.MouseEvent) => {
    e.stopPropagation();
    navigate({
      to: "/sessions/$sessionId/edit",
      params: { sessionId: session.id },
    });
  };

  const handleCardClick = () => {
    // Removed auto-navigation on card click
    if (onSelect) {
      onSelect(session);
    }
  };

  return (
    <>
      <div
        className={cn(
          "glass-card-enhanced p-4 transition-all duration-300",
          config.glowClass && `hover:${config.glowClass}`,
          isLoading && "opacity-50 pointer-events-none",
        )}
        onClick={handleCardClick}
      >
        <div className="flex items-start justify-between mb-4">
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2 mb-2">
              <StatusIcon className={cn("h-4 w-4", config.color)} />
              <span className={cn("text-sm font-medium", config.color)}>
                {config.label}
              </span>
            </div>

            <h3 className="font-semibold text-lg truncate mb-1">
              {session.id}
            </h3>

            {session.jid && (
              <p className="text-sm text-muted-foreground truncate">
                {session.jid}
              </p>
            )}

            <p className="text-xs text-muted-foreground mt-2">
              Last active:{" "}
              {new Date(
                session.updated_at || session.created_at,
              ).toLocaleString()}
            </p>
          </div>

          <DropdownMenu>
            <DropdownMenuTrigger asChild onClick={(e) => e.stopPropagation()}>
              <Button
                variant="ghost"
                size="icon"
                disabled={isLoading}
                className="glass-card hover-lift relative z-10"
              >
                <MoreVertical className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent
              align="end"
              className="glass-card-enhanced z-50"
            >
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

        {/* Action Buttons */}
        <div className="flex items-center justify-between gap-2 pt-4 border-t border-border/50">
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={handleEdit}
              disabled={isLoading}
              className="glass-card hover-glow-teal"
            >
              <Edit className="mr-2 h-3 w-3" />
              Edit
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={handleDelete}
              disabled={isLoading}
              className="glass-card text-destructive hover:text-destructive hover-glow-destructive"
            >
              <Trash2 className="mr-2 h-3 w-3" />
              Delete
            </Button>
          </div>
          <Button
            variant="outline"
            size="sm"
            onClick={handleManage}
            disabled={isLoading}
            className="glass-card hover-glow-emerald"
          >
            <Settings className="mr-2 h-3 w-3" />
            Manage
          </Button>
        </div>
      </div>

      {/* Delete Confirmation Dialog */}
      <AlertDialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
        <AlertDialogContent className="glass-card-enhanced">
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Session</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete the session "
              <span className="font-semibold text-foreground">
                {session.id}
              </span>
              "? This action cannot be undone and will permanently remove all
              session data.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel className="glass-card hover-glow-teal">
              Cancel
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={confirmDelete}
              className="glass-card text-destructive hover:text-destructive hover-glow-destructive border-destructive/30"
            >
              <Trash2 className="mr-2 h-4 w-4" />
              Delete Session
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
