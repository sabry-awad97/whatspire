import { Ban, MessageSquare, MoreVertical, Phone, UserX } from "lucide-react";
import { useState } from "react";
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
import type { Contact } from "./contact-list";

// ============================================================================
// Types
// ============================================================================

interface ContactCardProps {
  contact: Contact;
}

// ============================================================================
// Component
// ============================================================================

export function ContactCard({ contact }: ContactCardProps) {
  const [isBlocked, setIsBlocked] = useState(contact.isBlocked || false);

  const getInitials = (name: string) => {
    return name
      .split(" ")
      .map((n) => n[0])
      .join("")
      .toUpperCase()
      .slice(0, 2);
  };

  const formatLastSeen = (lastSeen?: string) => {
    if (!lastSeen) return "Never";

    const date = new Date(lastSeen);
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

  const handleMessage = () => {
    toast.info(`Opening chat with ${contact.name}`);
  };

  const handleCall = () => {
    toast.info(`Calling ${contact.name}`);
  };

  const handleBlock = () => {
    setIsBlocked(!isBlocked);
    toast.success(
      isBlocked
        ? `${contact.name} has been unblocked`
        : `${contact.name} has been blocked`,
    );
  };

  const handleDelete = () => {
    toast.success(`${contact.name} has been removed from contacts`);
  };

  return (
    <div
      className={cn(
        "glass-card-enhanced p-4 hover-lift transition-all",
        isBlocked && "opacity-60",
      )}
    >
      <div className="flex items-start gap-3">
        {/* Avatar */}
        <Avatar className="h-12 w-12 shrink-0">
          <AvatarImage src={contact.profilePicture} alt={contact.name} />
          <AvatarFallback className="bg-teal/20 text-teal font-semibold">
            {getInitials(contact.name)}
          </AvatarFallback>
        </Avatar>

        {/* Contact Info */}
        <div className="flex-1 min-w-0">
          <div className="flex items-start justify-between gap-2 mb-1">
            <h3 className="font-semibold truncate">{contact.name}</h3>
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
                  onClick={handleMessage}
                  className="hover-glow-teal"
                >
                  <MessageSquare className="mr-2 h-4 w-4" />
                  Send Message
                </DropdownMenuItem>
                <DropdownMenuItem
                  onClick={handleCall}
                  className="hover-glow-emerald"
                >
                  <Phone className="mr-2 h-4 w-4" />
                  Call
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem
                  onClick={handleBlock}
                  className="hover-glow-amber"
                >
                  <Ban className="mr-2 h-4 w-4" />
                  {isBlocked ? "Unblock" : "Block"}
                </DropdownMenuItem>
                <DropdownMenuItem
                  onClick={handleDelete}
                  className="text-destructive hover:text-destructive"
                >
                  <UserX className="mr-2 h-4 w-4" />
                  Delete Contact
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>

          <p className="text-sm text-muted-foreground truncate mb-2">
            {contact.phoneNumber}
          </p>

          {contact.status && (
            <p className="text-xs text-muted-foreground italic truncate mb-2">
              "{contact.status}"
            </p>
          )}

          {/* Last Seen */}
          <div className="flex items-center gap-2 text-xs text-muted-foreground">
            <div
              className={cn(
                "h-2 w-2 rounded-full",
                contact.lastSeen &&
                  new Date(contact.lastSeen).getTime() >
                    Date.now() - 5 * 60 * 1000
                  ? "bg-emerald animate-pulse"
                  : "bg-muted-foreground",
              )}
            />
            <span>Last seen {formatLastSeen(contact.lastSeen)}</span>
          </div>

          {/* Blocked Badge */}
          {isBlocked && (
            <div className="mt-2">
              <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium bg-destructive/10 text-destructive">
                <Ban className="h-3 w-3" />
                Blocked
              </span>
            </div>
          )}
        </div>
      </div>

      {/* Action Buttons */}
      {!isBlocked && (
        <div className="flex items-center gap-2 mt-4 pt-4 border-t border-border/50">
          <Button
            variant="outline"
            size="sm"
            onClick={handleMessage}
            className="glass-card hover-glow-teal flex-1"
          >
            <MessageSquare className="h-3 w-3 mr-2" />
            Message
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={handleCall}
            className="glass-card hover-glow-emerald flex-1"
          >
            <Phone className="h-3 w-3 mr-2" />
            Call
          </Button>
        </div>
      )}
    </div>
  );
}
