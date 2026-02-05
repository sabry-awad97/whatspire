import { useState } from "react";
import { Plus, RefreshCw, Search } from "lucide-react";

import { Button } from "../ui/button";
import { Input } from "../ui/input";
import { GroupCard } from "./group-card";

// ============================================================================
// Types
// ============================================================================

export interface Group {
  id: string;
  name: string;
  description?: string;
  profilePicture?: string;
  participantCount: number;
  unreadCount?: number;
  lastMessage?: {
    text: string;
    timestamp: string;
    sender: string;
  };
  isAdmin?: boolean;
}

interface GroupListProps {
  sessionId?: string;
  onSync?: () => void;
}

// ============================================================================
// Mock Data
// ============================================================================

const MOCK_GROUPS: Group[] = [
  {
    id: "1",
    name: "Family Group",
    description: "Our lovely family",
    profilePicture: "https://api.dicebear.com/7.x/identicon/svg?seed=family",
    participantCount: 8,
    unreadCount: 3,
    lastMessage: {
      text: "See you all tomorrow!",
      timestamp: new Date(Date.now() - 30 * 60 * 1000).toISOString(),
      sender: "Mom",
    },
    isAdmin: true,
  },
  {
    id: "2",
    name: "Work Team",
    description: "Project discussions",
    profilePicture: "https://api.dicebear.com/7.x/identicon/svg?seed=work",
    participantCount: 15,
    unreadCount: 12,
    lastMessage: {
      text: "Meeting at 3 PM",
      timestamp: new Date(Date.now() - 1 * 60 * 60 * 1000).toISOString(),
      sender: "John",
    },
    isAdmin: false,
  },
  {
    id: "3",
    name: "Friends Forever",
    profilePicture: "https://api.dicebear.com/7.x/identicon/svg?seed=friends",
    participantCount: 12,
    lastMessage: {
      text: "Who's up for dinner tonight?",
      timestamp: new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString(),
      sender: "Alice",
    },
    isAdmin: true,
  },
  {
    id: "4",
    name: "Book Club",
    description: "Monthly book discussions",
    profilePicture: "https://api.dicebear.com/7.x/identicon/svg?seed=books",
    participantCount: 6,
    unreadCount: 1,
    lastMessage: {
      text: "Next book: 1984 by George Orwell",
      timestamp: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(),
      sender: "Sarah",
    },
    isAdmin: false,
  },
];

// ============================================================================
// Component
// ============================================================================

export function GroupList({ sessionId, onSync }: GroupListProps) {
  const [groups, setGroups] = useState<Group[]>(MOCK_GROUPS);
  const [searchQuery, setSearchQuery] = useState("");
  const [isSyncing, setIsSyncing] = useState(false);

  const handleSync = async () => {
    setIsSyncing(true);
    // Simulate API call
    await new Promise((resolve) => setTimeout(resolve, 1500));
    setIsSyncing(false);
    onSync?.();
  };

  // Filter groups based on search
  const filteredGroups = groups.filter(
    (group) =>
      group.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      group.description?.toLowerCase().includes(searchQuery.toLowerCase()),
  );

  // Calculate total unread
  const totalUnread = groups.reduce(
    (sum, group) => sum + (group.unreadCount || 0),
    0,
  );

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="glass-card-enhanced p-4 space-y-4">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-lg font-semibold">Groups</h2>
            <p className="text-sm text-muted-foreground">
              {filteredGroups.length} of {groups.length} groups
              {totalUnread > 0 && (
                <span className="ml-2 text-teal">
                  • {totalUnread} unread messages
                </span>
              )}
            </p>
          </div>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={handleSync}
              disabled={isSyncing}
              className="glass-card hover-glow-teal"
            >
              <RefreshCw
                className={`h-4 w-4 mr-2 ${isSyncing ? "animate-spin" : ""}`}
              />
              {isSyncing ? "Syncing..." : "Sync Groups"}
            </Button>
            <Button
              variant="outline"
              size="sm"
              className="glass-card hover-glow-emerald"
            >
              <Plus className="h-4 w-4 mr-2" />
              Create Group
            </Button>
          </div>
        </div>

        {/* Search */}
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search groups by name or description..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="glass-card pl-10"
          />
        </div>
      </div>

      {/* Group List */}
      <div className="flex-1 overflow-y-auto p-4">
        {filteredGroups.length === 0 ? (
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
                d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"
              />
            </svg>
            <h3 className="text-lg font-semibold mb-2">No groups found</h3>
            <p className="text-sm text-muted-foreground mb-4">
              {searchQuery
                ? "Try adjusting your search query"
                : "Sync your groups to get started"}
            </p>
            {!searchQuery && (
              <Button
                onClick={handleSync}
                disabled={isSyncing}
                className="glass-card hover-glow-teal"
              >
                <RefreshCw
                  className={`h-4 w-4 mr-2 ${isSyncing ? "animate-spin" : ""}`}
                />
                Sync Groups
              </Button>
            )}
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {filteredGroups.map((group) => (
              <GroupCard key={group.id} group={group} />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
