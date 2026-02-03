import {
  Check,
  CheckCheck,
  Clock,
  File,
  FileText,
  Image as ImageIcon,
  Music,
  Video,
  X,
} from "lucide-react";

import { cn } from "@/lib/utils";

import type { Message } from "./message-list";

// ============================================================================
// Types
// ============================================================================

interface MessageItemProps {
  message: Message;
}

// ============================================================================
// Component
// ============================================================================

export function MessageItem({ message }: MessageItemProps) {
  const getTypeIcon = () => {
    switch (message.type) {
      case "image":
        return <ImageIcon className="h-4 w-4" />;
      case "video":
        return <Video className="h-4 w-4" />;
      case "audio":
        return <Music className="h-4 w-4" />;
      case "document":
        return <FileText className="h-4 w-4" />;
      case "sticker":
        return <File className="h-4 w-4" />;
      default:
        return null;
    }
  };

  const getStatusIcon = () => {
    switch (message.status) {
      case "sent":
        return <Check className="h-3 w-3 text-muted-foreground" />;
      case "delivered":
        return <CheckCheck className="h-3 w-3 text-muted-foreground" />;
      case "read":
        return <CheckCheck className="h-3 w-3 text-teal" />;
      case "failed":
        return <X className="h-3 w-3 text-destructive" />;
      default:
        return <Clock className="h-3 w-3 text-muted-foreground" />;
    }
  };

  const getStatusBadge = () => {
    const statusConfig = {
      sent: {
        label: "Sent",
        color: "text-muted-foreground",
        bgColor: "bg-muted/20",
      },
      delivered: {
        label: "Delivered",
        color: "text-amber",
        bgColor: "bg-amber/10",
      },
      read: {
        label: "Read",
        color: "text-emerald",
        bgColor: "bg-emerald/10",
      },
      failed: {
        label: "Failed",
        color: "text-destructive",
        bgColor: "bg-destructive/10",
      },
    };

    const config = statusConfig[message.status];

    return (
      <span
        className={cn(
          "inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium",
          config.color,
          config.bgColor,
        )}
      >
        {getStatusIcon()}
        {config.label}
      </span>
    );
  };

  const formatTime = (timestamp: string) => {
    const date = new Date(timestamp);
    const now = new Date();
    const diffInHours = (now.getTime() - date.getTime()) / (1000 * 60 * 60);

    if (diffInHours < 24) {
      return date.toLocaleTimeString("en-US", {
        hour: "numeric",
        minute: "2-digit",
        hour12: true,
      });
    }

    return date.toLocaleDateString("en-US", {
      month: "short",
      day: "numeric",
      hour: "numeric",
      minute: "2-digit",
      hour12: true,
    });
  };

  return (
    <div
      className={cn(
        "glass-card-enhanced p-4 hover-lift transition-all",
        message.isFromMe && "border-l-2 border-teal",
        !message.isFromMe && "border-l-2 border-amber",
      )}
    >
      <div className="flex items-start justify-between gap-4 mb-2">
        <div className="flex items-center gap-2 flex-1 min-w-0">
          {/* Type Icon */}
          {getTypeIcon() && (
            <div
              className={cn(
                "shrink-0 p-1.5 rounded-lg",
                message.isFromMe
                  ? "bg-teal/10 text-teal"
                  : "bg-amber/10 text-amber",
              )}
            >
              {getTypeIcon()}
            </div>
          )}

          {/* Sender Info */}
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2">
              <span className="font-medium truncate">
                {message.isFromMe ? "You" : message.senderName || "Unknown"}
              </span>
              <span className="text-xs text-muted-foreground shrink-0">
                {formatTime(message.timestamp)}
              </span>
            </div>
            <p className="text-xs text-muted-foreground truncate">
              {message.chatId}
            </p>
          </div>
        </div>

        {/* Status Badge */}
        {message.isFromMe && getStatusBadge()}
      </div>

      {/* Message Content */}
      <div className="space-y-2">
        {message.mediaUrl && (
          <div className="rounded-lg overflow-hidden glass-card">
            <img
              src={message.mediaUrl}
              alt="Message media"
              className="w-full h-auto max-h-64 object-cover"
            />
          </div>
        )}

        {message.content && (
          <p className="text-sm whitespace-pre-wrap wrap-break-word">
            {message.content}
          </p>
        )}
      </div>

      {/* Message Type Badge */}
      {message.type !== "text" && (
        <div className="mt-2">
          <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium bg-primary/10 text-primary">
            {getTypeIcon()}
            {message.type.charAt(0).toUpperCase() + message.type.slice(1)}
          </span>
        </div>
      )}
    </div>
  );
}
