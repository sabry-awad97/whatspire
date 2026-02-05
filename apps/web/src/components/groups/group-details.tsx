import {
  ArrowLeft,
  Crown,
  MessageSquare,
  UserMinus,
  UserPlus,
} from "lucide-react";
import { useNavigate } from "@tanstack/react-router";
import { toast } from "sonner";

import { Avatar, AvatarFallback, AvatarImage } from "../ui/avatar";
import { Button } from "../ui/button";
import type { Group } from "./group-list";

// ============================================================================
// Types
// ============================================================================

interface Participant {
  id: string;
  name: string;
  phoneNumber: string;
  profilePicture?: string;
  isAdmin?: boolean;
}

interface GroupDetailsProps {
  group: Group;
}

// ============================================================================
// Mock Data
// ============================================================================

const MOCK_PARTICIPANTS: Participant[] = [
  {
    id: "1",
    name: "John Doe",
    phoneNumber: "+1234567890",
    profilePicture: "https://api.dicebear.com/7.x/avataaars/svg?seed=John",
    isAdmin: true,
  },
  {
    id: "2",
    name: "Jane Smith",
    phoneNumber: "+9876543210",
    profilePicture: "https://api.dicebear.com/7.x/avataaars/svg?seed=Jane",
  },
  {
    id: "3",
    name: "Bob Johnson",
    phoneNumber: "+1122334455",
    profilePicture: "https://api.dicebear.com/7.x/avataaars/svg?seed=Bob",
  },
  {
    id: "4",
    name: "Alice Williams",
    phoneNumber: "+5544332211",
    profilePicture: "https://api.dicebear.com/7.x/avataaars/svg?seed=Alice",
    isAdmin: true,
  },
];

// ============================================================================
// Component
// ============================================================================

export function GroupDetails({ group }: GroupDetailsProps) {
  const navigate = useNavigate();

  const getInitials = (name: string) => {
    return name
      .split(" ")
      .map((n) => n[0])
      .join("")
      .toUpperCase()
      .slice(0, 2);
  };

  const handleAddParticipant = () => {
    toast.info("Add participant functionality coming soon");
  };

  const handleRemoveParticipant = (participant: Participant) => {
    toast.success(`${participant.name} removed from group`);
  };

  const handleMessage = (participant: Participant) => {
    toast.info(`Opening chat with ${participant.name}`);
  };

  return (
    <div className="min-h-screen network-bg">
      {/* Header */}
      <div className="glass-card border-b border-border/50 px-6 py-4">
        <div className="max-w-4xl mx-auto flex items-center gap-4">
          <Button
            variant="ghost"
            size="icon"
            onClick={() => navigate({ to: "/groups" })}
            className="glass-card hover-lift"
          >
            <ArrowLeft className="h-5 w-5" />
          </Button>
          <div>
            <h1 className="text-2xl font-bold">{group.name}</h1>
            <p className="text-sm text-muted-foreground">
              {group.participantCount} participants
            </p>
          </div>
        </div>
      </div>

      <div className="max-w-4xl mx-auto p-6 space-y-6">
        {/* Group Info */}
        <div className="glass-card-enhanced p-6">
          <div className="flex items-start gap-4">
            <Avatar className="h-20 w-20">
              <AvatarImage src={group.profilePicture} alt={group.name} />
              <AvatarFallback className="bg-emerald/20 text-emerald font-semibold text-2xl">
                {getInitials(group.name)}
              </AvatarFallback>
            </Avatar>

            <div className="flex-1">
              <h2 className="text-xl font-semibold mb-2">{group.name}</h2>
              {group.description && (
                <p className="text-sm text-muted-foreground mb-4">
                  {group.description}
                </p>
              )}
              <div className="flex items-center gap-2">
                {group.isAdmin && (
                  <span className="inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium bg-amber/10 text-amber">
                    <Crown className="h-3 w-3" />
                    Admin
                  </span>
                )}
                <span className="text-sm text-muted-foreground">
                  Created on{" "}
                  {new Date(
                    Date.now() - 30 * 24 * 60 * 60 * 1000,
                  ).toLocaleDateString()}
                </span>
              </div>
            </div>
          </div>
        </div>

        {/* Participants */}
        <div className="glass-card-enhanced p-6">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold">
              Participants ({MOCK_PARTICIPANTS.length})
            </h3>
            {group.isAdmin && (
              <Button
                onClick={handleAddParticipant}
                size="sm"
                className="glass-card hover-glow-teal"
              >
                <UserPlus className="h-4 w-4 mr-2" />
                Add Participant
              </Button>
            )}
          </div>

          <div className="space-y-3">
            {MOCK_PARTICIPANTS.map((participant) => (
              <div
                key={participant.id}
                className="glass-card p-4 rounded-lg hover-lift transition-all"
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3 flex-1 min-w-0">
                    <Avatar className="h-10 w-10">
                      <AvatarImage
                        src={participant.profilePicture}
                        alt={participant.name}
                      />
                      <AvatarFallback className="bg-teal/20 text-teal font-semibold">
                        {getInitials(participant.name)}
                      </AvatarFallback>
                    </Avatar>

                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2">
                        <p className="font-medium truncate">
                          {participant.name}
                        </p>
                        {participant.isAdmin && (
                          <Crown className="h-4 w-4 text-amber shrink-0" />
                        )}
                      </div>
                      <p className="text-sm text-muted-foreground truncate">
                        {participant.phoneNumber}
                      </p>
                    </div>
                  </div>

                  <div className="flex items-center gap-2">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => handleMessage(participant)}
                      className="glass-card hover-glow-teal"
                    >
                      <MessageSquare className="h-4 w-4" />
                    </Button>
                    {group.isAdmin && !participant.isAdmin && (
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handleRemoveParticipant(participant)}
                        className="glass-card text-destructive hover:text-destructive hover-glow-destructive"
                      >
                        <UserMinus className="h-4 w-4" />
                      </Button>
                    )}
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
