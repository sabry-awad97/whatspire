import {
  ArrowLeft,
  Calendar,
  CheckCircle2,
  Circle,
  Copy,
  Eye,
  EyeOff,
  ExternalLink,
  Hash,
  HelpCircle,
  Phone,
  QrCode,
  RefreshCw,
  ShieldCheck,
  User,
  Webhook,
  Wifi,
  WifiOff,
} from "lucide-react";
import { useState } from "react";
import { useNavigate } from "@tanstack/react-router";
import { toast } from "sonner";

import type { Session } from "@/lib/api-client";
import { cn } from "@/lib/utils";

import { Button } from "../ui/button";
import { Input } from "../ui/input";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "../ui/tabs";
import { QRCodeDisplay } from "./qr-code-display";

// ============================================================================
// Types
// ============================================================================

interface SessionDetailsProps {
  session: Session;
  onBack?: () => void;
}

// ============================================================================
// Component
// ============================================================================

export function SessionDetails({ session, onBack }: SessionDetailsProps) {
  const navigate = useNavigate();
  const [showQRCode, setShowQRCode] = useState(
    session.status === "pending" || session.status === "disconnected",
  );
  const [showApiToken, setShowApiToken] = useState(false);

  // Mock API token for demonstration
  const apiToken = "b295ef...438d";

  const statusConfig = {
    connected: {
      label: "Connected",
      color: "text-emerald",
      bgColor: "bg-emerald/10",
      glowClass: "glow-emerald",
      icon: Wifi,
      description: "The WhatsApp session is connected and ready to use",
    },
    pending: {
      label: "Needs QR Scan",
      color: "text-amber",
      bgColor: "bg-amber/10",
      glowClass: "glow-amber",
      icon: Circle,
      description: "The WhatsApp session needs QR code scanning to connect",
    },
    disconnected: {
      label: "Disconnected",
      color: "text-muted-foreground",
      bgColor: "bg-muted/10",
      glowClass: "",
      icon: WifiOff,
      description: "Session is disconnected. Scan QR code to reconnect",
    },
    error: {
      label: "Error",
      color: "text-destructive",
      bgColor: "bg-destructive/10",
      glowClass: "",
      icon: WifiOff,
      description: "An error occurred with this session",
    },
  };

  const config = statusConfig[session.status];
  const StatusIcon = config.icon;

  const handleAuthenticated = () => {
    setShowQRCode(false);
  };

  const handleError = (message: string) => {
    console.error("QR authentication error:", message);
  };

  const handleCopyToken = () => {
    navigator.clipboard.writeText(apiToken);
    toast.success("API token copied to clipboard");
  };

  const lastActive = session.updated_at
    ? new Date(session.updated_at).toLocaleString("en-US", {
        month: "numeric",
        day: "numeric",
        year: "numeric",
        hour: "numeric",
        minute: "numeric",
        hour12: true,
      })
    : "N/A";

  return (
    <div className="min-h-screen network-bg">
      {/* Header */}
      <div className="glass-card border-b border-border/50 px-6 py-4">
        <div className="max-w-7xl mx-auto flex items-center justify-between">
          <div className="flex items-center gap-4">
            {onBack && (
              <Button
                variant="ghost"
                size="icon"
                onClick={onBack}
                className="glass-card hover-lift"
              >
                <ArrowLeft className="h-5 w-5" />
              </Button>
            )}
            <div>
              <h1 className="text-2xl font-bold">{session.id}</h1>
              <div className="flex items-center gap-2 mt-1">
                <span className="text-sm text-muted-foreground">
                  {session.jid
                    ? `+${session.jid.split("@")[0]}`
                    : "+201091921210"}
                </span>
              </div>
            </div>
          </div>

          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="icon"
              className="glass-card hover-glow-teal"
            >
              <RefreshCw className="h-4 w-4" />
            </Button>
            <Button
              variant="outline"
              className="glass-card hover-glow-teal"
              onClick={() =>
                navigate({
                  to: "/sessions/$sessionId/webhooks",
                  params: { sessionId: session.id },
                })
              }
            >
              <Webhook className="mr-2 h-4 w-4" />
              Manage Webhook
            </Button>
            <Button variant="destructive" className="glass-card">
              Delete
            </Button>
          </div>
        </div>
      </div>

      <div className="max-w-7xl mx-auto p-6">
        <div className="grid gap-6 lg:grid-cols-3">
          {/* Left Column - Session Information */}
          <div className="lg:col-span-1 space-y-6">
            <div className="glass-card-enhanced p-6">
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-lg font-semibold">Session Information</h2>
              </div>

              <p className="text-sm text-muted-foreground mb-6">
                Details about this WhatsApp session
              </p>

              <div className="space-y-4">
                {/* Session Name */}
                <div>
                  <p className="text-sm text-muted-foreground mb-1">
                    Session Name
                  </p>
                  <p className="font-medium">{session.id}</p>
                </div>

                {/* Phone Number */}
                <div>
                  <p className="text-sm text-muted-foreground mb-1">
                    Phone Number
                  </p>
                  <div className="flex items-center gap-2">
                    <Phone className="h-4 w-4 text-muted-foreground" />
                    <p className="font-medium">
                      {session.jid
                        ? `+${session.jid.split("@")[0]}`
                        : "+201091921210"}
                    </p>
                  </div>
                </div>

                {/* Status */}
                <div>
                  <p className="text-sm text-muted-foreground mb-1">Status</p>
                  <div
                    className={cn(
                      "inline-flex items-center gap-2 px-3 py-1.5 rounded-lg glass-card",
                      config.bgColor,
                    )}
                  >
                    <CheckCircle2 className={cn("h-4 w-4", config.color)} />
                    <span className={cn("text-sm font-medium", config.color)}>
                      {config.label}
                    </span>
                  </div>
                  <p className="text-xs text-muted-foreground mt-2">
                    {config.description}
                  </p>
                </div>

                {/* Last Active */}
                <div>
                  <p className="text-sm text-muted-foreground mb-1">
                    Last Active
                  </p>
                  <p className="font-medium">{lastActive}</p>
                </div>

                {/* Action Buttons */}
                <div className="space-y-2 pt-4">
                  {session.status === "connected" && (
                    <>
                      <Button
                        variant="outline"
                        className="w-full glass-card hover-glow-amber"
                      >
                        <WifiOff className="mr-2 h-4 w-4" />
                        Disconnect
                      </Button>
                      <Button
                        variant="outline"
                        className="w-full glass-card hover-glow-teal"
                      >
                        <RefreshCw className="mr-2 h-4 w-4" />
                        Restart
                      </Button>
                    </>
                  )}
                  {(session.status === "pending" ||
                    session.status === "disconnected") && (
                    <Button
                      onClick={() => setShowQRCode(true)}
                      className="w-full glass-card hover-glow-teal"
                    >
                      <RefreshCw className="mr-2 h-4 w-4" />
                      Refresh QR
                    </Button>
                  )}
                </div>
              </div>
            </div>

            {/* Next Steps */}
            {session.status === "connected" && (
              <div className="glass-card-enhanced p-6">
                <h2 className="text-lg font-semibold mb-4">Next Steps</h2>

                <div className="space-y-4">
                  <div className="flex gap-3">
                    <div className="shrink-0 w-6 h-6 rounded-full bg-teal/20 flex items-center justify-center text-teal text-sm font-medium">
                      1
                    </div>
                    <div className="flex-1">
                      <p className="text-sm">
                        Copy your API key from the{" "}
                        <span className="text-teal font-medium">
                          Credentials
                        </span>{" "}
                        tab to use in your integration
                      </p>
                    </div>
                  </div>

                  <div className="flex gap-3">
                    <div className="shrink-0 w-6 h-6 rounded-full bg-teal/20 flex items-center justify-center text-teal text-sm font-medium">
                      2
                    </div>
                    <div className="flex-1">
                      <p className="text-sm">
                        Send your first test message using either cURL commands{" "}
                        <span className="text-emerald font-medium">
                          Postman
                        </span>{" "}
                        or the{" "}
                        <span className="text-emerald font-medium">
                          Test Sending
                        </span>{" "}
                        tab
                      </p>
                    </div>
                  </div>

                  <div className="flex gap-3">
                    <div className="shrink-0 w-6 h-6 rounded-full bg-teal/20 flex items-center justify-center text-teal text-sm font-medium">
                      3
                    </div>
                    <div className="flex-1">
                      <p className="text-sm">
                        View the complete{" "}
                        <span className="text-emerald font-medium">
                          API Documentation
                        </span>{" "}
                        for all available endpoints
                      </p>
                    </div>
                  </div>
                </div>
              </div>
            )}
          </div>

          {/* Right Column - QR Code or Tabs */}
          <div className="lg:col-span-2">
            {showQRCode ? (
              <div className="glass-card-enhanced p-6">
                <div className="flex items-center justify-between mb-6">
                  <Button
                    variant="outline"
                    className="glass-card hover-glow-emerald"
                  >
                    <HelpCircle className="mr-2 h-4 w-4" />
                    How to Scan QR Code
                  </Button>
                  <Button
                    variant="outline"
                    className="glass-card hover-glow-amber"
                  >
                    <ShieldCheck className="mr-2 h-4 w-4" />
                    Tips to avoid bans
                  </Button>
                </div>

                <div className="space-y-6">
                  <div>
                    <h2 className="text-xl font-semibold mb-2">Scan QR Code</h2>
                    <p className="text-sm text-muted-foreground">
                      Scan this QR code with your WhatsApp app to connect your
                      account
                    </p>
                  </div>

                  <QRCodeDisplay
                    sessionId={session.id}
                    onAuthenticated={handleAuthenticated}
                    onError={handleError}
                  />
                </div>
              </div>
            ) : (
              <div className="glass-card-enhanced">
                <Tabs defaultValue="credentials" className="w-full">
                  <TabsList className="w-full justify-start border-b border-border/50 rounded-none bg-transparent p-0">
                    <TabsTrigger
                      value="credentials"
                      className="data-[state=active]:border-b-2 data-[state=active]:border-teal rounded-none"
                    >
                      Credentials
                    </TabsTrigger>
                    <TabsTrigger
                      value="test-sending"
                      className="data-[state=active]:border-b-2 data-[state=active]:border-teal rounded-none"
                    >
                      Test Sending
                    </TabsTrigger>
                    <TabsTrigger
                      value="webhook"
                      className="data-[state=active]:border-b-2 data-[state=active]:border-teal rounded-none"
                    >
                      Webhook Simulator
                    </TabsTrigger>
                  </TabsList>

                  <TabsContent value="credentials" className="p-6 space-y-6">
                    <div>
                      <h2 className="text-xl font-semibold mb-2">
                        API Credentials
                      </h2>
                      <p className="text-sm text-muted-foreground">
                        Use these credentials to authenticate requests from your
                        application.
                      </p>
                    </div>

                    <div className="space-y-4">
                      {/* API Access Token */}
                      <div>
                        <div className="flex items-center justify-between mb-2">
                          <label className="text-sm font-medium">
                            API Access Token:
                          </label>
                          <Button
                            variant="ghost"
                            size="sm"
                            className="glass-card hover-glow-teal h-8"
                            onClick={handleCopyToken}
                          >
                            <Copy className="h-3 w-3 mr-1" />
                            Copy
                          </Button>
                        </div>
                        <div className="glass-card p-3 rounded-lg flex items-center justify-between">
                          <code className="text-sm font-mono">
                            {showApiToken ? apiToken : "••••••••••••"}
                          </code>
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-8 w-8"
                            onClick={() => setShowApiToken(!showApiToken)}
                          >
                            {showApiToken ? (
                              <EyeOff className="h-4 w-4" />
                            ) : (
                              <Eye className="h-4 w-4" />
                            )}
                          </Button>
                        </div>
                      </div>

                      {/* API Documentation Link */}
                      <div className="flex items-center justify-end">
                        <Button
                          variant="link"
                          className="text-teal hover:text-teal/80 p-0"
                        >
                          <ExternalLink className="h-4 w-4 mr-1" />
                          API Documentation
                        </Button>
                      </div>
                    </div>

                    {/* Connection Details */}
                    <div className="space-y-4 pt-4 border-t border-border/50">
                      <h3 className="font-semibold">Connection Details</h3>

                      {/* WhatsApp JID */}
                      {session.jid && (
                        <div className="flex items-start gap-3">
                          <div className="p-2 rounded-lg glass-card bg-primary/10">
                            <User className="h-5 w-5 text-primary" />
                          </div>
                          <div className="flex-1">
                            <p className="text-sm text-muted-foreground">
                              WhatsApp JID
                            </p>
                            <p className="font-medium break-all">
                              {session.jid}
                            </p>
                          </div>
                        </div>
                      )}

                      {/* Session ID */}
                      <div className="flex items-start gap-3">
                        <div className="p-2 rounded-lg glass-card bg-teal/10 glow-teal-sm">
                          <Hash className="h-5 w-5 text-teal" />
                        </div>
                        <div className="flex-1">
                          <p className="text-sm text-muted-foreground">
                            Session ID
                          </p>
                          <p className="font-medium break-all">{session.id}</p>
                        </div>
                      </div>

                      {/* Created At */}
                      <div className="flex items-start gap-3">
                        <div className="p-2 rounded-lg glass-card bg-primary/10">
                          <Calendar className="h-5 w-5 text-primary" />
                        </div>
                        <div className="flex-1">
                          <p className="text-sm text-muted-foreground">
                            Created
                          </p>
                          <p className="font-medium">
                            {new Date(session.created_at).toLocaleString()}
                          </p>
                        </div>
                      </div>
                    </div>
                  </TabsContent>

                  <TabsContent value="test-sending" className="p-6 space-y-6">
                    <div>
                      <h2 className="text-xl font-semibold mb-2">
                        Test Sending Capability
                      </h2>
                      <p className="text-sm text-muted-foreground">
                        Send a real WhatsApp message to verify your session is
                        working.
                      </p>
                    </div>

                    <div className="space-y-4">
                      {/* Destination Number */}
                      <div>
                        <label className="text-sm font-medium mb-2 block">
                          Destination Number
                        </label>
                        <Input
                          type="tel"
                          placeholder="e.g. +1234567890 (with country code)"
                          className="glass-card"
                        />
                        <p className="text-xs text-muted-foreground mt-1">
                          Enter the full phone number with country code (e.g.,
                          +1 for US)
                        </p>
                      </div>

                      {/* Message */}
                      <div>
                        <label className="text-sm font-medium mb-2 block">
                          Message
                        </label>
                        <textarea
                          placeholder="Hello!"
                          rows={6}
                          className="w-full glass-card rounded-lg px-3 py-2 text-sm bg-background border border-border focus:outline-none focus:ring-2 focus:ring-teal resize-none"
                          defaultValue="Hello!"
                        />
                      </div>

                      {/* Send Button */}
                      <Button className="glass-card hover-glow-teal">
                        <svg
                          className="mr-2 h-4 w-4"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth={2}
                            d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8"
                          />
                        </svg>
                        Send Test Message
                      </Button>
                    </div>

                    {/* OR SEND VIA CLI */}
                    <div className="pt-6 border-t border-border/50">
                      <h3 className="text-sm font-medium text-muted-foreground mb-4">
                        OR SEND VIA CLI
                      </h3>

                      <div className="relative glass-card rounded-lg p-4 bg-muted/5">
                        <Button
                          variant="ghost"
                          size="icon"
                          className="absolute top-2 right-2 glass-card hover-glow-teal h-8 w-8"
                          onClick={() => {
                            const code = `curl -X POST "https://wasenderapi.com/api/send-message" \\
  -H "Authorization: Bearer ${apiToken}" \\
  -H "Content-Type: application/json" \\
  -d '{"to": "+1234567890", "text": "Hello!"}'`;
                            navigator.clipboard.writeText(code);
                            toast.success("Command copied to clipboard");
                          }}
                        >
                          <Copy className="h-4 w-4" />
                        </Button>

                        <pre className="text-xs font-mono overflow-x-auto">
                          <code className="text-muted-foreground">
                            <span className="text-teal">curl</span>{" "}
                            <span className="text-amber">-X</span> POST{" "}
                            <span className="text-emerald">
                              "https://wasenderapi.com/api/send-message"
                            </span>{" "}
                            \{"\n"}
                            {"  "}
                            <span className="text-amber">-H</span>{" "}
                            <span className="text-emerald">
                              "Authorization: Bearer {apiToken}"
                            </span>{" "}
                            \{"\n"}
                            {"  "}
                            <span className="text-amber">-H</span>{" "}
                            <span className="text-emerald">
                              "Content-Type: application/json"
                            </span>{" "}
                            \{"\n"}
                            {"  "}
                            <span className="text-amber">-d</span>{" "}
                            <span className="text-emerald">
                              '{"{"}
                              "to": "+1234567890", "text": "Hello!"
                              {"}"}'
                            </span>
                          </code>
                        </pre>
                      </div>

                      <p className="text-xs text-muted-foreground mt-2">
                        This command uses your actual API key and the values
                        entered above.
                      </p>
                    </div>
                  </TabsContent>

                  <TabsContent value="webhook" className="p-6">
                    <div className="flex flex-col items-center justify-center py-16 space-y-6">
                      <div className="relative">
                        <div className="absolute inset-0 blur-xl opacity-20">
                          <Webhook className="h-24 w-24 text-teal" />
                        </div>
                        <Webhook className="relative h-24 w-24 text-muted-foreground" />
                      </div>

                      <div className="text-center space-y-2 max-w-md">
                        <h3 className="text-xl font-semibold">
                          Webhook Not Configured
                        </h3>
                        <p className="text-sm text-muted-foreground">
                          Please configure a webhook URL in your session
                          settings to use the simulator.
                        </p>
                      </div>

                      <Button className="glass-card hover-glow-teal">
                        <svg
                          className="mr-2 h-4 w-4"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth={2}
                            d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                          />
                        </svg>
                        Configure Webhook
                      </Button>
                    </div>
                  </TabsContent>
                </Tabs>
              </div>
            )}
          </div>
        </div>

        {/* Recent Activity Section */}
        {session.status === "connected" && (
          <div className="mt-6 space-y-6">
            {/* Tabs for Recent Activity and Session Logs */}
            <div className="glass-card-enhanced">
              <Tabs defaultValue="recent-activity" className="w-full">
                <TabsList className="w-full justify-start border-b border-border/50 rounded-none bg-transparent p-0 px-6">
                  <TabsTrigger
                    value="recent-activity"
                    className="data-[state=active]:border-b-2 data-[state=active]:border-teal rounded-none"
                  >
                    <svg
                      className="mr-2 h-4 w-4"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
                      />
                    </svg>
                    Recent Activity
                  </TabsTrigger>
                  <TabsTrigger
                    value="session-logs"
                    className="data-[state=active]:border-b-2 data-[state=active]:border-teal rounded-none"
                  >
                    <svg
                      className="mr-2 h-4 w-4"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                      />
                    </svg>
                    Session Logs
                  </TabsTrigger>
                </TabsList>

                <TabsContent value="recent-activity" className="p-6 space-y-6">
                  {/* Outgoing Message Activity */}
                  <div>
                    <div className="flex items-center justify-between mb-4">
                      <div>
                        <h3 className="text-lg font-semibold">
                          Outgoing Message Activity
                        </h3>
                        <p className="text-sm text-muted-foreground">
                          Track the status of messages sent via our API in
                          real-time
                        </p>
                      </div>
                      <Button
                        variant="outline"
                        size="sm"
                        className="glass-card hover-glow-teal"
                      >
                        <RefreshCw className="mr-2 h-3 w-3" />
                        Go Live
                      </Button>
                    </div>

                    {/* Info Banner */}
                    <div className="glass-card p-4 rounded-lg bg-blue-500/5 border-blue-500/20 mb-4">
                      <div className="flex gap-3">
                        <svg
                          className="h-5 w-5 text-blue-400 shrink-0 mt-0.5"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth={2}
                            d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                          />
                        </svg>
                        <div className="flex-1 space-y-2">
                          <p className="text-sm">
                            This activity monitor only shows{" "}
                            <span className="font-medium">
                              outgoing messages
                            </span>{" "}
                            sent through our API.
                          </p>
                          <p className="text-sm text-muted-foreground">
                            Messages you send manually from your Phone app or
                            WhatsApp Web will not appear here. This is a
                            developer dashboard for integration logs, not a full
                            chat history sync.
                          </p>
                          <p className="text-sm">
                            To receive and process{" "}
                            <span className="font-medium">
                              incoming messages
                            </span>{" "}
                            from your customers in your own app or n8n, you must{" "}
                            <button
                              onClick={() =>
                                navigate({
                                  to: "/sessions/$sessionId/webhooks",
                                  params: { sessionId: session.id },
                                })
                              }
                              className="text-teal hover:underline font-medium"
                            >
                              configure webhooks
                            </button>
                          </p>
                        </div>
                      </div>
                    </div>

                    {/* Message Activity Table Header */}
                    <div className="flex items-center justify-between mb-4">
                      <div className="text-sm text-muted-foreground">
                        <span className="font-medium">0</span> total messages
                      </div>
                      <div className="flex items-center gap-2">
                        <select className="glass-card px-3 py-1.5 rounded-lg text-sm bg-background border border-border">
                          <option>All statuses</option>
                          <option>Sent</option>
                          <option>Delivered</option>
                          <option>Read</option>
                          <option>Failed</option>
                        </select>
                        <select className="glass-card px-3 py-1.5 rounded-lg text-sm bg-background border border-border">
                          <option>25 per page</option>
                          <option>50 per page</option>
                          <option>100 per page</option>
                        </select>
                      </div>
                    </div>

                    {/* Empty State */}
                    <div className="glass-card rounded-lg p-12 text-center">
                      <svg
                        className="h-16 w-16 text-muted-foreground mx-auto mb-4"
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
                      <h4 className="text-lg font-semibold mb-2">
                        No messages yet
                      </h4>
                      <p className="text-sm text-muted-foreground">
                        Send messages through the API to see them appear here.
                      </p>
                    </div>
                  </div>
                </TabsContent>

                <TabsContent value="session-logs" className="p-6 space-y-6">
                  <div>
                    <div className="flex items-center justify-between mb-4">
                      <div>
                        <h3 className="text-lg font-semibold">Session Logs</h3>
                        <p className="text-sm text-muted-foreground">
                          Recent activity and system events for this session
                        </p>
                      </div>
                      <Button
                        variant="outline"
                        size="sm"
                        className="glass-card hover-glow-destructive"
                      >
                        <svg
                          className="mr-2 h-3 w-3"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth={2}
                            d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                          />
                        </svg>
                        Clear Logs
                      </Button>
                    </div>

                    <div className="text-sm text-muted-foreground mb-4">
                      <span className="font-medium">3</span> total logs
                    </div>

                    {/* Log Entries */}
                    <div className="space-y-3">
                      {/* Connected Status */}
                      <div className="glass-card p-4 rounded-lg hover-lift transition-all">
                        <div className="flex items-start gap-4">
                          <div className="shrink-0 w-10 h-10 rounded-full bg-emerald/10 flex items-center justify-center">
                            <CheckCircle2 className="h-5 w-5 text-emerald" />
                          </div>
                          <div className="flex-1 min-w-0">
                            <div className="flex items-start justify-between gap-4 mb-1">
                              <h4 className="font-medium">Status Change</h4>
                              <span className="shrink-0 text-xs px-2 py-1 rounded-full bg-emerald/10 text-emerald font-medium">
                                Connected
                              </span>
                            </div>
                            <p className="text-sm text-muted-foreground mb-2">
                              The WhatsApp session is connected and ready to
                              use.
                            </p>
                            <p className="text-xs text-muted-foreground">
                              2/3/2026, 5:33:14 PM
                            </p>
                          </div>
                        </div>
                      </div>

                      {/* Connecting Status */}
                      <div className="glass-card p-4 rounded-lg hover-lift transition-all">
                        <div className="flex items-start gap-4">
                          <div className="shrink-0 w-10 h-10 rounded-full bg-purple-500/10 flex items-center justify-center">
                            <Circle className="h-5 w-5 text-purple-400" />
                          </div>
                          <div className="flex-1 min-w-0">
                            <div className="flex items-start justify-between gap-4 mb-1">
                              <h4 className="font-medium">Status Change</h4>
                              <span className="shrink-0 text-xs px-2 py-1 rounded-full bg-purple-500/10 text-purple-400 font-medium">
                                Connecting
                              </span>
                            </div>
                            <p className="text-sm text-muted-foreground mb-2">
                              The WhatsApp session is in the process of
                              connecting.
                            </p>
                            <p className="text-xs text-muted-foreground">
                              2/3/2026, 5:33:10 PM
                            </p>
                          </div>
                        </div>
                      </div>

                      {/* Needs QR Scan Status */}
                      <div className="glass-card p-4 rounded-lg hover-lift transition-all">
                        <div className="flex items-start gap-4">
                          <div className="shrink-0 w-10 h-10 rounded-full bg-blue-500/10 flex items-center justify-center">
                            <QrCode className="h-5 w-5 text-blue-400" />
                          </div>
                          <div className="flex-1 min-w-0">
                            <div className="flex items-start justify-between gap-4 mb-1">
                              <h4 className="font-medium">Status Change</h4>
                              <span className="shrink-0 text-xs px-2 py-1 rounded-full bg-blue-500/10 text-blue-400 font-medium">
                                Needs QR Scan
                              </span>
                            </div>
                            <p className="text-sm text-muted-foreground mb-2">
                              The WhatsApp session needs QR code scanning to
                              connect.
                            </p>
                            <p className="text-xs text-muted-foreground">
                              2/3/2026, 5:32:21 PM
                            </p>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                </TabsContent>
              </Tabs>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
