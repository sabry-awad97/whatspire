import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { ArrowLeft, Eye, EyeOff, Save, X } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

import { useSessionStore } from "@/stores/session-store";

import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";

export const Route = createFileRoute("/sessions/$sessionId/webhooks")({
  component: WebhookConfigurationPage,
});

// Webhook event types
const WEBHOOK_EVENTS = [
  { id: "messages.received", label: "messages.received", category: "messages" },
  {
    id: "messages.groups.received",
    label: "messages-groups.received",
    category: "messages",
  },
  {
    id: "messages.newsletter.received",
    label: "messages-newsletter.received",
    category: "messages",
  },
  {
    id: "messages.personal.received",
    label: "messages-personal.received",
    category: "messages",
  },
  { id: "call", label: "call", category: "messages" },
  { id: "message.sent", label: "message.sent", category: "messages" },
  { id: "session.status", label: "session.status", category: "session" },
  { id: "qrcode.updated", label: "qrcode.updated", category: "session" },
  { id: "messages.upsert", label: "messages.upsert", category: "messages" },
  { id: "messages.update", label: "messages.update", category: "messages" },
  { id: "messages.delete", label: "messages.delete", category: "messages" },
  {
    id: "message-receipt.update",
    label: "message-receipt.update",
    category: "messages",
  },
  { id: "messages.reaction", label: "messages.reaction", category: "messages" },
  { id: "chats.upsert", label: "chats.upsert", category: "chats" },
  { id: "chats.update", label: "chats.update", category: "chats" },
  { id: "chats.delete", label: "chats.delete", category: "chats" },
  { id: "groups.upsert", label: "groups.upsert", category: "groups" },
  { id: "groups.update", label: "groups.update", category: "groups" },
  {
    id: "group-participants.update",
    label: "group-participants.update",
    category: "groups",
  },
  { id: "contacts.upsert", label: "contacts.upsert", category: "contacts" },
  { id: "contacts.update", label: "contacts.update", category: "contacts" },
  { id: "poll.results", label: "poll.results", category: "messages" },
];

