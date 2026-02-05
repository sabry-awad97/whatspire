import { useEffect, useRef, useState } from "react";
import { Filter, RefreshCw } from "lucide-react";

import { Button } from "../ui/button";
import { Input } from "../ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../ui/select";
import { MessageItem } from "./message-item";

// ============================================================================
// Types
// ============================================================================

export interface Message {
  id: string;
  sessionId: string;
  chatId: string;
  type: "text" | "image" | "video" | "audio" | "document" | "sticker";
  content: string;
  timestamp: string;
  status: "sent" | "delivered" | "read" | "failed";
  isFromMe: boolean;
  senderName?: string;
  mediaUrl?: string;
}

interface MessageListProps {
  sessionId?: string;
  onFilterChange?: (filters: MessageFilters) => void;
}

export interface MessageFilters {
  sessionId?: string;
  chatId?: string;
  dateFrom?: string;
  dateTo?: string;
  type?: string;
}

// ============================================================================
// Mock Data
// ============================================================================

const MOCK_MESSAGES: Message[] = [
  {
    id: "msg-1",
    sessionId: "business-account",
    chatId: "1234567890@s.whatsapp.net",
    type: "text",
    content: "Hello! How can I help you today?",
    timestamp: new Date(Date.now() - 5 * 60 * 1000).toISOString(),
    status: "read",
    isFromMe: true,
  },
  {
    id: "msg-2",
    sessionId: "business-account",
    chatId: "1234567890@s.whatsapp.net",
    type: "text",
    content: "Hi! I need help with my order",
    timestamp: new Date(Date.now() - 4 * 60 * 1000).toISOString(),
    status: "delivered",
    isFromMe: false,
    senderName: "John Doe",
  },
  {
    id: "msg-3",
    sessionId: "business-account",
    chatId: "1234567890@s.whatsapp.net",
    type: "text",
    content: "Sure! What's your order number?",
    timestamp: new Date(Date.now() - 3 * 60 * 1000).toISOString(),
    status: "read",
    isFromMe: true,
  },
  {
    id: "msg-4",
    sessionId: "business-account",
    chatId: "9876543210@s.whatsapp.net",
    type: "image",
    content: "Check out this product!",
    timestamp: new Date(Date.now() - 2 * 60 * 1000).toISOString(),
    status: "sent",
    isFromMe: true,
    mediaUrl: "https://via.placeholder.com/300",
  },
];

// ============================================================================
// Component
// ============================================================================

