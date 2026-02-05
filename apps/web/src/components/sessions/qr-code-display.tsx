import { CheckCircle2, Loader2, QrCode, RefreshCw } from "lucide-react";
import { useEffect, useState } from "react";
import { toast } from "sonner";

import { useSessionStore } from "@/stores/session-store";

import { Button } from "../ui/button";
import { QRCode } from "../kibo-ui/qr-code";
import { createQRWebSocket, type QRWebSocketEvent } from "@/lib/websocket";

// ============================================================================
// Types
// ============================================================================

interface QRCodeDisplayProps {
  sessionId: string;
  onAuthenticated?: (jid?: string) => void;
  onError?: (message: string) => void;
}

type QRStatus = "loading" | "ready" | "authenticated" | "expired";

// ============================================================================
// Component
// ============================================================================

export function QRCodeDisplay({
  sessionId,
  onAuthenticated,
  onError,
}: QRCodeDisplayProps) {
  const [status, setStatus] = useState<QRStatus>("loading");
  const [qrData, setQrData] = useState<string>("");
  const { updateSession } = useSessionStore();

  useEffect(() => {
    // Initialize WebSocket connection
    const ws = createQRWebSocket(sessionId);
    ws.connect();

    // Subscribe to events
    const unsubscribe = ws.subscribe((event: QRWebSocketEvent) => {
      switch (event.type) {
        case "qr":
          setQrData(event.data);
          setStatus("ready");
          break;

        case "authenticated":
          setStatus("authenticated");
          // Update session in store
          updateSession(sessionId, {
            status: "connected",
            jid: event.data,
            updated_at: new Date().toISOString(),
          });
          toast.success("Session authenticated successfully!");
          onAuthenticated?.(event.data);
          break;

        case "error":
          toast.error(`Authentication error: ${event.message}`);
          onError?.(event.message);
          break;

        case "timeout":
          setStatus("expired");
          break;
      }
    });

    // Cleanup
    return () => {
      unsubscribe();
      ws.disconnect();
    };
  }, [sessionId, updateSession, onAuthenticated, onError]);

  const handleRefresh = () => {
    // Re-mount component to trigger new connection/QR generation
    setStatus("loading");
    // In a real app we might want to trigger a backend retry via API or WS
    // For now, re-connecting WS should trigger a new QR from backend logic
    const ws = createQRWebSocket(sessionId);
    ws.connect();
  };

  const renderContent = () => {
    switch (status) {
      case "loading":
        return (
          <div className="flex flex-col items-center justify-center p-12 space-y-4">
            <Loader2 className="h-12 w-12 animate-spin text-teal" />
            <p className="text-sm text-muted-foreground">
              Waiting for QR code...
            </p>
          </div>
        );

      case "ready":
        return (
          <div className="flex flex-col items-center justify-center space-y-6">
            <div className="relative">
              {/* Render actual QR code using kibo-ui component */}
              <div className="w-80 h-80 border-4 border-teal rounded-2xl glass-card p-4 glow-teal bg-white">
                <QRCode
                  data={qrData}
                  className="w-full h-full"
                  robustness="M"
                />
              </div>
            </div>
            <div className="text-center space-y-2">
              <p className="text-sm font-medium">Scan with WhatsApp</p>
              <p className="text-xs text-muted-foreground max-w-md">
                Open WhatsApp on your phone → Settings → Linked Devices → Link a
                Device
              </p>
            </div>
          </div>
        );

      case "authenticated":
        return (
          <div className="flex flex-col items-center justify-center p-12 space-y-4">
            <CheckCircle2 className="h-16 w-16 text-emerald glow-emerald" />
            <p className="text-lg font-semibold text-emerald">
              Successfully authenticated!
            </p>
            <p className="text-sm text-muted-foreground">
              Your WhatsApp session is now connected
            </p>
          </div>
        );

      case "expired":
        return (
          <div className="flex flex-col items-center justify-center p-12 space-y-4">
            <QrCode className="h-16 w-16 text-muted-foreground" />
            <div className="text-center space-y-2">
              <p className="text-lg font-semibold text-muted-foreground">
                QR Code Expired
              </p>
              <p className="text-sm text-muted-foreground">
                Please refresh to generate a new code
              </p>
            </div>
            <Button
              onClick={handleRefresh}
              className="glass-card hover-glow-teal"
            >
              <RefreshCw className="mr-2 h-4 w-4" />
              Refresh Code
            </Button>
          </div>
        );
    }
  };

  return (
    <div className="w-full">
      <div className="glass-card-enhanced p-8 rounded-2xl">
        {renderContent()}
      </div>
    </div>
  );
}