function WebhookConfigurationPage() {
  const { sessionId } = Route.useParams();
  const navigate = useNavigate();
  const { getSession } = useSessionStore();

  const session = getSession(sessionId);

  const [endpointEnabled, setEndpointEnabled] = useState(false);
  const [webhookUrl, setWebhookUrl] = useState(
    "https://api.your-domain.com/webhook",
  );
  const [webhookSecret, setWebhookSecret] = useState(
    "••••••••••••••••••••••••••••••",
  );
  const [showSecret, setShowSecret] = useState(false);
  const [selectedEvents, setSelectedEvents] = useState<Set<string>>(new Set());
  const [messageFilteringExpanded, setMessageFilteringExpanded] =
    useState(false);
  const [ignoreGroups, setIgnoreGroups] = useState(false);
  const [ignoreBroadcasts, setIgnoreBroadcasts] = useState(false);
  const [ignoreChannels, setIgnoreChannels] = useState(false);

  if (!session) {
    return (
      <div className="min-h-screen network-bg flex items-center justify-center">
        <div className="glass-card-enhanced p-8 text-center">
          <h2 className="text-2xl font-bold mb-2">Session Not Found</h2>
          <p className="text-muted-foreground mb-4">
            The session you're looking for doesn't exist.
          </p>
          <button
            onClick={() => navigate({ to: "/sessions" })}
            className="text-primary hover:underline"
          >
            Back to Sessions
          </button>
        </div>
      </div>
    );
  }

  const handleToggleEvent = (eventId: string) => {
    const newSelected = new Set(selectedEvents);
    if (newSelected.has(eventId)) {
      newSelected.delete(eventId);
    } else {
      newSelected.add(eventId);
    }
    setSelectedEvents(newSelected);
  };

  const handleSave = () => {
    toast.success("Webhook configuration saved");
  };

  const handleCancel = () => {
    navigate({ to: "/sessions/$sessionId", params: { sessionId } });
  };

  const handleRotateSecret = () => {
    toast.success("Webhook secret rotated");
  };

  return (
    <div className="min-h-screen network-bg">
      {/* Header */}
      <div className="glass-card border-b border-border/50 px-6 py-4">
        <div className="max-w-7xl mx-auto flex items-center justify-between">
          <div className="flex items-center gap-4">
            <Button
              variant="ghost"
              size="icon"
              onClick={handleCancel}
              className="glass-card hover-lift"
            >
              <ArrowLeft className="h-5 w-5" />
            </Button>
            <div>
              <h1 className="text-2xl font-bold">Webhook Configuration</h1>
              <p className="text-sm text-muted-foreground">
                Manage real-time event notifications
              </p>
            </div>
          </div>

          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              onClick={handleCancel}
              className="glass-card"
            >
              <X className="mr-2 h-4 w-4" />
              Cancel
            </Button>
            <Button onClick={handleSave} className="glass-card hover-glow-teal">
              <Save className="mr-2 h-4 w-4" />
              Save Changes
            </Button>
          </div>
        </div>
      </div>

      <div className="max-w-7xl mx-auto p-6">
        <div className="grid gap-6 lg:grid-cols-3">
          {/* Left Column - Endpoint Settings */}
          <div className="lg:col-span-1 space-y-6">
            <div className="glass-card-enhanced p-6">
              <div className="flex items-center justify-between mb-4">
                <div className="flex items-center gap-3">
                  <div className="p-2 rounded-lg glass-card bg-teal/10">
                    <svg
                      className="h-5 w-5 text-teal"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M13 10V3L4 14h7v7l9-11h-7z"
                      />
                    </svg>
                  </div>
                  <h2 className="text-lg font-semibold">Endpoint Settings</h2>
                </div>
                <Switch
                  checked={endpointEnabled}
                  onCheckedChange={setEndpointEnabled}
                />
              </div>

              <p className="text-sm text-muted-foreground mb-6">
                Destination for POST requests
              </p>

              <div className="space-y-4">
                {/* Payload URL */}
                <div>
                  <Label htmlFor="webhook-url" className="text-sm mb-2 block">
                    PAYLOAD URL
                  </Label>
                  <Input
                    id="webhook-url"
                    type="url"
                    value={webhookUrl}
                    onChange={(e) => setWebhookUrl(e.target.value)}
                    disabled={!endpointEnabled}
                    className="glass-card"
                  />
                </div>

                {/* Webhook Secret */}
                <div>
                  <div className="flex items-center justify-between mb-2">
                    <Label htmlFor="webhook-secret" className="text-sm">
                      WEBHOOK SECRET
                    </Label>
                    <Button
                      variant="link"
                      size="sm"
                      onClick={handleRotateSecret}
                      disabled={!endpointEnabled}
                      className="text-xs h-auto p-0 text-teal hover:text-teal/80"
                    >
                      Rotate
                    </Button>
                  </div>
                  <div className="relative">
                    <Input
                      id="webhook-secret"
                      type={showSecret ? "text" : "password"}
                      value={webhookSecret}
                      disabled={!endpointEnabled}
                      className="glass-card pr-10"
                      readOnly
                    />
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => setShowSecret(!showSecret)}
                      disabled={!endpointEnabled}
                      className="absolute right-1 top-1/2 -translate-y-1/2 h-8 w-8"
                    >
                      {showSecret ? (
                        <EyeOff className="h-4 w-4" />
                      ) : (
                        <Eye className="h-4 w-4" />
                      )}
                    </Button>
                  </div>
                </div>

                {/* Message Filtering */}
                <div className="pt-4 border-t border-border/50">
                  <button
                    type="button"
                    onClick={() =>
                      setMessageFilteringExpanded(!messageFilteringExpanded)
                    }
                    className="w-full flex items-center justify-between mb-4"
                  >
                    <span className="text-sm font-medium">
                      Message Filtering
                    </span>
                    <span className="text-muted-foreground">
                      {messageFilteringExpanded ? "▲" : "▼"}
                    </span>
                  </button>

                  {messageFilteringExpanded && (
                    <div className="space-y-3">
                      <p className="text-xs text-muted-foreground mb-3">
                        Choose which types of messages to ignore. Ignored
                        messages won't trigger webhooks or be processed.
                      </p>

                      <div className="flex items-center space-x-2">
                        <Checkbox
                          id="ignore-groups"
                          checked={ignoreGroups}
                          onCheckedChange={(checked) =>
                            setIgnoreGroups(checked as boolean)
                          }
                          disabled={!endpointEnabled}
                        />
                        <div className="flex-1">
                          <Label
                            htmlFor="ignore-groups"
                            className="text-sm font-normal cursor-pointer"
                          >
                            Ignore Groups
                          </Label>
                          <p className="text-xs text-muted-foreground">
                            Skip group messages
                          </p>
                        </div>
                      </div>

                      <div className="flex items-center space-x-2">
                        <Checkbox
                          id="ignore-broadcasts"
                          checked={ignoreBroadcasts}
                          onCheckedChange={(checked) =>
                            setIgnoreBroadcasts(checked as boolean)
                          }
                          disabled={!endpointEnabled}
                        />
                        <div className="flex-1">
                          <Label
                            htmlFor="ignore-broadcasts"
                            className="text-sm font-normal cursor-pointer"
                          >
                            Ignore Broadcasts
                          </Label>
                          <p className="text-xs text-muted-foreground">
                            Skip broadcast lists
                          </p>
                        </div>
                      </div>

                      <div className="flex items-center space-x-2">
                        <Checkbox
                          id="ignore-channels"
                          checked={ignoreChannels}
                          onCheckedChange={(checked) =>
                            setIgnoreChannels(checked as boolean)
                          }
                          disabled={!endpointEnabled}
                        />
                        <div className="flex-1">
                          <Label
                            htmlFor="ignore-channels"
                            className="text-sm font-normal cursor-pointer"
                          >
                            Ignore Channels
                          </Label>
                          <p className="text-xs text-muted-foreground">
                            Skip channel updates
                          </p>
                        </div>
                      </div>
                    </div>
                  )}
                </div>

                {/* Security Note */}
                <div className="pt-4 border-t border-border/50">
                  <div className="flex items-start gap-2">
                    <svg
                      className="h-4 w-4 text-muted-foreground mt-0.5 shrink-0"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"
                      />
                    </svg>
                    <div>
                      <p className="text-xs font-medium">Security</p>
                      <p className="text-xs text-muted-foreground">
                        Verify the{" "}
                        <code className="text-teal">X-Webhook-Signature</code>{" "}
                        header
                      </p>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* Right Column - Subscriptions */}
          <div className="lg:col-span-2">
            <div className="glass-card-enhanced p-6">
              <div className="flex items-center justify-between mb-6">
                <div className="flex items-center gap-3">
                  <div className="p-2 rounded-lg glass-card bg-amber/10">
                    <svg
                      className="h-5 w-5 text-amber"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"
                      />
                    </svg>
                  </div>
                  <div>
                    <h2 className="text-lg font-semibold">Subscriptions</h2>
                    <p className="text-sm text-muted-foreground">
                      Trigger events
                    </p>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-sm text-muted-foreground">
                    {selectedEvents.size} Active
                  </span>
                </div>
              </div>

              <div className="grid grid-cols-3 gap-3">
                {WEBHOOK_EVENTS.map((event) => (
                  <button
                    key={event.id}
                    onClick={() => handleToggleEvent(event.id)}
                    disabled={!endpointEnabled}
                    className={`glass-card p-3 rounded-lg text-left transition-all hover-lift ${
                      selectedEvents.has(event.id)
                        ? "bg-teal/10 border-teal/50 glow-teal-sm"
                        : "hover:bg-muted/5"
                    } ${!endpointEnabled ? "opacity-50 cursor-not-allowed" : ""}`}
                  >
                    <span className="text-sm font-mono">{event.label}</span>
                  </button>
                ))}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
