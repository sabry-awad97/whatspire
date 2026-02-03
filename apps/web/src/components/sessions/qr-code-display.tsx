import { CheckCircle2, Loader2, QrCode, XCircle } from "lucide-react";
import { useEffect, useState } from "react";
import { toast } from "sonner";

import { createQRWebSocket, type QRWebSocketEvent } from "@/lib/websocket";
import { useSessionStore } from "@/stores/session-store";

import { Card } from "../ui/card";

// ============================================================================
// Types
// ============================================================================

interface QRCodeDisplayProps {
  sessionId: string;
  onAuthenticated?: (jid: string) => void;
  onError?: (message: string) => void;
}

type QRStatus =
  | "connecting"
  | "waiting"
  | "authenticated"
  | "error"
  | "timeout";

// ============================================================================
// Component
// ============================================================================

export function QRCodeDisplay({
  sessionId,
  onAuthenticated,
  onError,
}: QRCodeDisplayProps) {
  const [status, setStatus] = useState<QRStatus>("connecting");
  const [errorMessage, setErrorMessage] = useState<string>("");
  const { qrCodes, setQRCode, clearQRCode, updateSession } = useSessionStore();

  const qrData = qrCodes.get(sessionId);

  useEffect(() => {
    const ws = createQRWebSocket(sessionId);

    // Handle WebSocket events
    const unsubscribe = ws.subscribe((event: QRWebSocketEvent) => {
      switch (event.type) {
        case "qr":
          setStatus("waiting");
          setQRCode(sessionId, event.data);
          break;

        case "authenticated":
          setStatus("authenticated");
          clearQRCode(sessionId);
          updateSession(sessionId, {
            status: "connected",
            jid: event.data,
          });
          toast.success("Session authenticated successfully!");
          onAuthenticated?.(event.data);
          break;

        case "error":
          setStatus("error");
          setErrorMessage(event.message);
          toast.error(event.message);
          onError?.(event.message);
          break;

        case "timeout":
          setStatus("timeout");
          setErrorMessage("QR code expired. Please try again.");
          toast.error("QR code expired");
          onError?.("QR code expired");
          break;
      }
    });

    // Handle connection events
    const unsubscribeOpen = ws.onOpen(() => {
      console.log("QR WebSocket connected");
    });

    const unsubscribeClose = ws.onClose(() => {
      console.log("QR WebSocket closed");
    });

    const unsubscribeError = ws.onError((error) => {
      console.error("QR WebSocket error:", error);
      setStatus("error");
      setErrorMessage("Connection error. Please try again.");
    });

    // Connect
    ws.connect();

    // Cleanup
    return () => {
      unsubscribe();
      unsubscribeOpen();
      unsubscribeClose();
      unsubscribeError();
      ws.disconnect();
      clearQRCode(sessionId);
    };
  }, [
    sessionId,
    setQRCode,
    clearQRCode,
    updateSession,
    onAuthenticated,
    onError,
  ]);

  const renderContent = () => {
    switch (status) {
      case "connecting":
        return (
          <div className="flex flex-col items-center justify-center p-8 space-y-4">
            <Loader2 className="h-12 w-12 animate-spin text-primary" />
            <p className="text-sm text-muted-foreground">
              Connecting to server...
            </p>
          </div>
        );

      case "waiting":
        return (
          <div className="flex flex-col items-center justify-center p-8 space-y-4">
            {qrData ? (
              <>
                <img
                  src={`data:image/png;base64,${qrData.qrCode}`}
                  alt="QR Code"
                  className="w-64 h-64 border-4 border-primary rounded-lg"
                />
                <div className="text-center space-y-2">
                  <p className="text-sm font-medium">Scan with WhatsApp</p>
                  <p className="text-xs text-muted-foreground">
                    Open WhatsApp on your phone → Settings → Linked Devices →
                    Link a Device
                  </p>
                </div>
              </>
            ) : (
              <>
                <QrCode className="h-12 w-12 text-muted-foreground" />
                <p className="text-sm text-muted-foreground">
                  Waiting for QR code...
                </p>
              </>
            )}
          </div>
        );

      case "authenticated":
        return (
          <div className="flex flex-col items-center justify-center p-8 space-y-4">
            <CheckCircle2 className="h-12 w-12 text-green-500" />
            <p className="text-sm font-medium text-green-500">
              Successfully authenticated!
            </p>
          </div>
        );

      case "error":
      case "timeout":
        return (
          <div className="flex flex-col items-center justify-center p-8 space-y-4">
            <XCircle className="h-12 w-12 text-destructive" />
            <div className="text-center space-y-2">
              <p className="text-sm font-medium text-destructive">
                {status === "timeout" ? "QR Code Expired" : "Error"}
              </p>
              <p className="text-xs text-muted-foreground">{errorMessage}</p>
            </div>
          </div>
        );
    }
  };

  return (
    <Card className="w-full max-w-md mx-auto">
      <div className="p-6">
        <h3 className="text-lg font-semibold mb-4 text-center">
          WhatsApp Authentication
        </h3>
        {renderContent()}
      </div>
    </Card>
  );
}
