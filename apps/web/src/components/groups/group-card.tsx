import {
  Crown,
  LogOut,
  MessageSquare,
  MoreVertical,
  Settings,
  Users,
} from "lucide-react";
import { useState } from "react";
import { useNavigate } from "@tanstack/react-router";
import { toast } from "sonner";

import { cn } from "@/lib/utils";

import { Avatar, AvatarFallback, AvatarImage } from "../ui/avatar";
import { Button } from "../ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "../ui/dropdown-menu";
import type { Group } from "./group-list";

// ============================================================================
// Types
// ============================================================================

interface GroupCardProps {
  group: Group;
}

// ============================================================================
// Component
// ============================================================================

export function GroupCard({ group }: GroupCardProps) {
  const navigate = useNavigate();
  const [unreadCount, setUnreadCount] = useState(group.unreadCount || 0);

  const getInitials = (name: string) => {
    return name
      .split(" ")
      .map((n) => n[0])
      .join("")
      .toUpperCase()
      .slice(0, 2);
  };

  const formatTimestamp = (timestamp: string) => {
    const date = new Date(timestamp);
    const now = new Date();
    const diffInMinutes = Math.floor(
      (now.getTime() - date.getTime()) / (1000 * 60),
    );

    if (diffInMinutes < 1) return "Just now";
    if (diffInMinutes < 60) return `${diffInMinutes}m ago`;

    const diffInHours = Math.floor(diffInMinutes / 60);
    if (diffInHours < 24) return `${diffInHours}h ago`;

    const diffInDays = Math.floor(diffInHours / 24);
    if (diffInDays === 1) return "Yesterday";
    if (diffInDays < 7) return `${diffInDays}d ago`;

    return date.toLocaleDateString("en-US", {
      month: "short",
      day: "numeric",
    });
  };

  const handleViewDetails = () => {
    navigate({
      to: "/groups/$groupId",
      params: { groupId: group.id },
    });
  };

  const handleOpenChat = () => {
    setUnreadCount(0);
    toast.info(`Opening chat for ${group.name}`);
  };

  const handleLeave = () => {
    toast.success(`Left ${group.name}`);
  };

  return (
    <div className="glass-card-enhanced p-4 hover-lift transition-all">
      <div className="flex items-start gap-3">
        {/* Avatar */}
        <div className="relative">
          <Avatar className="h-12 w-12 shrink-0">
            <AvatarImage src={group.profilePicture} alt={group.name} />
            <AvatarFallback className="bg-emerald/20 text-emerald font-semibold">
              {getInitials(group.name)}
            </AvatarFallback>
          </Avatar>
          {unreadCount > 0 && (
            <div className="absolute -top-1 -right-1 h-5 w-5 rounded-full bg-teal flex items-center justify-center text-xs font-bold text-background">
              {unreadCount > 9 ? "9+" : unreadCount}
            </div>
          )}
        </div>

        {/* Group Info */}
        <div className="flex-1 min-w-0">
          <div className="flex items-start justify-between gap-2 mb-1">
            <div className="flex items-center gap-2 flex-1 min-w-0">
              <h3 className="font-semibold truncate">{group.name}</h3>
              {group.isAdmin && (
                <Crown className="h-4 w-4 text-amber shrink-0" />
              )}
            </div>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-8 w-8 shrink-0"
                >
                  <MoreVertical className="h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="glass-card-enhanced">
                <DropdownMenuItem
                  onClick={handleOpenChat}
                  className="hover-glow-teal"
                >
                  <MessageSquare className="mr-2 h-4 w-4" />
                  Open Chat
                </DropdownMenuItem>
                <DropdownMenuItem
                  onClick={handleViewDetails}
                  className="hover-glow-emerald"
                >
                  <Users className="mr-2 h-4 w-4" />
                  View Details
                </DropdownMenuItem>
                {group.isAdmin && (
                  <DropdownMenuItem className="hover-glow-amber">
                    <Settings className="mr-2 h-4 w-4" />
                    Group Settings
                  </DropdownMenuItem>
                )}
                <DropdownMenuSeparator />
                <DropdownMenuItem
                  onClick={handleLeave}
                  className="text-destructive hover:text-destructive"
                >
                  <LogOut className="mr-2 h-4 w-4" />
                  Leave Group
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>

          {group.description && (
            <p className="text-xs text-muted-foreground truncate mb-2">
              {group.description}
            </p>
          )}

          {/* Participant Count */}
          <div className="flex items-center gap-2 text-xs text-muted-foreground mb-2">
            <Users className="h-3 w-3" />
            <span>{group.participantCount} participants</span>
          </div>

          {/* Last Message */}
          {group.lastMessage && (
            <div className="text-xs">
              <p className="text-muted-foreground truncate">
                <span className="font-medium">{group.lastMessage.sender}:</span>{" "}
                {group.lastMessage.text}
              </p>
              <p className="text-muted-foreground mt-1">
                {formatTimestamp(group.lastMessage.timestamp)}
              </p>
            </div>
          )}
        </div>
      </div>

      {/* Action Buttons */}
      <div className="flex items-center gap-2 mt-4 pt-4 border-t border-border/50">
        <Button
          variant="outline"
          size="sm"
          onClick={handleOpenChat}
          className="glass-card hover-glow-teal flex-1"
        >
          <MessageSquare className="h-3 w-3 mr-2" />
          Open Chat
        </Button>
        <Button
          variant="outline"
          size="sm"
          onClick={handleViewDetails}
          className="glass-card hover-glow-emerald flex-1"
        >
          <Users className="h-3 w-3 mr-2" />
          Details
        </Button>
      </div>
    </div>
  );
}
