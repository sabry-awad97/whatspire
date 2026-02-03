import { CheckCircle2, Loader2, QrCode, RefreshCw } from "lucide-react";
import { useEffect, useState } from "react";
import { toast } from "sonner";

import { useSessionStore } from "@/stores/session-store";

import { Button } from "../ui/button";

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
  const [countdown, setCountdown] = useState(60);
  const { updateSession } = useSessionStore();

  // Mock QR code - in real implementation this would come from WebSocket
  const mockQRCode =
    "data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHZpZXdCb3g9IjAgMCAyOTAgMjkwIj48cmVjdCB3aWR0aD0iMjkwIiBoZWlnaHQ9IjI5MCIgZmlsbD0iI2ZmZmZmZiIvPjxwYXRoIGQ9Ik0xMCAxMGgzMHYzMEgxMHpNNTAgMTBoMzB2MzBINTB6TTkwIDEwaDMwdjMwSDkwek0xMzAgMTBoMzB2MzBIMTMwek0xNzAgMTBoMzB2MzBIMTcwek0yMTAgMTBoMzB2MzBIMjEwek0yNTAgMTBoMzB2MzBIMjUwek0xMCA1MGgzMHYzMEgxMHpNMjUwIDUwaDMwdjMwSDI1MHpNMTAgOTBoMzB2MzBIMTB6TTkwIDkwaDMwdjMwSDkwek0xNzAgOTBoMzB2MzBIMTcwek0yNTAgOTBoMzB2MzBIMjUwek0xMCAxMzBoMzB2MzBIMTB6TTkwIDEzMGgzMHYzMEg5MHpNMTcwIDEzMGgzMHYzMEgxNzB6TTI1MCAxMzBoMzB2MzBIMjUwek0xMCAxNzBoMzB2MzBIMTB6TTkwIDE3MGgzMHYzMEg5MHpNMTcwIDE3MGgzMHYzMEgxNzB6TTI1MCAxNzBoMzB2MzBIMjUwek0xMCAyMTBoMzB2MzBIMTB6TTUwIDIxMGgzMHYzMEg1MHpNOTAgMjEwaDMwdjMwSDkwek0xMzAgMjEwaDMwdjMwSDEzMHpNMTcwIDIxMGgzMHYzMEgxNzB6TTIxMCAyMTBoMzB2MzBIMjEwek0yNTAgMjEwaDMwdjMwSDI1MHpNMTAgMjUwaDMwdjMwSDEwek0yNTAgMjUwaDMwdjMwSDI1MHoiIGZpbGw9IiMwMDAwMDAiLz48L3N2Zz4=";

  useEffect(() => {
    // Simulate loading
    const loadTimer = setTimeout(() => {
      setStatus("ready");
    }, 1000);

    return () => clearTimeout(loadTimer);
  }, []);

  useEffect(() => {
    if (status !== "ready") return;

    // Countdown timer
    const timer = setInterval(() => {
      setCountdown((prev) => {
        if (prev <= 1) {
          setStatus("expired");
          toast.error("QR code expired");
          return 0;
        }
        return prev - 1;
      });
    }, 1000);

    return () => clearInterval(timer);
  }, [status]);

  const handleRefresh = () => {
    setStatus("loading");
    setCountdown(60);
    setTimeout(() => {
      setStatus("ready");
    }, 1000);
  };

  const handleSimulateConnect = () => {
    setStatus("authenticated");
    const mockJid = `${Math.floor(Math.random() * 9000000000) + 1000000000}@s.whatsapp.net`;

    // Update session in store
    updateSession(sessionId, {
      status: "connected",
      jid: mockJid,
      updated_at: new Date().toISOString(),
    });

    toast.success("Session authenticated successfully!");
    onAuthenticated?.(mockJid);
  };

  const renderContent = () => {
    switch (status) {
      case "loading":
        return (
          <div className="flex flex-col items-center justify-center p-12 space-y-4">
            <Loader2 className="h-12 w-12 animate-spin text-teal" />
            <p className="text-sm text-muted-foreground">
              Generating QR code...
            </p>
          </div>
        );

      case "ready":
        return (
          <div className="flex flex-col items-center justify-center space-y-6">
            <div className="relative">
              <img
                src={mockQRCode}
                alt="QR Code"
                className="w-80 h-80 border-4 border-teal rounded-2xl glass-card p-4 glow-teal"
              />
            </div>
            <div className="text-center space-y-2">
              <p className="text-sm font-medium">
                Expires in {countdown} seconds
              </p>
              <p className="text-xs text-muted-foreground max-w-md">
                Open WhatsApp on your phone → Settings → Linked Devices → Link a
                Device
              </p>
            </div>
            <Button
              onClick={handleRefresh}
              variant="outline"
              className="glass-card hover-glow-teal"
            >
              <RefreshCw className="mr-2 h-4 w-4" />
              Refresh QR Code
            </Button>

            {/* Mock connect button for testing */}
            <Button
              onClick={handleSimulateConnect}
              className="glass-card hover-glow-emerald"
            >
              <CheckCircle2 className="mr-2 h-4 w-4" />
              Simulate Connection (Dev Only)
            </Button>
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
                Click refresh to generate a new QR code
              </p>
            </div>
            <Button
              onClick={handleRefresh}
              className="glass-card hover-glow-teal"
            >
              <RefreshCw className="mr-2 h-4 w-4" />
              Refresh QR Code
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