export function MessageList({ sessionId, onFilterChange }: MessageListProps) {
  const [messages, setMessages] = useState<Message[]>(MOCK_MESSAGES);
  const [filters, setFilters] = useState<MessageFilters>({
    sessionId,
  });
  const [searchQuery, setSearchQuery] = useState("");
  const [autoScroll, setAutoScroll] = useState(true);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  // Auto-scroll to bottom when new messages arrive
  useEffect(() => {
    if (autoScroll && messagesEndRef.current) {
      messagesEndRef.current.scrollIntoView({ behavior: "smooth" });
    }
  }, [messages, autoScroll]);

  // Detect manual scroll to disable auto-scroll
  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    const handleScroll = () => {
      const { scrollTop, scrollHeight, clientHeight } = container;
      const isAtBottom = scrollHeight - scrollTop - clientHeight < 50;
      setAutoScroll(isAtBottom);
    };

    container.addEventListener("scroll", handleScroll);
    return () => container.removeEventListener("scroll", handleScroll);
  }, []);

  // Filter messages
  const filteredMessages = messages.filter((msg) => {
    if (filters.sessionId && msg.sessionId !== filters.sessionId) return false;
    if (filters.chatId && msg.chatId !== filters.chatId) return false;
    if (filters.type && msg.type !== filters.type) return false;
    if (
      searchQuery &&
      !msg.content.toLowerCase().includes(searchQuery.toLowerCase())
    )
      return false;
    return true;
  });

  const handleFilterChange = (key: keyof MessageFilters, value: string) => {
    const newFilters = { ...filters, [key]: value || undefined };
    setFilters(newFilters);
    onFilterChange?.(newFilters);
  };

  const handleRefresh = () => {
    // In real app, this would refetch from API
    console.log("Refreshing messages...");
  };

  const scrollToBottom = () => {
    setAutoScroll(true);
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  };

  return (
    <div className="flex flex-col h-full">
      {/* Filters */}
      <div className="glass-card-enhanced p-4 space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold">Messages</h2>
          <Button
            variant="outline"
            size="sm"
            onClick={handleRefresh}
            className="glass-card hover-glow-teal"
          >
            <RefreshCw className="h-4 w-4 mr-2" />
            Refresh
          </Button>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-4 gap-3">
          {/* Search */}
          <Input
            placeholder="Search messages..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="glass-card"
          />

          {/* Chat Filter */}
          <Select
            value={filters.chatId || "all"}
            onValueChange={(value) =>
              handleFilterChange("chatId", value === "all" ? "" : value)
            }
          >
            <SelectTrigger className="glass-card">
              <SelectValue placeholder="All Chats" />
            </SelectTrigger>
            <SelectContent className="glass-card-enhanced">
              <SelectItem value="all">All Chats</SelectItem>
              <SelectItem value="1234567890@s.whatsapp.net">
                John Doe
              </SelectItem>
              <SelectItem value="9876543210@s.whatsapp.net">
                Jane Smith
              </SelectItem>
            </SelectContent>
          </Select>

          {/* Type Filter */}
          <Select
            value={filters.type || "all"}
            onValueChange={(value) =>
              handleFilterChange("type", value === "all" ? "" : value)
            }
          >
            <SelectTrigger className="glass-card">
              <SelectValue placeholder="All Types" />
            </SelectTrigger>
            <SelectContent className="glass-card-enhanced">
              <SelectItem value="all">All Types</SelectItem>
              <SelectItem value="text">Text</SelectItem>
              <SelectItem value="image">Image</SelectItem>
              <SelectItem value="video">Video</SelectItem>
              <SelectItem value="audio">Audio</SelectItem>
              <SelectItem value="document">Document</SelectItem>
            </SelectContent>
          </Select>

          {/* Filter Button */}
          <Button variant="outline" className="glass-card hover-glow-emerald">
            <Filter className="h-4 w-4 mr-2" />
            More Filters
          </Button>
        </div>

        <div className="text-sm text-muted-foreground">
          Showing {filteredMessages.length} of {messages.length} messages
        </div>
      </div>

      {/* Messages */}
      <div ref={containerRef} className="flex-1 overflow-y-auto p-4 space-y-3">
        {filteredMessages.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full text-center">
            <svg
              className="h-16 w-16 text-muted-foreground mb-4"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={1.5}
                d="M8 10h.01M12 10h.01M16 10h.01M9 16H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-5l-5 5v-5z"
              />
            </svg>
            <h3 className="text-lg font-semibold mb-2">No messages found</h3>
            <p className="text-sm text-muted-foreground">
              Try adjusting your filters or search query
            </p>
          </div>
        ) : (
          <>
            {filteredMessages.map((message) => (
              <MessageItem key={message.id} message={message} />
            ))}
            <div ref={messagesEndRef} />
          </>
        )}
      </div>

      {/* Scroll to Bottom Button */}
      {!autoScroll && (
        <div className="absolute bottom-6 right-6">
          <Button
            onClick={scrollToBottom}
            className="glass-card hover-glow-teal rounded-full h-12 w-12 p-0 shadow-lg"
          >
            <svg
              className="h-5 w-5"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M19 14l-7 7m0 0l-7-7m7 7V3"
              />
            </svg>
          </Button>
        </div>
      )}
    </div>
  );
}
